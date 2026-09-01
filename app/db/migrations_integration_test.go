package bd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func TestMigrationsIntegration_SeedRollbackAndTenantAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL migration integration test in short mode")
	}

	databaseURL := os.Getenv("MIGRATION_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("MIGRATION_TEST_DATABASE_URL is not set; skipping PostgreSQL migration integration test")
	}

	ctx := context.Background()
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse migration test database URL: %v", err)
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatalf("connect migration test database: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(ctx) })

	var databaseName string
	if err := conn.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		t.Fatalf("read migration test database name: %v", err)
	}
	if !strings.Contains(strings.ToLower(databaseName), "test") {
		t.Skipf("refusing destructive migration test against database %q", databaseName)
	}

	t.Cleanup(func() { cleanupMigrationSchemas(t, ctx, conn) })
	cleanupMigrationSchemas(t, ctx, conn)

	applyMigrationFile(t, ctx, conn, "000001_init.up.sql")
	applyMigrationFile(t, ctx, conn, "000002_seed.up.sql")
	applyMigrationFile(t, ctx, conn, "000002_seed.down.sql")

	for _, query := range []string{
		"SELECT count(*) FROM sesion.politica",
		"SELECT count(*) FROM sesion.perfil",
		"SELECT count(*) FROM sesion.recurso",
		"SELECT count(*) FROM sesion.cat_accion",
		"SELECT count(*) FROM sistema.cfg_parametros",
		"SELECT count(*) FROM institucion.cat_tenants",
	} {
		var count int
		if err := conn.QueryRow(ctx, query).Scan(&count); err != nil {
			t.Fatalf("query seed rollback result %q: %v", query, err)
		}
		if count != 0 {
			t.Fatalf("seed rollback left %d rows for query %q", count, query)
		}
	}

	applyMigrationFile(t, ctx, conn, "000002_seed.up.sql")
	applyMigrationFile(t, ctx, conn, "000003_tenant_isolation_expand.up.sql")

	var rlsEnabled, forceRLS bool
	err = conn.QueryRow(ctx, `
		SELECT c.relrowsecurity, c.relforcerowsecurity
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'sistema' AND c.relname = 'bitacora_cambios'`).Scan(&rlsEnabled, &forceRLS)
	if err != nil {
		t.Fatalf("read bitacora RLS state: %v", err)
	}
	if !rlsEnabled || !forceRLS {
		t.Fatalf("bitacora_cambios must retain enabled FORCE RLS, got enabled=%t force=%t", rlsEnabled, forceRLS)
	}

	var tenantID string
	if err := conn.QueryRow(ctx, "SELECT tenant_id FROM institucion.cat_tenants WHERE subdomain = 'admin'").Scan(&tenantID); err != nil {
		t.Fatalf("read seeded tenant id: %v", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tenant audit transaction: %v", err)
	}
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		t.Fatalf("set transaction-local tenant id: %v", err)
	}

	personaID := uuid.Must(uuid.NewV4()).String()
	_, err = tx.Exec(ctx, `
		INSERT INTO institucion.persona (uuid_persona, tenant_id, nombre, apellido1, clave_issemym, curp, rfc, razon_i)
		VALUES ($1::uuid, $2::uuid, 'Migration', 'Test', $3, $4, $5, 'Integration test')`,
		personaID, tenantID, "TEST-"+personaID[:8], "TEST"+strings.ReplaceAll(personaID[:14], "-", ""), "TEST"+strings.ReplaceAll(personaID[:9], "-", ""))
	if err != nil {
		t.Fatalf("insert tenant-scoped row and audit record: %v", err)
	}

	var auditTenantID string
	if err := tx.QueryRow(ctx, "SELECT tenant_id FROM sistema.bitacora_cambios WHERE registro_pk = $1", personaID).Scan(&auditTenantID); err != nil {
		t.Fatalf("read tenant-aware audit record: %v", err)
	}
	if auditTenantID != tenantID {
		t.Fatalf("audit tenant_id = %q, want %q", auditTenantID, tenantID)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback tenant audit transaction: %v", err)
	}

	applyMigrationFile(t, ctx, conn, "000003_tenant_isolation_expand.down.sql")
	var tenantIDColumnExists bool
	if err := conn.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'sistema' AND table_name = 'bitacora_cambios' AND column_name = 'tenant_id'
		)`).Scan(&tenantIDColumnExists); err != nil {
		t.Fatalf("verify non-destructive expand rollback: %v", err)
	}
	if !tenantIDColumnExists {
		t.Fatal("expand rollback must retain the tenant_id column")
	}
}

func applyMigrationFile(t *testing.T, ctx context.Context, conn *pgx.Conn, name string) {
	t.Helper()
	path := filepath.Join("..", "..", "sql", "migrations", name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	if _, err := conn.Exec(ctx, string(content)); err != nil {
		t.Fatalf("apply migration %s: %v", name, err)
	}
}

func cleanupMigrationSchemas(t *testing.T, ctx context.Context, conn *pgx.Conn) {
	t.Helper()
	if _, err := conn.Exec(ctx, "DROP SCHEMA IF EXISTS sistema, sesion, institucion CASCADE"); err != nil {
		t.Fatalf("clean migration test schemas: %v", err)
	}
}
