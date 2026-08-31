package bd

import (
	"PlantillaInstitucional/shared/contextkey"
	"PlantillaInstitucional/shared/tenantcontext"
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTenantTxIntegration_ContextLeak(t *testing.T) {
	connString := os.Getenv("TEST_DATABASE_URL")
	if connString == "" {
		t.Skip("TEST_DATABASE_URL is not set. Skipping integration test.")
	}

	ctx := context.Background()
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	// Force a single connection pool to ensure the same connection is reused
	config.MaxConns = 1
	config.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	runner := NewTenantTxRunner(pool, pool)

	// Context 1
	ctx1 := tenantcontext.WithTenant(ctx, tenantcontext.TenantContext{
		ID: "tenant-1", Schema: "public",
	})
	ctx1 = context.WithValue(ctx1, contextkey.AuditUUIDKey, "uuid-1")

	// Context 2
	ctx2 := tenantcontext.WithTenant(ctx, tenantcontext.TenantContext{
		ID: "tenant-2", Schema: "public",
	})
	ctx2 = context.WithValue(ctx2, contextkey.AuditUUIDKey, "uuid-2")

	// Run first transaction
	err = runner.Read(ctx1, func(tx pgx.Tx) error {
		var tenantID string
		err := tx.QueryRow(ctx, "SELECT current_setting('app.tenant_id', true)").Scan(&tenantID)
		if err != nil {
			return err
		}
		if tenantID != "tenant-1" {
			t.Errorf("expected tenant-1, got %s", tenantID)
		}

		var auditUUID string
		err = tx.QueryRow(ctx, "SELECT current_setting('session_variables.uuid_persona', true)").Scan(&auditUUID)
		if err != nil {
			return err
		}
		if auditUUID != "uuid-1" {
			t.Errorf("expected uuid-1, got %s", auditUUID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx1 failed: %v", err)
	}

	// Wait briefly
	time.Sleep(10 * time.Millisecond)

	// Run second transaction on the exact same connection (due to MaxConns=1)
	err = runner.Read(ctx2, func(tx pgx.Tx) error {
		var tenantID string
		err := tx.QueryRow(ctx, "SELECT current_setting('app.tenant_id', true)").Scan(&tenantID)
		if err != nil {
			return err
		}
		if tenantID != "tenant-2" {
			t.Errorf("expected tenant-2, got %s (context leak!)", tenantID)
		}

		var auditUUID string
		err = tx.QueryRow(ctx, "SELECT current_setting('session_variables.uuid_persona', true)").Scan(&auditUUID)
		if err != nil {
			return err
		}
		if auditUUID != "uuid-2" {
			t.Errorf("expected uuid-2, got %s (context leak!)", auditUUID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx2 failed: %v", err)
	}
}
