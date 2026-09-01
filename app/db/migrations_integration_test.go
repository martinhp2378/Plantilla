package bd_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bd "PlantillaInstitucional/app/db"
	loginDB "PlantillaInstitucional/login/db"
	"PlantillaInstitucional/shared/tenantcontext"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	applyMigrationFile(t, ctx, conn, "000004_account_activation.up.sql")

	var activationRLS bool
	if err := conn.QueryRow(ctx, `SELECT relrowsecurity FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'sesion' AND c.relname = 'password_reset_token'`).Scan(&activationRLS); err != nil {
		t.Fatalf("read activation token RLS state: %v", err)
	}
	if !activationRLS {
		t.Fatal("password_reset_token must enable RLS")
	}

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

	verifyActivationTokenTenantTransaction(t, ctx, conn, databaseURL, tenantID)

	applyMigrationFile(t, ctx, conn, "000004_account_activation.down.sql")
	var activationTableExists bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'sesion' AND table_name = 'password_reset_token')`).Scan(&activationTableExists); err != nil {
		t.Fatalf("verify activation rollback: %v", err)
	}
	if activationTableExists {
		t.Fatal("activation rollback must remove password_reset_token")
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

func verifyActivationTokenTenantTransaction(t *testing.T, ctx context.Context, conn *pgx.Conn, databaseURL, tenantID string) {
	t.Helper()
	if _, err := conn.Exec(ctx, `CREATE SCHEMA tenant_test`); err != nil {
		t.Fatalf("create tenant search-path schema: %v", err)
	}

	personID := uuid.Must(uuid.NewV4()).String()
	setup, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin activation setup: %v", err)
	}
	defer func() { _ = setup.Rollback(ctx) }()
	if _, err := setup.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		t.Fatalf("set activation tenant: %v", err)
	}
	if _, err := setup.Exec(ctx, `
		INSERT INTO institucion.persona (uuid_persona, tenant_id, nombre, apellido1, clave_issemym, curp, rfc, razon_i)
		VALUES ($1::uuid, $2::uuid, 'Activation', 'Test', $3, $4, $5, 'Integration test')`,
		personID, tenantID, "ACT-"+personID[:8], "ACT"+strings.ReplaceAll(personID[:14], "-", ""), "ACT"+strings.ReplaceAll(personID[:9], "-", "")); err != nil {
		t.Fatalf("insert activation persona: %v", err)
	}
	if _, err := setup.Exec(ctx, `INSERT INTO sesion.usuario (uuid_persona, tenant_id, email, password, perfil, is_active) VALUES ($1::uuid, $2::uuid, 'activation@example.test', 'disabled', 'operador', false)`, personID, tenantID); err != nil {
		t.Fatalf("insert inactive activation user: %v", err)
	}
	if err := setup.Commit(ctx); err != nil {
		t.Fatalf("commit activation setup: %v", err)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create tenant transaction pool: %v", err)
	}
	t.Cleanup(pool.Close)
	repo := loginDB.NewRepositorioPsql(bd.NewAppPools(pool, pool))
	tenantCtx := tenantcontext.WithTenant(ctx, tenantcontext.TenantContext{ID: tenantID, Schema: "tenant_test", CanonicalHost: "tenant-test.example"})
	tokenHash := []byte("01234567890123456789012345678901")
	if found, err := repo.IssuePasswordReset(tenantCtx, "activation@example.test", tokenHash, time.Now().Add(time.Hour)); err != nil || !found {
		t.Fatalf("issue token through tenant transaction = found:%t err:%v", found, err)
	}
	if err := repo.ConsumePasswordReset(tenantCtx, tokenHash, "updated-password-hash"); err != nil {
		t.Fatalf("consume token through tenant transaction: %v", err)
	}
	if err := repo.ConsumePasswordReset(tenantCtx, tokenHash, "another-password-hash"); !errors.Is(err, loginDB.ErrInvalidOrExpiredResetToken) {
		t.Fatalf("replay token error = %v, want ErrInvalidOrExpiredResetToken", err)
	}

	var active bool
	var passwordValue string
	if err := conn.QueryRow(ctx, `SELECT is_active, password FROM sesion.usuario WHERE uuid_persona = $1::uuid`, personID).Scan(&active, &passwordValue); err != nil {
		t.Fatalf("read consumed activation user: %v", err)
	}
	if !active || passwordValue != "updated-password-hash" {
		t.Fatalf("consumed activation user = active:%t password:%q", active, passwordValue)
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
	if _, err := conn.Exec(ctx, "DROP SCHEMA IF EXISTS tenant_test, sistema, sesion, institucion CASCADE"); err != nil {
		t.Fatalf("clean migration test schemas: %v", err)
	}
}
