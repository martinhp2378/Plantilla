package bd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"PlantillaInstitucional/shared/tenantcontext"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestTenantTxRunnerFailsClosedBeforeBeginningTransaction(t *testing.T) {
	t.Parallel()

	pool := &fakeTxStarter{}
	runner := NewTenantTxRunner(pool, pool)
	unsafeTenant := tenantcontext.WithTenant(context.Background(), tenantcontext.TenantContext{
		ID: "tenant-secret", Schema: "public", CanonicalHost: "institution.example",
	})
	mismatchedTenant := tenantcontext.WithTenant(context.Background(), tenantcontext.TenantContext{
		ID: " tenant-secret ", Schema: "institution_one", CanonicalHost: "institution.example",
	})

	for _, tt := range []struct {
		name string
		ctx  context.Context
	}{
		{name: "missing tenant", ctx: context.Background()},
		{name: "unsafe schema", ctx: unsafeTenant},
		{name: "mutated tenant ID", ctx: mismatchedTenant},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := runner.Write(tt.ctx, func(pgx.Tx) error {
				t.Fatal("callback must not run")
				return nil
			})
			if !errors.Is(err, ErrTenantTransaction) {
				t.Fatalf("Write() error = %v, want ErrTenantTransaction", err)
			}
			if pool.beginCalls != 0 {
				t.Fatalf("BeginTx calls = %d, want 0", pool.beginCalls)
			}
			if strings.Contains(err.Error(), "tenant-secret") || strings.Contains(err.Error(), "public") {
				t.Fatalf("error leaked tenant context: %q", err)
			}
		})
	}
}

func TestTenantTxRunnerReadSetsTransactionLocalContextBeforeCallback(t *testing.T) {
	t.Parallel()

	tx := &fakeTx{}
	pool := &fakeTxStarter{tx: tx}
	runner := NewTenantTxRunner(pool, pool)
	ctx := tenantContext("tenant-1", "institution_one")

	err := runner.Read(ctx, func(got pgx.Tx) error {
		if got != tx {
			t.Fatal("callback received an unexpected transaction")
		}
		tx.calls = append(tx.calls, "callback")
		return nil
	})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if pool.options.AccessMode != pgx.ReadOnly {
		t.Fatalf("read access mode = %v, want ReadOnly", pool.options.AccessMode)
	}
	want := []string{"begin", `SET LOCAL search_path TO "institution_one", institucion, sesion, sistema, pg_catalog`, "set_config", "set_config", "callback", "commit"}
	if diff := strings.Join(tx.calls, ","); diff != strings.Join(want, ",") {
		t.Fatalf("transaction order = %v, want %v", tx.calls, want)
	}
	if len(tx.execArgs) != 3 || len(tx.execArgs[1]) != 1 || tx.execArgs[1][0] != "tenant-1" {
		t.Fatalf("set_config arguments = %#v, want tenant ID", tx.execArgs)
	}
}

func TestTenantTxRunnerRollsBackOnCallbackFailureAndCancellation(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		callback func(context.CancelFunc) error
	}{
		{name: "callback failure", callback: func(context.CancelFunc) error { return errors.New("callback failed") }},
		{name: "cancellation", callback: func(cancel context.CancelFunc) error { cancel(); return nil }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tx := &fakeTx{}
			pool := &fakeTxStarter{tx: tx}
			runner := NewTenantTxRunner(pool, pool)
			ctx, cancel := context.WithCancel(tenantContext("tenant-2", "institution_two"))
			defer cancel()

			err := runner.Write(ctx, func(pgx.Tx) error { return tt.callback(cancel) })
			if err == nil {
				t.Fatalf("Write() expected error, got nil")
			}
			if !errors.Is(err, ErrTenantTransaction) && err.Error() != "callback failed" && err.Error() != "context canceled" {
				t.Fatalf("Write() unexpected error = %v", err)
			}
			if !tx.rolledBack || tx.committed {
				t.Fatalf("lifecycle committed=%t rolledBack=%t, want false/true", tx.committed, tx.rolledBack)
			}
		})
	}
}

func TestTenantTxRunnerRollsBackAndRepairsPanic(t *testing.T) {
	t.Parallel()

	tx := &fakeTx{}
	pool := &fakeTxStarter{tx: tx}
	runner := NewTenantTxRunner(pool, pool)

	defer func() {
		if recovered := recover(); recovered != "boom" {
			t.Fatalf("panic = %v, want boom", recovered)
		}
		if !tx.rolledBack || tx.committed {
			t.Fatalf("lifecycle committed=%t rolledBack=%t, want false/true", tx.committed, tx.rolledBack)
		}
	}()
	_ = runner.Write(tenantContext("tenant-3", "institution_three"), func(pgx.Tx) error {
		panic("boom")
	})
}

func tenantContext(id, schema string) context.Context {
	return tenantcontext.WithTenant(context.Background(), tenantcontext.TenantContext{
		ID: id, Schema: schema, CanonicalHost: "institution.example", Source: tenantcontext.DirectHost,
	})
}

type fakeTxStarter struct {
	tx         pgx.Tx
	beginErr   error
	beginCalls int
	options    pgx.TxOptions
}

func (p *fakeTxStarter) BeginTx(_ context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	p.beginCalls++
	p.options = options
	if p.beginErr != nil {
		return nil, p.beginErr
	}
	p.tx.(*fakeTx).calls = append(p.tx.(*fakeTx).calls, "begin")
	return p.tx, nil
}

type fakeTx struct {
	calls      []string
	execArgs   [][]any
	committed  bool
	rolledBack bool
}

func (tx *fakeTx) Begin(context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *fakeTx) Commit(context.Context) error {
	tx.committed = true
	tx.calls = append(tx.calls, "commit")
	return nil
}
func (tx *fakeTx) Rollback(context.Context) error { tx.rolledBack = true; return nil }
func (tx *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (tx *fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (tx *fakeTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (tx *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (tx *fakeTx) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.HasPrefix(sql, "SET LOCAL search_path") {
		tx.calls = append(tx.calls, sql)
	} else {
		tx.calls = append(tx.calls, "set_config")
	}
	tx.execArgs = append(tx.execArgs, args)
	return pgconn.CommandTag{}, nil
}
func (tx *fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (tx *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (tx *fakeTx) Conn() *pgx.Conn                                         { return nil }
