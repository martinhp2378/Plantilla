package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStringTransformations(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSnake  string
		wantPascal string
		wantPlural string
	}{
		{
			name:       "simple singular noun ending in consonant",
			input:      "proveedor",
			wantSnake:  "proveedor",
			wantPascal: "Proveedor",
			wantPlural: "proveedores",
		},
		{
			name:       "hyphenated module name",
			input:      "orden-compra",
			wantSnake:  "orden_compra",
			wantPascal: "OrdenCompra",
			wantPlural: "orden_compras",
		},
		{
			name:       "camelCase module name",
			input:      "TipoContrato",
			wantSnake:  "tipo_contrato",
			wantPascal: "TipoContrato",
			wantPlural: "tipo_contratos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSnake := toSnakeCase(tt.input)
			if gotSnake != tt.wantSnake {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, gotSnake, tt.wantSnake)
			}
			gotPascal := toPascalCase(gotSnake)
			if gotPascal != tt.wantPascal {
				t.Errorf("toPascalCase(%q) = %q, want %q", gotSnake, gotPascal, tt.wantPascal)
			}
			gotPlural := autoPlural(gotSnake)
			if gotPlural != tt.wantPlural {
				t.Errorf("autoPlural(%q) = %q, want %q", gotSnake, gotPlural, tt.wantPlural)
			}
		})
	}
}

func TestGetModuleName(t *testing.T) {
	tmp := t.TempDir()
	goMod := filepath.Join(tmp, "go.mod")

	if err := os.WriteFile(goMod, []byte("module SistemaInventario\n\ngo 1.26.2\n"), 0600); err != nil {
		t.Fatalf("failed writing go.mod: %v", err)
	}

	if got := getModuleName(goMod); got != "SistemaInventario" {
		t.Errorf("getModuleName() = %q, want %q", got, "SistemaInventario")
	}

	if got := getModuleName(filepath.Join(tmp, "missing.mod")); got != "PlantillaInstitucional" {
		t.Errorf("fallback getModuleName() = %q, want %q", got, "PlantillaInstitucional")
	}
}

func TestMountAndRemoveModuleLifecycle(t *testing.T) {
	tmp := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmp, "cmd"), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "sql", "migrations"), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "proveedor", "service"), 0750); err != nil {
		t.Fatal(err)
	}

	upMigr := filepath.Join(tmp, "sql", "migrations", "000011_create_proveedor_table.up.sql")
	if err := os.WriteFile(upMigr, []byte("-- up"), 0600); err != nil {
		t.Fatal(err)
	}

	mainPath := filepath.Join(tmp, "cmd", "main.go")
	initialMain := `package main

import (
	miperfil_handler "MiSistema/miperfil/handler"
)

func main() {
			miperfil_handler.Register(a, r)
}
`
	if err := os.WriteFile(mainPath, []byte(initialMain), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		Module: "MiSistema",
		Name:   "proveedor",
	}

	if err := mountInMainFile(mainPath, cfg); err != nil {
		t.Fatalf("mountInMainFile failed: %v", err)
	}

	// #nosec G304 -- test path inside t.TempDir
	mountedBytes, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	mounted := string(mountedBytes)
	if !strings.Contains(mounted, `proveedor_handler "MiSistema/proveedor/handler"`) {
		t.Errorf("expected import in cmd/main.go, got:\n%s", mounted)
	}
	if !strings.Contains(mounted, "proveedor_handler.Register(a, r)") {
		t.Errorf("expected Register call in cmd/main.go, got:\n%s", mounted)
	}

	if err := removeModule(tmp, "proveedor", "MiSistema"); err != nil {
		t.Fatalf("removeModule failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "proveedor")); !os.IsNotExist(err) {
		t.Errorf("expected proveedor directory to be removed")
	}
	if _, err := os.Stat(upMigr); !os.IsNotExist(err) {
		t.Errorf("expected migration file to be removed")
	}

	// #nosec G304 -- test path inside t.TempDir
	cleanedBytes, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	cleaned := string(cleanedBytes)
	if strings.Contains(cleaned, "proveedor_handler") {
		t.Errorf("expected proveedor_handler to be removed from cmd/main.go, got:\n%s", cleaned)
	}
}

func TestRemoveModuleProtectsCoreDirectories(t *testing.T) {
	tmp := t.TempDir()
	if err := removeModule(tmp, "shared", "PlantillaInstitucional"); err == nil {
		t.Fatal("expected error when attempting to remove protected core directory 'shared'")
	}
}

func TestRenameProjectModule(t *testing.T) {
	tmp := t.TempDir()
	goMod := filepath.Join(tmp, "go.mod")
	sampleGo := filepath.Join(tmp, "sample.go")

	if err := os.WriteFile(goMod, []byte("module PlantillaInstitucional\n\ngo 1.26.2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sampleGo, []byte("package main\nimport \"PlantillaInstitucional/shared/entorno\"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := renameProjectModule(tmp, "ControlVehicular"); err != nil {
		t.Fatalf("renameProjectModule failed: %v", err)
	}

	if got := getModuleName(goMod); got != "ControlVehicular" {
		t.Errorf("module name after rename = %q, want %q", got, "ControlVehicular")
	}

	// #nosec G304 -- test path inside t.TempDir
	updatedGo, err := os.ReadFile(sampleGo)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updatedGo), `"ControlVehicular/shared/entorno"`) {
		t.Errorf("expected updated import path, got:\n%s", string(updatedGo))
	}
}

func TestEnsureEnvFile(t *testing.T) {
	tmp := t.TempDir()
	envExample := `AMBIENTE=LOCAL
SERVICIO=plantilla
PQ_WRITE_DB=plantilla
PQ_WRITE_USER=usuplantilla
PQ_WRITE_PASSWORD=
PQ_READ_DB=plantilla
PQ_READ_USER=usuplantilla
PQ_READ_PASSWORD=
COOKIE_NOMBRE=plantilla_v2
`
	if err := os.WriteFile(filepath.Join(tmp, ".env.example"), []byte(envExample), 0600); err != nil {
		t.Fatal(err)
	}

	if err := ensureEnvFile(tmp, "inventario", "postgres", "secreto123"); err != nil {
		t.Fatalf("ensureEnvFile failed: %v", err)
	}

	// #nosec G304 -- test path inside t.TempDir
	envBytes, err := os.ReadFile(filepath.Join(tmp, ".env"))
	if err != nil {
		t.Fatal(err)
	}
	envContent := string(envBytes)

	for _, expected := range []string{
		"SERVICIO=inventario",
		"PQ_WRITE_DB=inventario",
		"PQ_READ_DB=inventario",
		"PQ_WRITE_USER=postgres",
		"PQ_WRITE_PASSWORD=secreto123",
		"COOKIE_NOMBRE=inventario_v2",
	} {
		if !strings.Contains(envContent, expected) {
			t.Errorf("expected .env to contain %q, got:\n%s", expected, envContent)
		}
	}
}

func TestPromptInputAndExtractLeadingPositional(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("MiSistema\n\n"))
	got := promptInput(reader, "Proyecto", "DefaultMod")
	if got != "MiSistema" {
		t.Errorf("promptInput() = %q, want %q", got, "MiSistema")
	}

	gotDefault := promptInput(reader, "Servicio", "sistema")
	if gotDefault != "sistema" {
		t.Errorf("promptInput() with empty line = %q, want %q", gotDefault, "sistema")
	}

	pos, rest := extractLeadingPositional([]string{"orden-compra", "-schema", "institucion"})
	if pos != "orden-compra" || len(rest) != 2 {
		t.Errorf("extractLeadingPositional() = (%q, %v), want ('orden-compra', ['-schema', 'institucion'])", pos, rest)
	}
}
