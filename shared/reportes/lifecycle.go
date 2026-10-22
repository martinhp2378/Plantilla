package shared

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const reportRetention = 24 * time.Hour

type retainedReport struct {
	ID, Claim uuid.UUID
	TenantID  string
	Path      string
}

type retentionRepository interface {
	prepare(context.Context, string, string) (uuid.UUID, error)
	activate(context.Context, string, uuid.UUID) error
	remove(context.Context, string, uuid.UUID) error
	tenants(context.Context) ([]string, error)
	claim(context.Context, string, int) ([]retainedReport, error)
	complete(context.Context, retainedReport) error
	retry(context.Context, retainedReport, error) error
}

type privateObjects interface {
	exists(context.Context, string) (bool, error)
	delete(context.Context, string) error
}

type managedObjects struct{ manager *archivos.ManejadorArchivos }

func (o managedObjects) exists(ctx context.Context, path string) (bool, error) {
	return o.manager.ArchivoPrivado(path).Existe(ctx)
}
func (o managedObjects) delete(ctx context.Context, path string) error {
	return o.manager.ArchivoPrivado(path).Eliminar(ctx)
}

// Lifecycle owns durable report registration and restart-safe cleanup.
type Lifecycle struct {
	repo     retentionRepository
	objects  privateObjects
	interval time.Duration
	cancel   context.CancelFunc
	done     chan struct{}
	mu       sync.Mutex
}

func NewLifecycle(pool *pgxpool.Pool, manager *archivos.ManejadorArchivos, interval time.Duration) *Lifecycle {
	return newLifecycle(&pgRetentionRepo{pool: pool}, managedObjects{manager}, interval)
}

func newLifecycle(repo retentionRepository, objects privateObjects, interval time.Duration) *Lifecycle {
	return &Lifecycle{repo: repo, objects: objects, interval: interval}
}

type lifecycleContextKey struct{}

func (l *Lifecycle) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, lifecycleContextKey{}, l)
}

func (l *Lifecycle) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(l.Context(r.Context())))
	})
}

func lifecycleFromContext(ctx context.Context) (*Lifecycle, error) {
	lifecycle, ok := ctx.Value(lifecycleContextKey{}).(*Lifecycle)
	if !ok || lifecycle == nil {
		return nil, errors.New("report retention unavailable")
	}
	return lifecycle, nil
}

func (l *Lifecycle) Start(parent context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	l.cancel, l.done = cancel, make(chan struct{})
	go func() {
		defer close(l.done)
		_ = l.Sweep(ctx)
		ticker := time.NewTicker(l.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = l.Sweep(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (l *Lifecycle) Stop(ctx context.Context) error {
	l.mu.Lock()
	cancel, done := l.cancel, l.done
	l.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *Lifecycle) Sweep(ctx context.Context) error {
	tenants, err := l.repo.tenants(ctx)
	if err != nil {
		return err
	}
	var result error
	for _, tenantID := range tenants {
		items, err := l.repo.claim(ctx, tenantID, 50)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		for _, item := range items {
			if err := l.cleanup(ctx, item); err != nil {
				result = errors.Join(result, err)
			}
		}
	}
	return result
}

func (l *Lifecycle) cleanup(ctx context.Context, item retainedReport) error {
	var cleanupErr error
	if ctx.Err() != nil {
		cleanupErr = ctx.Err()
	} else if exists, err := l.objects.exists(ctx, item.Path); err != nil {
		cleanupErr = err
	} else if exists {
		cleanupErr = l.objects.delete(ctx, item.Path)
	}
	finalizeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if cleanupErr != nil {
		if err := l.repo.retry(finalizeCtx, item, cleanupErr); err != nil {
			return errors.Join(cleanupErr, err)
		}
		return cleanupErr
	}
	return l.repo.complete(finalizeCtx, item)
}

type pgRetentionRepo struct{ pool *pgxpool.Pool }

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (r *pgRetentionRepo) resolveTenantID(ctx context.Context, tenantID string) string {
	if uuidRegex.MatchString(tenantID) {
		return tenantID
	}
	var fallbackID string
	err := r.pool.QueryRow(ctx, `SELECT tenant_id::text FROM institucion.cat_tenants LIMIT 1`).Scan(&fallbackID)
	if err == nil && fallbackID != "" {
		return fallbackID
	}
	return tenantID
}

func (r *pgRetentionRepo) tenantTx(ctx context.Context, tenantID string, fn func(pgx.Tx) error) error {
	tenantID = r.resolveTenantID(ctx, tenantID)
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		return err
	}
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *pgRetentionRepo) prepare(ctx context.Context, tenantID, path string) (uuid.UUID, error) {
	tenantID = r.resolveTenantID(ctx, tenantID)
	id := uuid.Must(uuid.NewV4())
	err := r.tenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		// Pending rows reserve five minutes solely for orphan compensation before activation.
		_, err := tx.Exec(ctx, `INSERT INTO sistema.reportes_generados
			(report_id, tenant_id, storage_path, state, expires_at)
			VALUES ($1, $2::uuid, $3, 'pending', NOW() + INTERVAL '5 minutes')`, id, tenantID, path)
		return err
	})
	return id, err
}

func (r *pgRetentionRepo) activate(ctx context.Context, tenantID string, id uuid.UUID) error {
	tenantID = r.resolveTenantID(ctx, tenantID)
	return r.tenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE sistema.reportes_generados SET state='active', expires_at=created_at + INTERVAL '24 hours' WHERE report_id=$1`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return errors.New("report activation affected no rows")
		}
		return nil
	})
}

func (r *pgRetentionRepo) remove(ctx context.Context, tenantID string, id uuid.UUID) error {
	tenantID = r.resolveTenantID(ctx, tenantID)
	return r.tenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM sistema.reportes_generados WHERE report_id=$1`, id)
		return err
	})
}

func (r *pgRetentionRepo) tenants(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT tenant_id::text FROM institucion.cat_tenants`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tenants []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		tenants = append(tenants, id)
	}
	return tenants, rows.Err()
}

func (r *pgRetentionRepo) claim(ctx context.Context, tenantID string, limit int) (items []retainedReport, err error) {
	claim := uuid.Must(uuid.NewV4())
	err = r.tenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `WITH due AS (
			SELECT report_id FROM sistema.reportes_generados
			WHERE (state IN ('pending','active') AND expires_at <= NOW())
			   OR (state='retry' AND next_retry_at <= NOW())
			   OR (state='claimed' AND claimed_at <= NOW() - INTERVAL '15 minutes')
			ORDER BY expires_at FOR UPDATE SKIP LOCKED LIMIT $1)
			UPDATE sistema.reportes_generados r SET state='claimed', claim_token=$2, claimed_at=NOW()
			FROM due WHERE r.report_id=due.report_id
			RETURNING r.report_id, r.tenant_id::text, r.storage_path`, limit, claim)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			item := retainedReport{Claim: claim}
			if err := rows.Scan(&item.ID, &item.TenantID, &item.Path); err != nil {
				return err
			}
			items = append(items, item)
		}
		return rows.Err()
	})
	return items, err
}

func (r *pgRetentionRepo) complete(ctx context.Context, item retainedReport) error {
	return r.tenantTx(ctx, item.TenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `DELETE FROM sistema.reportes_generados WHERE report_id=$1 AND claim_token=$2`, item.ID, item.Claim)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return errors.New("report completion affected no rows")
		}
		return nil
	})
}

func (r *pgRetentionRepo) retry(ctx context.Context, item retainedReport, cause error) error {
	return r.tenantTx(ctx, item.TenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE sistema.reportes_generados SET
			retry_count=LEAST(retry_count+1,5), state=CASE WHEN retry_count+1>=5 THEN 'failed' ELSE 'retry' END,
			next_retry_at=NOW() + make_interval(secs => (60 * power(2,retry_count))::int),
			last_error=LEFT($3,500), claim_token=NULL, claimed_at=NULL
			WHERE report_id=$1 AND claim_token=$2`, item.ID, item.Claim, cause.Error())
		return err
	})
}

func (l *Lifecycle) register(ctx context.Context, tenantID, path string) (uuid.UUID, error) {
	return l.repo.prepare(ctx, tenantID, path)
}
