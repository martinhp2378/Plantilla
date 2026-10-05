package bd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestBootstrapRolesContract(t *testing.T) {
	rolesSQLPath := filepath.Join("..", "..", "sql", "bootstrap", "roles.sql")
	contentBytes, err := os.ReadFile(rolesSQLPath)
	if err != nil {
		t.Fatalf("failed to read roles.sql: %v", err)
	}
	content := string(contentBytes)

	t.Run("Roles specify NOBYPASSRLS", func(t *testing.T) {
		if !strings.Contains(content, "app_runtime_tenant") {
			t.Fatal("roles.sql must define app_runtime_tenant")
		}
		if !strings.Contains(content, "app_migrator") {
			t.Fatal("roles.sql must define app_migrator")
		}
		// Must enforce NOBYPASSRLS
		if !strings.Contains(content, "NOBYPASSRLS") {
			t.Fatal("roles.sql must enforce NOBYPASSRLS on runtime and migration roles")
		}
		if matched, _ := regexp.MatchString(`(?i)[^\w]BYPASSRLS\b`, content); matched {
			t.Fatal("roles.sql must never grant BYPASSRLS without NO prefix")
		}
	})

	t.Run("Revokes public permissions fail-closed", func(t *testing.T) {
		if !strings.Contains(content, "REVOKE ALL ON SCHEMA public FROM PUBLIC") {
			t.Fatal("roles.sql must revoke public schema privileges from PUBLIC")
		}
		if !strings.Contains(content, "REVOKE ALL ON SCHEMA public FROM app_runtime_tenant") {
			t.Fatal("roles.sql must explicitly revoke public schema access from app_runtime_tenant")
		}
	})

	t.Run("Grants least-privilege DML to app_runtime_tenant", func(t *testing.T) {
		schemas := []string{"institucion", "sesion", "sistema"}
		for _, s := range schemas {
			if !strings.Contains(content, s) {
				t.Fatalf("roles.sql must configure schema %s", s)
			}
		}
		if !strings.Contains(content, "ALTER DEFAULT PRIVILEGES") {
			t.Fatal("roles.sql must configure default privileges for future objects")
		}
	})
}

func TestTenantIsolationExpandMigration(t *testing.T) {
	upSQLPath := filepath.Join("..", "..", "sql", "migrations", "000003_tenant_isolation_expand.up.sql")
	upBytes, err := os.ReadFile(upSQLPath)
	if err != nil {
		t.Fatalf("failed to read 000003 up migration: %v", err)
	}
	upContent := string(upBytes)

	downSQLPath := filepath.Join("..", "..", "sql", "migrations", "000003_tenant_isolation_expand.down.sql")
	downBytes, err := os.ReadFile(downSQLPath)
	if err != nil {
		t.Fatalf("failed to read 000003 down migration: %v", err)
	}
	downContent := string(downBytes)

	tenantTables := []struct {
		schema string
		table  string
	}{
		{schema: "institucion", table: "persona"},
		{schema: "sesion", table: "usuario"},
		{schema: "sesion", table: "sesion"},
		{schema: "sesion", table: "prelogin"},
		{schema: "sistema", table: "bitacora_cambios"},
		{schema: "sistema", table: "cfg_parametros"},
	}

	t.Run("Enforces FORCE ROW LEVEL SECURITY on all tenant tables", func(t *testing.T) {
		for _, tt := range tenantTables {
			fullTable := tt.schema + "." + tt.table
			enableRLS := "ALTER TABLE " + fullTable + " ENABLE ROW LEVEL SECURITY;"
			forceRLS := "ALTER TABLE " + fullTable + " FORCE ROW LEVEL SECURITY;"

			if !strings.Contains(upContent, enableRLS) {
				t.Errorf("migration missing %s", enableRLS)
			}
			if !strings.Contains(upContent, forceRLS) {
				t.Errorf("migration missing %s", forceRLS)
			}
		}
	})

	t.Run("Enforces tenant_id comparison and WITH CHECK in policies", func(t *testing.T) {
		settingRef := "current_setting('app.tenant_id', true)"
		if !strings.Contains(upContent, settingRef) {
			t.Fatalf("policies must reference %s", settingRef)
		}
		if !strings.Contains(upContent, "WITH CHECK") {
			t.Fatal("policies must enforce WITH CHECK on modifications")
		}
	})

	t.Run("Creates indexes on tenant_id", func(t *testing.T) {
		for _, tt := range tenantTables {
			if tt.table == "prelogin" {
				continue
			}
			idxRef := "idx_" + tt.table + "_tenant_id"
			if !strings.Contains(upContent, idxRef) {
				t.Errorf("migration missing index %s", idxRef)
			}
		}
	})

	t.Run("Down migration is non-destructive", func(t *testing.T) {
		if strings.Contains(strings.ToUpper(downContent), "DROP COLUMN") {
			t.Fatal("down migration must follow expand-contract forward fix: do not DROP COLUMN destructively")
		}
		if !strings.Contains(downContent, "DROP POLICY") {
			t.Fatal("down migration must drop policies")
		}
		if !strings.Contains(downContent, "DISABLE ROW LEVEL SECURITY") {
			t.Fatal("down migration must disable RLS")
		}
	})

	t.Run("Audit function derives and stores the transaction-local tenant", func(t *testing.T) {
		if !strings.Contains(upContent, "CREATE OR REPLACE FUNCTION sistema.fn_registrar_auditoria()") {
			t.Fatal("migration must replace the audit function after tenant_id is introduced")
		}
		if !strings.Contains(upContent, "NULLIF(current_setting('app.tenant_id', true), '')::UUID") {
			t.Fatal("audit function must safely derive tenant_id from the transaction-local setting")
		}
		if strings.Count(upContent, "INSERT INTO sistema.bitacora_cambios (tenant_id,") != 3 {
			t.Fatal("every audit operation must insert tenant_id")
		}
		if !strings.Contains(downContent, "INSERT INTO sistema.bitacora_cambios (uuid_persona,") {
			t.Fatal("down migration must restore the pre-expand audit function")
		}
	})
}

func TestSeedMigrationContract(t *testing.T) {
	upSQLPath := filepath.Join("..", "..", "sql", "migrations", "000002_seed.up.sql")
	upBytes, err := os.ReadFile(upSQLPath)
	if err != nil {
		t.Fatalf("failed to read 000002 up migration: %v", err)
	}
	upContent := string(upBytes)

	downSQLPath := filepath.Join("..", "..", "sql", "migrations", "000002_seed.down.sql")
	downBytes, err := os.ReadFile(downSQLPath)
	if err != nil {
		t.Fatalf("failed to read 000002 down migration: %v", err)
	}
	downContent := string(downBytes)

	t.Run("Does not ship a known superadmin credential", func(t *testing.T) {
		for _, forbidden := range []string{"admin@issemym.gob.mx", "admin123", "WJK.5U64OTFkWlAGeTFAUePwpsxGnB5w6Lqbox72QkjbB3RtxqSbm"} {
			if strings.Contains(upContent, forbidden) {
				t.Fatalf("seed migration must not contain known superadmin credential material: %q", forbidden)
			}
		}
	})

	t.Run("Deletes seeded foreign-key dependents before catalogs", func(t *testing.T) {
		orderedDeletes := []string{
			"DELETE FROM sesion.politica",
			"DELETE FROM sesion.perfil",
			"DELETE FROM sesion.recurso",
			"DELETE FROM sesion.cat_accion",
		}
		last := -1
		for _, statement := range orderedDeletes {
			index := strings.Index(downContent, statement)
			if index == -1 {
				t.Fatalf("seed down migration missing %s", statement)
			}
			if index <= last {
				t.Fatalf("seed down migration must delete %s after its dependents", statement)
			}
			last = index
		}
	})

	t.Run("Deletes the complete seeded catalog set", func(t *testing.T) {
		for _, seeded := range []string{
			"'Acceso Usuarios'", "'EMAIL_SYSTEM_NAME'", "'admin'", "'anonimo'", "'parametros'", "'aprobar'",
		} {
			if !strings.Contains(downContent, seeded) {
				t.Fatalf("seed down migration must reverse seeded row %s", seeded)
			}
		}
	})
}

func TestMigrationVersionsAreUnique(t *testing.T) {
	migrationDir := filepath.Join("..", "..", "sql", "migrations")
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		t.Fatalf("read migrations directory: %v", err)
	}

	migrationPattern := regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)
	versions := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationPattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		version, name := matches[1], matches[2]
		if existing, exists := versions[version]; exists && existing != name {
			t.Fatalf("migration version %s is used by both %q and %q", version, existing, name)
		}
		versions[version] = name
	}
}

func TestDedicatedAdminSchemaMigration(t *testing.T) {
	upPath := filepath.Join("..", "..", "sql", "migrations", "000004_dedicated_admin_schema.up.sql")
	upBytes, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("failed to read admin schema up migration: %v", err)
	}
	upContent := string(upBytes)

	downPath := filepath.Join("..", "..", "sql", "migrations", "000004_dedicated_admin_schema.down.sql")
	downBytes, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("failed to read admin schema down migration: %v", err)
	}
	downContent := string(downBytes)

	if !strings.Contains(upContent, "CREATE SCHEMA IF NOT EXISTS admin") {
		t.Fatal("admin schema migration must create a dedicated admin schema")
	}
	if !strings.Contains(upContent, "SET schema_name = 'admin'") {
		t.Fatal("admin schema migration must move the bootstrap tenant to admin")
	}
	if !strings.Contains(upContent, "GRANT USAGE ON SCHEMA admin") {
		t.Fatal("admin schema migration must grant runtime schema usage")
	}
	if !strings.Contains(downContent, "SET schema_name = 'public'") || !strings.Contains(downContent, "DROP SCHEMA IF EXISTS admin") {
		t.Fatal("admin schema down migration must reverse the schema move")
	}
}
