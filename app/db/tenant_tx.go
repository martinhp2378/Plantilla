package bd

import (
	"context"
	"errors"
	"strings"

	"PlantillaInstitucional/shared/tenantcontext"
	"github.com/jackc/pgx/v5"
)

// ErrTenantTransaction intentionally contains no tenant or database details.
var ErrTenantTransaction = errors.New("tenant transaction unavailable")

type txStarter interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

// TenantTxRunner owns one tenant-scoped transaction at a time.
type TenantTxRunner struct {
	pool txStarter
}

func NewTenantTxRunner(pool txStarter) *TenantTxRunner {
	return &TenantTxRunner{pool: pool}
}

func (r *TenantTxRunner) Read(ctx context.Context, callback func(pgx.Tx) error) error {
	return r.run(ctx, pgx.ReadOnly, callback)
}

func (r *TenantTxRunner) Write(ctx context.Context, callback func(pgx.Tx) error) error {
	return r.run(ctx, pgx.ReadWrite, callback)
}

func (r *TenantTxRunner) run(ctx context.Context, accessMode pgx.TxAccessMode, callback func(pgx.Tx) error) error {
	if r == nil || r.pool == nil || callback == nil || ctx.Err() != nil {
		return ErrTenantTransaction
	}

	tenant, err := tenantcontext.TenantFromContext(ctx)
	if err != nil || tenant.ID != strings.TrimSpace(tenant.ID) {
		return ErrTenantTransaction
	}
	searchPath, err := buildTenantSearchPath(tenant.Schema)
	if err != nil {
		return ErrTenantTransaction
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: accessMode})
	if err != nil {
		return ErrTenantTransaction
	}
	committed := false
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			panic(recovered)
		}
		if !committed {
			_ = tx.Rollback(context.WithoutCancel(ctx))
		}
	}()

	if _, err = tx.Exec(ctx, "SET LOCAL search_path TO "+searchPath); err != nil {
		return ErrTenantTransaction
	}
	if _, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenant.ID); err != nil {
		return ErrTenantTransaction
	}
	if err = callback(tx); err != nil || ctx.Err() != nil {
		return ErrTenantTransaction
	}
	if err = tx.Commit(ctx); err != nil {
		return ErrTenantTransaction
	}
	committed = true
	return nil
}
