package bd_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bd "PlantillaInstitucional/app/db"
	"PlantillaInstitucional/configuracion/usuario"
	usuarioDB "PlantillaInstitucional/configuracion/usuario/repository"
	loginDB "PlantillaInstitucional/login/db"
	"PlantillaInstitucional/pagination"
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

	var tenantID string
	if err := conn.QueryRow(ctx, "SELECT tenant_id FROM institucion.cat_tenants WHERE subdomain = 'admin'").Scan(&tenantID); err != nil {
		t.Fatalf("read seeded tenant id: %v", err)
	}
	historicalAudit := createAuditRedactionFixture(t, ctx, conn, tenantID, "historical")
	assertAuditFixtureContainsSecrets(t, ctx, conn, historicalAudit)
	applyMigrationFile(t, ctx, conn, "000005_audit_secret_redaction.up.sql")
	assertAuditFixtureRedacted(t, ctx, conn, historicalAudit)
	assertAuditFixtureRedacted(t, ctx, conn, createAuditRedactionFixture(t, ctx, conn, historicalAudit.tenantID, "new"))

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
	verifyUsuarioRepositoryTenantTransaction(t, ctx, conn, databaseURL, tenantID)

	applyMigrationFile(t, ctx, conn, "000005_audit_secret_redaction.down.sql")
	assertAuditFixtureRedacted(t, ctx, conn, historicalAudit)
	assertAuditFixtureRedacted(t, ctx, conn, createAuditRedactionFixture(t, ctx, conn, tenantID, "rollback"))
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

type auditRedactionFixture struct {
	tenantID    string
	userID      string
	parameterID string
}

func createAuditRedactionFixture(t *testing.T, ctx context.Context, conn *pgx.Conn, tenantID, label string) auditRedactionFixture {
	t.Helper()
	userID := uuid.Must(uuid.NewV4()).String()
	parameterID := uuid.Must(uuid.NewV4()).String()
	compactID := strings.ReplaceAll(userID, "-", "")

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin %s audit fixture: %v", label, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		t.Fatalf("set %s audit fixture tenant: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO institucion.persona
			(uuid_persona, tenant_id, nombre, apellido1, clave_issemym, curp, rfc, razon_i)
		VALUES ($1::uuid, $2::uuid, $3, 'Audit', $4, $5, $6, 'Audit redaction test')`,
		userID, tenantID, label, "AUD"+compactID[:8], "AUD"+compactID[:15], "AUD"+compactID[:10]); err != nil {
		t.Fatalf("insert %s audit persona: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO sesion.usuario (uuid_persona, tenant_id, email, password, perfil, is_active)
		VALUES ($1::uuid, $2::uuid, $3, $4, 'operador', true)`, userID, tenantID, label+"-audit@example.test", label+"-insert-password"); err != nil {
		t.Fatalf("insert %s audited user: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `UPDATE sesion.usuario SET password = $2, email = $3 WHERE uuid_persona = $1::uuid`, userID, label+"-update-password", label+"-updated-audit@example.test"); err != nil {
		t.Fatalf("update %s audited user: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sesion.usuario WHERE uuid_persona = $1::uuid`, userID); err != nil {
		t.Fatalf("delete %s audited user: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO sistema.cfg_parametros (parametro_id, tenant_id, clave, valor, tipo_dato)
		VALUES ($1::uuid, $2::uuid, $3, $4, 'STRING')`, parameterID, tenantID, label+"-audit-secret", label+"-insert-value"); err != nil {
		t.Fatalf("insert %s audited parameter: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `UPDATE sistema.cfg_parametros SET valor = $2 WHERE parametro_id = $1::uuid`, parameterID, label+"-update-value"); err != nil {
		t.Fatalf("update %s audited parameter: %v", label, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sistema.cfg_parametros WHERE parametro_id = $1::uuid`, parameterID); err != nil {
		t.Fatalf("delete %s audited parameter: %v", label, err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit %s audit fixture: %v", label, err)
	}
	return auditRedactionFixture{tenantID: tenantID, userID: userID, parameterID: parameterID}
}

func assertAuditFixtureContainsSecrets(t *testing.T, ctx context.Context, conn *pgx.Conn, fixture auditRedactionFixture) {
	t.Helper()
	for _, target := range []struct{ table, id, secret string }{
		{table: "usuario", id: fixture.userID, secret: "password"},
		{table: "cfg_parametros", id: fixture.parameterID, secret: "valor"},
	} {
		var leakedSnapshots int
		if err := conn.QueryRow(ctx, `SELECT count(*) FROM sistema.bitacora_cambios WHERE nombre_tabla = $1 AND registro_pk = $2 AND (COALESCE(datos_viejos, '{}'::jsonb) ? $3 OR COALESCE(datos_nuevos, '{}'::jsonb) ? $3)`, target.table, target.id, target.secret).Scan(&leakedSnapshots); err != nil {
			t.Fatalf("read historical %s audit secrets: %v", target.table, err)
		}
		if leakedSnapshots != 3 {
			t.Fatalf("historical %s secret-bearing audit rows = %d, want 3", target.table, leakedSnapshots)
		}
	}
}

func assertAuditFixtureRedacted(t *testing.T, ctx context.Context, conn *pgx.Conn, fixture auditRedactionFixture) {
	t.Helper()
	for _, target := range []struct{ table, id, secret, retained string }{
		{table: "usuario", id: fixture.userID, secret: "password", retained: "email"},
		{table: "cfg_parametros", id: fixture.parameterID, secret: "valor", retained: "clave"},
	} {
		rows, err := conn.Query(ctx, `SELECT accion, tenant_id, datos_viejos, datos_nuevos FROM sistema.bitacora_cambios WHERE nombre_tabla = $1 AND registro_pk = $2 ORDER BY fecha_i, uuid_bitacora`, target.table, target.id)
		if err != nil {
			t.Fatalf("query redacted %s audit rows: %v", target.table, err)
		}
		actions := make(map[string]int)
		for rows.Next() {
			var action, tenantID string
			var oldData, newData []byte
			if err := rows.Scan(&action, &tenantID, &oldData, &newData); err != nil {
				rows.Close()
				t.Fatalf("scan redacted %s audit row: %v", target.table, err)
			}
			actions[action]++
			if tenantID != fixture.tenantID {
				t.Errorf("%s audit tenant_id = %q, want %q", target.table, tenantID, fixture.tenantID)
			}
			for _, snapshot := range [][]byte{oldData, newData} {
				if snapshot == nil {
					continue
				}
				var data map[string]any
				if err := json.Unmarshal(snapshot, &data); err != nil {
					t.Fatalf("decode %s audit snapshot: %v", target.table, err)
				}
				if _, exists := data[target.secret]; exists {
					t.Errorf("%s %s audit snapshot contains secret key %q", target.table, action, target.secret)
				}
				if _, exists := data[target.retained]; !exists {
					t.Errorf("%s %s audit snapshot lost non-secret key %q", target.table, action, target.retained)
				}
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			t.Fatalf("iterate redacted %s audit rows: %v", target.table, err)
		}
		if len(actions) != 3 || actions["INSERT"] != 1 || actions["UPDATE"] != 1 || actions["DELETE"] != 1 {
			t.Errorf("%s audit actions = %v, want one each of INSERT UPDATE DELETE", target.table, actions)
		}
	}

	var enabled, forced, policyExists bool
	if err := conn.QueryRow(ctx, `
		SELECT c.relrowsecurity, c.relforcerowsecurity,
			exists (SELECT 1 FROM pg_policies p WHERE p.schemaname = 'sistema' AND p.tablename = 'bitacora_cambios' AND p.policyname = 'tenant_isolation_bitacora')
		FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'sistema' AND c.relname = 'bitacora_cambios'`).Scan(&enabled, &forced, &policyExists); err != nil {
		t.Fatalf("read audit RLS state after redaction: %v", err)
	}
	if !enabled || !forced || !policyExists {
		t.Fatalf("audit RLS changed after redaction: enabled=%t forced=%t policy=%t", enabled, forced, policyExists)
	}
}

func verifyUsuarioRepositoryTenantTransaction(t *testing.T, ctx context.Context, conn *pgx.Conn, databaseURL, tenantID string) {
	t.Helper()

	otherTenantID := uuid.Must(uuid.NewV4()).String()
	personID := uuid.Must(uuid.NewV4())
	otherPersonID := uuid.Must(uuid.NewV4())
	roleName := "usuario_repo_test_" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	quotedRole := pgx.Identifier{roleName}.Sanitize()

	if _, err := conn.Exec(ctx, "CREATE ROLE "+quotedRole+" NOLOGIN NOBYPASSRLS"); err != nil {
		t.Fatalf("create usuario repository runtime role: %v", err)
	}
	defer func() {
		if _, err := conn.Exec(ctx, "DROP OWNED BY "+quotedRole); err != nil {
			t.Errorf("drop usuario repository runtime privileges: %v", err)
		}
		if _, err := conn.Exec(ctx, "DROP ROLE IF EXISTS "+quotedRole); err != nil {
			t.Errorf("drop usuario repository runtime role: %v", err)
		}
	}()
	for _, grant := range []string{
		"GRANT USAGE ON SCHEMA tenant_test, institucion, sesion, sistema TO " + quotedRole,
		"GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA institucion, sesion, sistema TO " + quotedRole,
		"GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA institucion, sesion, sistema TO " + quotedRole,
	} {
		if _, err := conn.Exec(ctx, grant); err != nil {
			t.Fatalf("grant usuario repository runtime privileges: %v", err)
		}
	}

	if _, err := conn.Exec(ctx, `
		INSERT INTO institucion.cat_tenants (tenant_id, subdomain, schema_name, nombre_depto)
		VALUES ($1::uuid, $2, $2, 'Other tenant')`, otherTenantID, "other-"+otherPersonID.String()[:8]); err != nil {
		t.Fatalf("insert other tenant: %v", err)
	}
	for _, person := range []struct {
		id       uuid.UUID
		tenantID string
		prefix   string
	}{
		{id: personID, tenantID: tenantID, prefix: "USR"},
		{id: otherPersonID, tenantID: otherTenantID, prefix: "OTH"},
	} {
		compactID := strings.ReplaceAll(person.id.String(), "-", "")
		if _, err := conn.Exec(ctx, `
			INSERT INTO institucion.persona
				(uuid_persona, tenant_id, nombre, apellido1, clave_issemym, curp, rfc, razon_i)
			VALUES ($1, $2::uuid, $3, 'Repository', $4, $5, $6, 'Integration test')`,
			person.id, person.tenantID, person.prefix, person.prefix+compactID[:8], person.prefix+compactID[:15], person.prefix+compactID[:10]); err != nil {
			t.Fatalf("insert %s tenant persona: %v", person.prefix, err)
		}
	}
	if _, err := conn.Exec(ctx, `
		INSERT INTO sesion.usuario (uuid_persona, tenant_id, email, password, perfil, is_active)
		VALUES ($1, $2::uuid, 'other-user@example.test', 'other-password', 'operador', true)`, otherPersonID, otherTenantID); err != nil {
		t.Fatalf("insert other tenant user: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse usuario repository pool config: %v", err)
	}
	poolConfig.MaxConns = 1
	poolConfig.AfterConnect = func(ctx context.Context, connection *pgx.Conn) error {
		_, err := connection.Exec(ctx, "SET ROLE "+quotedRole)
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("create usuario repository pool: %v", err)
	}
	defer pool.Close()

	repo := usuarioDB.NewUsuarioRepository(bd.NewAppPools(pool, pool))
	tenantCtx := tenantcontext.WithTenant(ctx, tenantcontext.TenantContext{ID: tenantID, Schema: "tenant_test", CanonicalHost: "tenant-test.example"})
	otherTenantCtx := tenantcontext.WithTenant(ctx, tenantcontext.TenantContext{ID: otherTenantID, Schema: "tenant_test", CanonicalHost: "other-tenant-test.example"})
	tokenHash := []byte("usuario-repository-token-hash-0001")
	created := usuario.Usuario{
		UuidPersona: personID,
		Email:       "tenant-user@example.test",
		Password:    "initial-password",
		Perfil:      "operador",
		Razon_i:     "Integration test",
	}
	if err := repo.CrearUsuario(tenantCtx, created, tokenHash, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create user through tenant transaction: %v", err)
	}

	criteria := pagination.CriteriosPagina[usuario.Usuario]{First: 0, Size: 10, SortedBy: "email", Direction: "ASC"}
	users, total, err := repo.ConsultarUsuarios(tenantCtx, criteria)
	if err != nil {
		t.Fatalf("list users through tenant transaction: %v", err)
	}
	foundCreated := false
	for _, listedUser := range users {
		if listedUser.UuidPersona == otherPersonID {
			t.Fatalf("tenant list exposed other tenant user %s", otherPersonID)
		}
		if listedUser.UuidPersona == personID {
			foundCreated = true
		}
	}
	if total != len(users) || !foundCreated {
		t.Fatalf("tenant list = total:%d users:%v, want created user %s and matching total", total, users, personID)
	}
	if _, err := repo.ConsultarUsuarioPorID(tenantCtx, otherPersonID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("cross-tenant read error = %v, want pgx.ErrNoRows", err)
	}

	created.Email = "updated-tenant-user@example.test"
	created.Password = "updated-password"
	created.Razon_m = "Integration update"
	if err := repo.ActualizarUsuario(tenantCtx, created); err != nil {
		t.Fatalf("update user through tenant transaction: %v", err)
	}
	if err := repo.ActualizarPassword(tenantCtx, personID, "final-password"); err != nil {
		t.Fatalf("update password through tenant transaction: %v", err)
	}
	if err := repo.ActualizarUsuario(tenantCtx, usuario.Usuario{UuidPersona: otherPersonID, Email: "blocked@example.test", Password: "blocked", Perfil: "operador"}); err == nil {
		t.Fatal("cross-tenant update unexpectedly succeeded")
	}

	updated, err := repo.ConsultarUsuarioPorID(tenantCtx, personID)
	if err != nil {
		t.Fatalf("read updated user through tenant transaction: %v", err)
	}
	if updated.Email != created.Email || updated.Password != "final-password" {
		t.Fatalf("updated user = email:%q password:%q", updated.Email, updated.Password)
	}
	if err := repo.EliminarUsuario(tenantCtx, personID); err != nil {
		t.Fatalf("delete user through tenant transaction: %v", err)
	}
	if _, err := repo.ConsultarUsuarioPorID(tenantCtx, personID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("deleted user read error = %v, want pgx.ErrNoRows", err)
	}
	otherUser, err := repo.ConsultarUsuarioPorID(otherTenantCtx, otherPersonID)
	if err != nil {
		t.Fatalf("read other tenant user after tenant delete: %v", err)
	}
	if otherUser.Email != "other-user@example.test" {
		t.Fatalf("other tenant user email = %q", otherUser.Email)
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
