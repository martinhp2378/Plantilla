package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	defaultTemplateRepo = "https://github.com/issemym-repos/PlantillaInstitucional.git"
	jrmpVersion         = "v1.0.0"
)

type Config struct {
	Module      string // e.g. PlantillaInstitucional
	Name        string // e.g. proveedor
	NameUpper   string // e.g. Proveedor
	Plural      string // e.g. proveedores
	PluralUpper string // e.g. Proveedores
	Schema      string // e.g. institucion
	Prefix      string // e.g. /institucion/proveedores
	Table       string // e.g. institucion.proveedor
	MigrNum     string // e.g. 000011
}

func main() {
	args := os.Args[1:]
	isTTY := isTerminalFile(os.Stdin) && isTerminalFile(os.Stdout) && os.Getenv("JRMP_NO_TUI") != "1"
	if shouldLaunchTUI(args, isTTY) {
		runTUI()
		return
	}
	if len(args) == 0 {
		printHelp()
		return
	}

	// Backward compatibility when invoked directly with flags (e.g., -name or -init-project)
	if strings.HasPrefix(args[0], "-") {
		if args[0] == "-h" || args[0] == "--help" {
			printHelp()
			return
		}
		if args[0] == "-v" || args[0] == "--version" {
			fmt.Printf("jrmp %s\n", jrmpVersion)
			return
		}
		runLegacyFlags(args)
		return
	}

	subcmd := strings.ToLower(strings.TrimSpace(args[0]))
	subArgs := args[1:]

	if subcmd != "create" && subcmd != "new-project" && subcmd != "help" && subcmd != "version" && subcmd != "tui" && subcmd != "ui" {
		ensureProjectRoot()
	}

	switch subcmd {
	case "tui", "ui":
		runTUI()
	case "create", "new-project":
		runCreateProject(subArgs, os.Stdin)
	case "init", "init-project":
		runInitProject(subArgs, os.Stdin)
	case "new-module", "add":
		runNewModule(subArgs, os.Stdin)
	case "remove-module", "rm":
		runRemoveModule(subArgs, os.Stdin)
	case "check":
		runCheck()
	case "templ-fmt":
		mustRunCmd(".", "templ", "fmt", ".")
	case "templ-check":
		mustRunCmd(".", "templ", "fmt", "-fail", ".")
		mustRunCmd(".", "templ", "generate")
	case "lint":
		mustRunCmd(".", "golangci-lint", "run", "./...")
	case "test":
		mustRunCmd(".", "go", "test", "./...")
	case "dev":
		runDev()
	case "up":
		mustRunCmd(".", "docker", "compose", "-f", "docker-compose.local.yml", "up", "--build", "-d")
	case "down":
		mustRunCmd(".", "docker", "compose", "-f", "docker-compose.local.yml", "down")
	case "logs":
		mustRunCmd(".", "docker", "compose", "-f", "docker-compose.local.yml", "logs", "-f", "app")
	case "version":
		fmt.Printf("jrmp %s\n", jrmpVersion)
	case "help":
		printHelp()
	default:
		fmt.Printf("❌ Comando desconocido: '%s'\n\n", subcmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`🏛️  JRMP CLI — Herramienta Oficial de la Plantilla Institucional (` + jrmpVersion + `)

Uso:
  jrmp                                  Abre la interfaz interactiva TUI (estilo gentle-ai)
  jrmp tui                              Abre la interfaz interactiva TUI explícitamente
  jrmp <comando> [argumentos]           Ejecuta un comando directo sin abrir la TUI

🌱 Creación e Inicialización de Proyectos:
  jrmp create [NombreProyecto]          Clona la plantilla desde cero, desvincula Git, renombra el proyecto y crea el .env (interactivo)
  jrmp init [NombreProyecto]            Inicializa un repositorio ya clonado manualmente (renombra módulo Go, desvincula origin y crea .env)

⚡ Generador de Módulos CRUD (Scaffolding):
  jrmp new-module [nombre] [opciones]   Crea un módulo CRUD completo (tabla, modales, PDF/Excel, tests, migraciones y menú)
                                        Opciones: -schema sistema -title "Título con Acentos" -plural "..."
  jrmp remove-module [nombre]           Elimina un módulo, sus migraciones, su registro en main.go y revierte su tabla en PostgreSQL

🧰 Desarrollo, Calidad y Contenedores (Mac, Linux y Windows):
  jrmp up                               Enciende la aplicación y PostgreSQL en Docker con recarga en vivo
  jrmp down                             Apaga los contenedores de Docker
  jrmp logs                             Muestra los logs del servidor en vivo
  jrmp dev                              Ejecuta Air localmente fuera de Docker (compatible con Windows .exe, Mac y Linux)
  jrmp check                            Ejecuta la validación completa (templ-check + lint + test)
  jrmp templ-fmt                        Formatea todos los archivos .templ
  jrmp templ-check                      Verifica formato .templ y ejecuta templ generate
  jrmp lint                             Ejecuta golangci-lint run ./...
  jrmp test                             Ejecuta go test ./...`)
}

func promptInput(r *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return defaultVal
	}
	val := strings.TrimSpace(line)
	if val == "" {
		return defaultVal
	}
	return val
}

func extractLeadingPositional(args []string) (string, []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return strings.TrimSpace(args[0]), args[1:]
	}
	return "", args
}

func ensureProjectRoot() {
	if _, err := os.Stat("cmd/main.go"); err == nil {
		return
	}
	if _, err := os.Stat("../../cmd/main.go"); err == nil {
		_ = os.Chdir("../..")
	}
}

func runCreateProject(args []string, in io.Reader) {
	posName, flagArgs := extractLeadingPositional(args)
	fsFlags := flag.NewFlagSet("create", flag.ExitOnError)
	repoFlag := fsFlags.String("repo", defaultTemplateRepo, "Template repository URL")
	serviceFlag := fsFlags.String("service", "", "Short service and database name")
	dbPassFlag := fsFlags.String("db-password", "", "Local PostgreSQL password")
	remoteFlag := fsFlags.String("remote", "", "New Git origin URL")
	upFlag := fsFlags.Bool("up", false, "Start Docker containers after creation")
	nonInteractiveFlag := fsFlags.Bool("y", false, "Use defaults without interactive prompts")
	_ = fsFlags.Parse(flagArgs)

	reader := bufio.NewReader(in)
	projectName := posName
	if projectName == "" && fsFlags.NArg() > 0 {
		projectName = strings.TrimSpace(fsFlags.Arg(0))
	}
	if projectName == "" {
		projectName = promptInput(reader, "📝 Nombre de tu nuevo proyecto (sin espacios, ej. SistemaInventario)", "")
	}
	projectName = strings.ReplaceAll(strings.TrimSpace(projectName), " ", "")
	if projectName == "" {
		fmt.Println("❌ Error: el nombre del proyecto es obligatorio.")
		os.Exit(1)
	}

	defaultSvc := toSnakeCase(projectName)
	if defaultSvc == "" {
		defaultSvc = "sistema"
	}

	serviceName := strings.TrimSpace(*serviceFlag)
	dbPass := strings.TrimSpace(*dbPassFlag)
	newRemote := strings.TrimSpace(*remoteFlag)
	startDocker := *upFlag

	if !*nonInteractiveFlag {
		if serviceName == "" {
			serviceName = toSnakeCase(promptInput(reader, "🗄️  Nombre corto para servicio y base de datos local", defaultSvc))
		}
		if dbPass == "" {
			dbPass = promptInput(reader, "🔑 Contraseña para la base de datos PostgreSQL local", "postgres")
		}
		if newRemote == "" {
			newRemote = promptInput(reader, "🔗 URL de tu nuevo repositorio Git (opcional, Enter para omitir)", "")
		}
		if !startDocker {
			ans := strings.ToLower(promptInput(reader, "🐳 ¿Deseas construir y encender el proyecto con Docker al finalizar? (s/N)", "n"))
			startDocker = ans == "s" || ans == "si" || ans == "sí" || ans == "y" || ans == "yes"
		}
	} else {
		if serviceName == "" {
			serviceName = defaultSvc
		}
		if dbPass == "" {
			dbPass = "postgres"
		}
	}

	if err := createProjectCore(*repoFlag, projectName, serviceName, dbPass, newRemote, startDocker); err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
}

func createProjectCore(repoURL, projectName, serviceName, dbPass, newRemote string, startDocker bool) error {
	projectName = strings.ReplaceAll(strings.TrimSpace(projectName), " ", "")
	if projectName == "" {
		return fmt.Errorf("el nombre del proyecto es obligatorio")
	}
	if repoURL == "" {
		repoURL = defaultTemplateRepo
	}
	if serviceName == "" {
		serviceName = toSnakeCase(projectName)
	}
	if dbPass == "" {
		dbPass = "postgres"
	}

	fmt.Printf("\n📥 1/4 Clonando Plantilla Institucional en ./%s...\n", projectName)
	if err := runCmdInDir(".", "git", "clone", repoURL, projectName); err != nil {
		return fmt.Errorf("error clonando repositorio: %w", err)
	}

	fmt.Println("🔗 2/4 Desvinculando repositorio de la plantilla original...")
	_ = runCmdInDir(projectName, "git", "remote", "remove", "origin")
	if newRemote != "" {
		if err := runCmdInDir(projectName, "git", "remote", "add", "origin", newRemote); err != nil {
			return fmt.Errorf("error agregando remote origin: %w", err)
		}
		fmt.Printf("✅ Conectado al nuevo repositorio: %s\n", newRemote)
	}

	fmt.Printf("🏷️  3/4 Renombrando módulo Go a '%s' y configurando .env...\n", projectName)
	if err := renameProjectModule(projectName, projectName); err != nil {
		return fmt.Errorf("error renombrando proyecto: %w", err)
	}
	if err := ensureEnvFile(projectName, serviceName, "postgres", dbPass); err != nil {
		fmt.Printf("⚠️ No se pudo generar .env automaticamente: %v\n", err)
	}

	fmt.Println("⚙️  4/4 Generando componentes Templ...")
	_ = runCmdInDir(projectName, "templ", "generate")

	fmt.Printf("\n✨ ¡Proyecto '%s' creado y configurado exitosamente!\n", projectName)
	fmt.Printf("📁 Carpeta: ./%s\n", projectName)

	if startDocker {
		fmt.Println("\n🐳 Construyendo y encendiendo contenedores con Docker...")
		if err := runCmdInDir(projectName, "docker", "compose", "-f", "docker-compose.local.yml", "up", "--build", "-d"); err != nil {
			return fmt.Errorf("error encendiendo Docker: %w", err)
		}
		fmt.Println("🌐 Listo en: http://localhost:8080/login (admin@issemym.gob.mx / admin123)")
	} else {
		fmt.Printf("\n👉 Siguientes pasos:\n   cd %s\n   jrmp up\n", projectName)
	}
	return nil
}

func runInitProject(args []string, in io.Reader) {
	posName, flagArgs := extractLeadingPositional(args)
	fsFlags := flag.NewFlagSet("init", flag.ExitOnError)
	modFlag := fsFlags.String("module", "", "New Go module name")
	serviceFlag := fsFlags.String("service", "", "Short service and database name")
	dbPassFlag := fsFlags.String("db-password", "postgres", "Local PostgreSQL password")
	_ = fsFlags.Parse(flagArgs)

	newMod := strings.TrimSpace(*modFlag)
	if newMod == "" {
		newMod = posName
	}
	if newMod == "" && fsFlags.NArg() > 0 {
		newMod = strings.TrimSpace(fsFlags.Arg(0))
	}
	if newMod == "" {
		reader := bufio.NewReader(in)
		newMod = promptInput(reader, "📝 Nombre de tu proyecto (sin espacios, ej. SistemaInventario)", "")
	}
	newMod = strings.ReplaceAll(strings.TrimSpace(newMod), " ", "")
	if newMod == "" {
		fmt.Println("❌ Error: el nombre del proyecto es obligatorio. Ej: jrmp init MiNuevoProyecto")
		os.Exit(1)
	}

	if err := initProjectCore(newMod, strings.TrimSpace(*serviceFlag), *dbPassFlag); err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
}

func initProjectCore(newMod, svc, dbPass string) error {
	newMod = strings.ReplaceAll(strings.TrimSpace(newMod), " ", "")
	if newMod == "" {
		return fmt.Errorf("el nombre del proyecto es obligatorio")
	}

	// Si el remote origin sigue apuntando a PlantillaInstitucional, desvincularlo por seguridad
	if out, err := exec.Command("git", "remote", "get-url", "origin").Output(); err == nil {
		if strings.Contains(string(out), "PlantillaInstitucional") {
			_ = exec.Command("git", "remote", "remove", "origin").Run()
			fmt.Println("🔗 Vínculo con el repositorio original PlantillaInstitucional eliminado por seguridad.")
		}
	}

	if err := renameProjectModule(".", newMod); err != nil {
		return fmt.Errorf("error renombrando proyecto: %w", err)
	}

	if svc == "" {
		svc = toSnakeCase(newMod)
	}
	if dbPass == "" {
		dbPass = "postgres"
	}
	if err := ensureEnvFile(".", svc, "postgres", dbPass); err != nil {
		fmt.Printf("⚠️ Advertencia al preparar .env: %v\n", err)
	}

	fmt.Println("⚙️ Ejecutando templ generate...")
	_ = exec.Command("templ", "generate").Run()
	fmt.Printf("✨ Proyecto inicializado exitosamente como '%s'\n", newMod)
	return nil
}

func runNewModule(args []string, in io.Reader) {
	posName, flagArgs := extractLeadingPositional(args)
	fsFlags := flag.NewFlagSet("new-module", flag.ExitOnError)
	nameFlag := fsFlags.String("name", "", "Module name in singular (e.g. proveedor, articulo)")
	schemaFlag := fsFlags.String("schema", "", "Database schema (default: sistema)")
	titleFlag := fsFlags.String("title", "", "Human readable title (default: auto capitalized)")
	pluralFlag := fsFlags.String("plural", "", "Plural name (default: auto)")
	noMountFlag := fsFlags.Bool("no-mount", false, "Do not auto-mount in cmd/main.go")
	_ = fsFlags.Parse(flagArgs)

	rawName := strings.TrimSpace(*nameFlag)
	if rawName == "" {
		rawName = posName
	}
	if rawName == "" && fsFlags.NArg() > 0 {
		rawName = strings.TrimSpace(fsFlags.Arg(0))
	}

	schemaRaw := strings.TrimSpace(*schemaFlag)
	titleRaw := strings.TrimSpace(*titleFlag)

	if rawName == "" {
		reader := bufio.NewReader(in)
		rawName = promptInput(reader, "📦 Nombre del módulo en singular (ej. proveedor, orden-compra)", "")
		if schemaRaw == "" {
			schemaRaw = promptInput(reader, "🗂️  Esquema de PostgreSQL (sistema o institucion)", "sistema")
		}
		if titleRaw == "" {
			titleRaw = promptInput(reader, "🏷️  Título en pantalla y menú (Enter para autogenerar)", "")
		}
	}
	if schemaRaw == "" {
		schemaRaw = "sistema"
	}

	executeScaffold(rawName, schemaRaw, titleRaw, strings.TrimSpace(*pluralFlag), *noMountFlag, false)
}

func runRemoveModule(args []string, in io.Reader) {
	posName, flagArgs := extractLeadingPositional(args)
	fsFlags := flag.NewFlagSet("remove-module", flag.ExitOnError)
	nameFlag := fsFlags.String("name", "", "Module name to remove")
	_ = fsFlags.Parse(flagArgs)

	rawName := strings.TrimSpace(*nameFlag)
	if rawName == "" {
		rawName = posName
	}
	if rawName == "" && fsFlags.NArg() > 0 {
		rawName = strings.TrimSpace(fsFlags.Arg(0))
	}
	if rawName == "" {
		reader := bufio.NewReader(in)
		rawName = promptInput(reader, "🧹 Nombre del módulo a eliminar", "")
	}

	executeScaffold(rawName, "sistema", "", "", false, true)
}

func runLegacyFlags(args []string) {
	ensureProjectRoot()
	fsFlags := flag.NewFlagSet("jrmp", flag.ExitOnError)
	nameFlag := fsFlags.String("name", "", "Module name in singular (e.g. proveedor, articulo)")
	schemaFlag := fsFlags.String("schema", "sistema", "Database schema (default: sistema)")
	titleFlag := fsFlags.String("title", "", "Human readable title (default: auto capitalized)")
	pluralFlag := fsFlags.String("plural", "", "Plural name (default: auto)")
	noMountFlag := fsFlags.Bool("no-mount", false, "Do not auto-mount in cmd/main.go")
	removeFlag := fsFlags.Bool("remove", false, "Remove an existing scaffolded module and its migrations")
	initProjectFlag := fsFlags.String("init-project", "", "Rename the Go module across go.mod, .go, and .templ files")
	_ = fsFlags.Parse(args)

	if strings.TrimSpace(*initProjectFlag) != "" {
		runInitProject([]string{"-module", strings.TrimSpace(*initProjectFlag)}, os.Stdin)
		return
	}
	executeScaffold(strings.TrimSpace(*nameFlag), strings.TrimSpace(*schemaFlag), strings.TrimSpace(*titleFlag), strings.TrimSpace(*pluralFlag), *noMountFlag, *removeFlag)
}

func executeScaffold(rawName, schemaRaw, titleRaw, pluralRaw string, noMount, isRemove bool) {
	summary, err := scaffoldModuleCore(rawName, schemaRaw, titleRaw, pluralRaw, noMount, isRemove)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(summary)
}

func scaffoldModuleCore(rawName, schemaRaw, titleRaw, pluralRaw string, noMount, isRemove bool) (string, error) {
	if strings.TrimSpace(rawName) == "" {
		return "", fmt.Errorf("el nombre del módulo es obligatorio (ej. proveedor)")
	}

	name := toSnakeCase(rawName)
	if name == "" || !regexp.MustCompile(`^[a-z][a-z0-9_]*$`).MatchString(name) {
		return "", fmt.Errorf("el nombre '%s' no produce un identificador válido para Go (resultado: '%s')", rawName, name)
	}

	moduleName := getModuleName("go.mod")

	if isRemove {
		if err := removeModule(".", name, moduleName); err != nil {
			return "", fmt.Errorf("eliminando modulo '%s': %w", name, err)
		}
		return fmt.Sprintf("🧹 Módulo '%s' eliminado exitosamente (carpeta, migraciones, cmd/main.go y PostgreSQL).", name), nil
	}

	schema := toSnakeCase(schemaRaw)
	if schema == "" {
		schema = "sistema"
	}

	plural := pluralRaw
	if plural == "" {
		plural = autoPlural(name)
	} else {
		plural = toSnakeCase(plural)
	}

	nameUpper := toPascalCase(name)
	pluralUpper := toPascalCase(plural)
	if titleRaw != "" {
		pluralUpper = titleRaw
	}

	prefix := fmt.Sprintf("/%s/%s", schema, plural)
	table := fmt.Sprintf("%s.%s", schema, name)
	migrNum := getNextMigrationNumber()

	cfg := Config{
		Module:      moduleName,
		Name:        name,
		NameUpper:   nameUpper,
		Plural:      plural,
		PluralUpper: pluralUpper,
		Schema:      schema,
		Prefix:      prefix,
		Table:       table,
		MigrNum:     migrNum,
	}

	if err := generateModuleFiles(cfg); err != nil {
		return "", fmt.Errorf("generando archivos: %w", err)
	}

	if err := generateMigrations(cfg); err != nil {
		return "", fmt.Errorf("generando migraciones: %w", err)
	}

	if !noMount {
		if err := mountInMain(cfg); err != nil {
			fmt.Printf("⚠️ No se pudo auto-montar en cmd/main.go: %v (podes agregarlo manualmente)\n", err)
		}
	}

	_ = exec.Command("templ", "generate").Run()
	_ = exec.Command("templ", "fmt", ".").Run()

	return fmt.Sprintf(
		"✨ Módulo '%s' generado exitosamente!\n📁 Carpeta: ./%s\n📜 Migración: ./sql/migrations/%s_create_%s_table.up.sql\n🌐 Endpoint: http://localhost:8080%s",
		cfg.NameUpper, cfg.Name, cfg.MigrNum, cfg.Name, cfg.Prefix,
	), nil
}

func runCheck() {
	if err := checkProjectCore(); err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ ¡Validación completa exitosa!")
}

func checkProjectCore() error {
	fmt.Println("🔍 [1/4] Verificando formato de archivos .templ...")
	if err := runCmdInDir(".", "templ", "fmt", "-fail", "."); err != nil {
		return fmt.Errorf("templ fmt falló: %w", err)
	}
	fmt.Println("⚙️  [2/4] Generando código Go desde .templ...")
	if err := runCmdInDir(".", "templ", "generate"); err != nil {
		return fmt.Errorf("templ generate falló: %w", err)
	}
	fmt.Println("🧹 [3/4] Ejecutando golangci-lint...")
	if err := runCmdInDir(".", "golangci-lint", "run", "./..."); err != nil {
		return fmt.Errorf("golangci-lint falló en raíz: %w", err)
	}
	if _, err := os.Stat(filepath.Join("cmd", "jrmp", "go.mod")); err == nil {
		if err := runCmdInDir(filepath.Join("cmd", "jrmp"), "golangci-lint", "run", "-c", "../../.golangci.yml", "./..."); err != nil {
			return fmt.Errorf("golangci-lint falló en cmd/jrmp: %w", err)
		}
	}
	fmt.Println("🧪 [4/4] Ejecutando pruebas unitarias...")
	if err := runCmdInDir(".", "go", "test", "./..."); err != nil {
		return fmt.Errorf("go test falló en raíz: %w", err)
	}
	if _, err := os.Stat(filepath.Join("cmd", "jrmp", "go.mod")); err == nil {
		if err := runCmdInDir(filepath.Join("cmd", "jrmp"), "go", "test", "./..."); err != nil {
			return fmt.Errorf("go test falló en cmd/jrmp: %w", err)
		}
	}
	return nil
}

func runDev() {
	if runtime.GOOS == "windows" {
		mustRunCmd(".", "air", "-c", ".air.toml", "--build.bin", `.\\tmp\\main.exe`, "--build.cmd", `go build -mod=vendor -o .\\tmp\\main.exe .\\cmd\\.`)
		return
	}
	mustRunCmd(".", "air", "-c", ".air.toml")
}

func runCmdInDir(dir, name string, args ...string) error {
	// #nosec G204,G702 -- CLI wrapper deliberately runs developer toolchain commands
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func mustRunCmd(dir, name string, args ...string) {
	if err := runCmdInDir(dir, name, args...); err != nil {
		fmt.Printf("❌ Error ejecutando '%s %s': %v\n", name, strings.Join(args, " "), err)
		os.Exit(1)
	}
}

func ensureEnvFile(rootDir, serviceName, dbUser, dbPass string) error {
	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	if _, err := root.Stat(".env"); err == nil {
		return nil // .env ya existe; respetamos la configuración actual
	}

	exampleBytes, err := root.ReadFile(".env.example")
	if err != nil {
		return err
	}

	if serviceName == "" {
		serviceName = "plantilla"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPass == "" {
		dbPass = "postgres"
	}

	var lines []string
	for _, line := range strings.Split(string(exampleBytes), "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "SERVICIO="):
			lines = append(lines, "SERVICIO="+serviceName)
		case strings.HasPrefix(trimmed, "PQ_WRITE_DB="):
			lines = append(lines, "PQ_WRITE_DB="+serviceName)
		case strings.HasPrefix(trimmed, "PQ_READ_DB="):
			lines = append(lines, "PQ_READ_DB="+serviceName)
		case strings.HasPrefix(trimmed, "PQ_WRITE_USER="):
			lines = append(lines, "PQ_WRITE_USER="+dbUser)
		case strings.HasPrefix(trimmed, "PQ_READ_USER="):
			lines = append(lines, "PQ_READ_USER="+dbUser)
		case strings.HasPrefix(trimmed, "PQ_WRITE_PASSWORD="):
			lines = append(lines, "PQ_WRITE_PASSWORD="+dbPass)
		case strings.HasPrefix(trimmed, "PQ_READ_PASSWORD="):
			lines = append(lines, "PQ_READ_PASSWORD="+dbPass)
		case strings.HasPrefix(trimmed, "COOKIE_NOMBRE="):
			lines = append(lines, "COOKIE_NOMBRE="+serviceName+"_v2")
		default:
			lines = append(lines, line)
		}
	}

	return root.WriteFile(".env", []byte(strings.Join(lines, "\n")), 0600)
}

func autoPlural(s string) string {
	last := s[len(s)-1:]
	if strings.ContainsAny(last, "aeiou") {
		return s + "s"
	}
	return s + "es"
}

func toSnakeCase(s string) string {
	s = strings.TrimSpace(s)
	re := regexp.MustCompile(`[\-\s]+`)
	s = re.ReplaceAllString(s, "_")
	reCamel := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	s = reCamel.ReplaceAllString(s, "${1}_${2}")
	reInvalid := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	s = reInvalid.ReplaceAllString(s, "")
	return strings.ToLower(s)
}

func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, "")
}

func getNextMigrationNumber() string {
	files, err := filepath.Glob("sql/migrations/*.sql")
	if err != nil || len(files) == 0 {
		return "000001"
	}

	re := regexp.MustCompile(`^(\d+)`)
	maxNum := 0

	for _, f := range files {
		base := filepath.Base(f)
		matches := re.FindStringSubmatch(base)
		if len(matches) > 1 {
			if n, err := strconv.Atoi(matches[1]); err == nil && n > maxNum {
				maxNum = n
			}
		}
	}

	return fmt.Sprintf("%06d", maxNum+1)
}

func generateModuleFiles(cfg Config) error {
	dirs := []string{
		cfg.Name,
		filepath.Join(cfg.Name, "repository"),
		filepath.Join(cfg.Name, "service"),
		filepath.Join(cfg.Name, "handler"),
		filepath.Join(cfg.Name, "view", "components"),
		filepath.Join(cfg.Name, "view", "layouts"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		filepath.Join(cfg.Name, "types.go"):                                   tmplTypes,
		filepath.Join(cfg.Name, "rutas.go"):                                   tmplRutas,
		filepath.Join(cfg.Name, "repository", "types.go"):                     tmplRepoTypes,
		filepath.Join(cfg.Name, "repository", "queries.go"):                   tmplRepoQueries,
		filepath.Join(cfg.Name, "repository", "repository.go"):                tmplRepoMain,
		filepath.Join(cfg.Name, "service", "types.go"):                        tmplServiceTypes,
		filepath.Join(cfg.Name, "service", "service.go"):                      tmplServiceMain,
		filepath.Join(cfg.Name, "service", "service_test.go"):                 tmplServiceTest,
		filepath.Join(cfg.Name, "service", "reportes.go"):                     tmplServiceReports,
		filepath.Join(cfg.Name, "handler", "types.go"):                        tmplHandlerTypes,
		filepath.Join(cfg.Name, "handler", "register.go"):                     tmplHandlerRegister,
		filepath.Join(cfg.Name, "handler", "crud.go"):                         tmplHandlerCRUD,
		filepath.Join(cfg.Name, "handler", "views.go"):                        tmplHandlerViews,
		filepath.Join(cfg.Name, "handler", "files.go"):                        tmplHandlerFiles,
		filepath.Join(cfg.Name, "view", "components", "table_v2.templ"):       tmplViewTable,
		filepath.Join(cfg.Name, "view", "components", "form_v2.templ"):        tmplViewForm,
		filepath.Join(cfg.Name, "view", "layouts", cfg.Name+"_v2.templ"):      tmplViewLayout,
	}

	for path, tmplStr := range files {
		t, err := template.New(path).Delims("<<", ">>").Parse(tmplStr)
		if err != nil {
			return fmt.Errorf("error parseando plantilla para %s: %w", path, err)
		}
		f, err := os.Create(filepath.Clean(path))
		if err != nil {
			return fmt.Errorf("error creando archivo %s: %w", path, err)
		}
		if err := t.Execute(f, cfg); err != nil {
			_ = f.Close()
			return fmt.Errorf("error ejecutando plantilla %s: %w", path, err)
		}
		_ = f.Close()
	}

	return nil
}

func generateMigrations(cfg Config) error {
	upPath := filepath.Clean(filepath.Join("sql", "migrations", fmt.Sprintf("%s_create_%s_table.up.sql", cfg.MigrNum, cfg.Name)))
	downPath := filepath.Clean(filepath.Join("sql", "migrations", fmt.Sprintf("%s_create_%s_table.down.sql", cfg.MigrNum, cfg.Name)))

	upTmpl, err := template.New("up").Delims("<<", ">>").Parse(tmplMigrationUp)
	if err != nil {
		return err
	}
	downTmpl, err := template.New("down").Delims("<<", ">>").Parse(tmplMigrationDown)
	if err != nil {
		return err
	}

	fUp, err := os.Create(upPath)
	if err != nil {
		return err
	}
	defer fUp.Close()
	if err := upTmpl.Execute(fUp, cfg); err != nil {
		return err
	}

	fDown, err := os.Create(downPath)
	if err != nil {
		return err
	}
	defer fDown.Close()
	return downTmpl.Execute(fDown, cfg)
}

const mainFile = "cmd/main.go"

func getModuleName(goModPath string) string {
	// #nosec G304 -- goModPath is controlled by internal callers
	data, err := os.ReadFile(filepath.Clean(goModPath))
	if err != nil {
		return "PlantillaInstitucional"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			if mod != "" {
				return mod
			}
		}
	}
	return "PlantillaInstitucional"
}

func mountInMain(cfg Config) error {
	return mountInMainFile(mainFile, cfg)
}

func mountInMainFile(targetFile string, cfg Config) error {
	// #nosec G304 -- targetFile is a controlled path to cmd/main.go
	content, err := os.ReadFile(filepath.Clean(targetFile))
	if err != nil {
		return err
	}
	s := string(content)

	mod := cfg.Module
	if mod == "" {
		mod = getModuleName("go.mod")
	}

	importTarget := fmt.Sprintf("%s_handler \"%s/%s/handler\"", cfg.Name, mod, cfg.Name)
	if strings.Contains(s, importTarget) {
		return nil // Ya montado
	}

	// Inyectar import
	importMarker := "import ("
	idx := strings.Index(s, importMarker)
	if idx == -1 {
		return fmt.Errorf("no se encontro bloque import en %s", targetFile)
	}
	insertImportPos := idx + len(importMarker)
	s = s[:insertImportPos] + "\n\t" + importTarget + s[insertImportPos:]

	// Inyectar llamada de Register en el grupo protegido respetando la indentación
	registerCall := fmt.Sprintf("%s_handler.Register(a, r)", cfg.Name)
	groupMarker := "miperfil_handler.Register(a, r)"
	if strings.Contains(s, groupMarker) {
		indent := "\t\t\t"
		for _, line := range strings.Split(s, "\n") {
			if strings.Contains(line, groupMarker) {
				indent = line[:strings.Index(line, groupMarker)]
				break
			}
		}
		s = strings.Replace(s, groupMarker, groupMarker+"\n"+indent+registerCall, 1)
	} else {
		navMarker := "r.Use(app_middleware.NavigationMiddleware(a))"
		s = strings.Replace(s, navMarker, navMarker+"\n\n\t\t\t"+registerCall, 1)
	}

	// #nosec G703 -- targetFile is validated and cleaned
	return os.WriteFile(filepath.Clean(targetFile), []byte(s), 0600)
}

func unmountFromMainFile(targetFile, name, moduleName string) error {
	// #nosec G304 -- targetFile is a controlled path to cmd/main.go
	content, err := os.ReadFile(filepath.Clean(targetFile))
	if err != nil {
		return err
	}
	if moduleName == "" {
		moduleName = getModuleName("go.mod")
	}

	importTarget := fmt.Sprintf("%s_handler \"%s/%s/handler\"", name, moduleName, name)
	registerCall := fmt.Sprintf("%s_handler.Register(a, r)", name)

	var kept []string
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == importTarget || trimmed == registerCall {
			continue
		}
		kept = append(kept, line)
	}

	// #nosec G703 -- targetFile is validated and cleaned
	return os.WriteFile(filepath.Clean(targetFile), []byte(strings.Join(kept, "\n")), 0600)
}

func removeModule(rootDir, name, moduleName string) error {
	protected := map[string]bool{
		"app": true, "cmd": true, "configuracion": true, "crud": true,
		"dashboard": true, "docs": true, "institucion": true, "login": true,
		"miperfil": true, "pagination": true, "parametros": true, "shared": true,
		"sql": true, "static": true, "vendor": true,
	}
	if protected[name] {
		return fmt.Errorf("el directorio '%s' pertenece al nucleo de la plantilla y no puede eliminarse", name)
	}

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	if err := root.RemoveAll(name); err != nil {
		return fmt.Errorf("error eliminando carpeta %s: %w", name, err)
	}

	pattern := filepath.Join(rootDir, "sql", "migrations", fmt.Sprintf("*_create_%s_table.*.sql", name))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	var downScripts []string
	for _, m := range matches {
		relMigr := filepath.Join("sql", "migrations", filepath.Base(m))
		if strings.HasSuffix(m, ".down.sql") {
			if b, readErr := root.ReadFile(relMigr); readErr == nil {
				downScripts = append(downScripts, string(b))
			}
		}
		if err := root.Remove(relMigr); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error eliminando migracion %s: %w", relMigr, err)
		}
	}

	maxRemaining := getMaxMigrationNumberInDir(filepath.Join(rootDir, "sql", "migrations"))
	revertModuleDatabase(rootDir, downScripts, maxRemaining)

	mainPath := filepath.Join(rootDir, mainFile)
	if _, err := os.Stat(mainPath); err == nil {
		if err := unmountFromMainFile(mainPath, name, moduleName); err != nil {
			return err
		}
	}
	return nil
}

func getMaxMigrationNumberInDir(migrationsDir string) int {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return 0
	}
	re := regexp.MustCompile(`^(\d+)`)
	maxNum := 0
	for _, f := range files {
		matches := re.FindStringSubmatch(filepath.Base(f))
		if len(matches) > 1 {
			if n, err := strconv.Atoi(matches[1]); err == nil && n > maxNum {
				maxNum = n
			}
		}
	}
	return maxNum
}

func revertModuleDatabase(rootDir string, downScripts []string, maxRemainingVersion int) {
	if len(downScripts) == 0 && maxRemainingVersion <= 0 {
		return
	}

	var sb strings.Builder
	for _, script := range downScripts {
		sb.WriteString(script)
		sb.WriteString("\n")
	}
	if maxRemainingVersion > 0 {
		fmt.Fprintf(&sb, "UPDATE public.schema_migrations SET version = %d, dirty = false WHERE version > %d;\n", maxRemainingVersion, maxRemainingVersion)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"docker", "compose", "-f", "docker-compose.local.yml",
		"exec", "-T", "db",
		"sh", "-c", `psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1`,
	)
	cmd.Dir = rootDir
	cmd.Stdin = strings.NewReader(sb.String())

	if err := cmd.Run(); err != nil {
		fmt.Println("ℹ️ Contenedor de base de datos no activo; se omitio limpieza SQL en vivo.")
		return
	}
	fmt.Println("🗄️ Tabla, politicas ABAC y schema_migrations revertidos en PostgreSQL.")
}

func renameProjectModule(rootDir, newModule string) error {
	newModule = strings.TrimSpace(newModule)
	if newModule == "" || strings.ContainsAny(newModule, " \t\n\"'") {
		return fmt.Errorf("nombre de modulo invalido: %q", newModule)
	}

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	goModBytes, err := root.ReadFile("go.mod")
	if err != nil {
		return err
	}
	oldModule := "PlantillaInstitucional"
	for _, line := range strings.Split(string(goModBytes), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			if mod := strings.TrimSpace(strings.TrimPrefix(line, "module ")); mod != "" {
				oldModule = mod
			}
			break
		}
	}
	if oldModule == "" || oldModule == newModule {
		return nil
	}

	updatedGoMod := strings.Replace(string(goModBytes), "module "+oldModule, "module "+newModule, 1)
	if err := root.WriteFile("go.mod", []byte(updatedGoMod), 0600); err != nil {
		return err
	}

	oldPrefix := oldModule + "/"
	newPrefix := newModule + "/"

	return fs.WalkDir(root.FS(), ".", func(relPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "tmp", ".codegraph":
				return fs.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(d.Name())
		if ext != ".go" && ext != ".templ" {
			return nil
		}
		b, err := root.ReadFile(relPath)
		if err != nil {
			return err
		}
		content := string(b)
		if !strings.Contains(content, oldPrefix) {
			return nil
		}
		updated := strings.ReplaceAll(content, oldPrefix, newPrefix)
		return root.WriteFile(relPath, []byte(updated), 0600)
	})
}

// =========================================================================
// PLANTILLAS DE CÓDIGO
// =========================================================================

const tmplTypes = `package <<.Name>>

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type <<.NameUpper>> struct {
	ID          uuid.UUID ` + "`db:\"id\" json:\"id\"`" + `
	Nombre      string    ` + "`db:\"nombre\" json:\"nombre\" validate:\"required,min=2,max=150\"`" + `
	Codigo      string    ` + "`db:\"codigo\" json:\"codigo\" validate:\"required,min=2,max=50\"`" + `
	Descripcion string    ` + "`db:\"descripcion\" json:\"descripcion\"`" + `
	IsActive    bool      ` + "`db:\"is_active\" json:\"is_active\"`" + `
	FechaI      time.Time ` + "`db:\"fecha_i\" json:\"fecha_i\"`" + `
	FechaM      time.Time ` + "`db:\"fecha_m\" json:\"fecha_m\"`" + `
}
`

const tmplRutas = `package <<.Name>>

import "<<.Module>>/shared/pagina"

// Rutas contiene la configuracion canonica de rutas para el modulo <<.NameUpper>>.
var Rutas = pagina.NuevaRuta("<<.Prefix>>")
`

const tmplRepoTypes = `package repository

import (
	bd "<<.Module>>/app/db"
	"<<.Module>>/<<.Name>>"
	"time"

	"github.com/gofrs/uuid/v5"
)

var (
	colsTexto     = map[string]bool{"nombre": true, "codigo": true, "descripcion": true}
	colsNumericas = map[string]bool{}
	colsUUID      = map[string]bool{"id": true}
	colsOrden     = map[string]bool{"id": true, "nombre": true, "codigo": true, "fecha_i": true, "is_active": true}
)

type Row<<.NameUpper>> struct {
	ID          uuid.UUID ` + "`db:\"id\"`" + `
	Nombre      string    ` + "`db:\"nombre\"`" + `
	Codigo      string    ` + "`db:\"codigo\"`" + `
	Descripcion string    ` + "`db:\"descripcion\"`" + `
	IsActive    bool      ` + "`db:\"is_active\"`" + `
	FechaI      time.Time ` + "`db:\"fecha_i\"`" + `
	FechaM      time.Time ` + "`db:\"fecha_m\"`" + `
}

func (r Row<<.NameUpper>>) ToEntity() <<.Name>>.<<.NameUpper>> {
	return <<.Name>>.<<.NameUpper>>{
		ID:          r.ID,
		Nombre:      r.Nombre,
		Codigo:      r.Codigo,
		Descripcion: r.Descripcion,
		IsActive:    r.IsActive,
		FechaI:      r.FechaI,
		FechaM:      r.FechaM,
	}
}

type <<.NameUpper>>Repository struct {
	db *bd.AppPools
}

func New<<.NameUpper>>Repository(db *bd.AppPools) *<<.NameUpper>>Repository {
	return &<<.NameUpper>>Repository{db: db}
}
`

const tmplRepoQueries = `package repository

const (
	Select<<.NameUpper>>Filtered = ` + "`" + `
		SELECT id, nombre, codigo, COALESCE(descripcion, '') as descripcion, is_active, fecha_i, COALESCE(fecha_m, fecha_i) as fecha_m, COUNT(*) OVER() AS total_count
		FROM <<.Table>>
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	` + "`" + `

	Select<<.NameUpper>>ByID = ` + "`" + `
		SELECT id, nombre, codigo, COALESCE(descripcion, '') as descripcion, is_active, fecha_i, COALESCE(fecha_m, fecha_i) as fecha_m
		FROM <<.Table>>
		WHERE id = $1
	` + "`" + `

	SelectAll<<.NameUpper>> = ` + "`" + `
		SELECT id, nombre, codigo, COALESCE(descripcion, '') as descripcion, is_active, fecha_i, COALESCE(fecha_m, fecha_i) as fecha_m
		FROM <<.Table>>
		ORDER BY fecha_i DESC
	` + "`" + `

	Insert<<.NameUpper>> = ` + "`" + `
		INSERT INTO <<.Table>> (nombre, codigo, descripcion, is_active, tenant_id, razon_i)
		VALUES ($1, $2, $3, $4, CASE WHEN current_setting('app.tenant_id', true) ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN current_setting('app.tenant_id', true)::uuid ELSE NULL END, 'Creacion')
		RETURNING id
	` + "`" + `

	Update<<.NameUpper>> = ` + "`" + `
		UPDATE <<.Table>>
		SET nombre = $1, codigo = $2, descripcion = $3, is_active = $4, razon_m = 'Actualizacion'
		WHERE id = $5
	` + "`" + `

	Delete<<.NameUpper>> = ` + "`" + `
		DELETE FROM <<.Table>>
		WHERE id = $1
	` + "`" + `
)
`

const tmplRepoMain = `package repository

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/pagination"
	shared "<<.Module>>/shared/db"
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *<<.NameUpper>>Repository) Consultar(ctx context.Context, criterios pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]) ([]<<.Name>>.<<.NameUpper>>, int, error) {
	config := shared.FilterConfig{
		Texto:     colsTexto,
		Numericas: colsNumericas,
		UUID:      colsUUID,
	}
	filtros, args := shared.BuildDynamicWhere(criterios.FilteredBy, config, "<<.Name>>")
	whereClause := ""
	if len(filtros) > 0 {
		whereClause = " WHERE " + strings.Join(filtros, " AND ")
	}

	orden, err := shared.BuildOrderBy(criterios.SortedBy, criterios.Direction, colsOrden, "", "nombre ASC")
	if err != nil {
		return nil, 0, err
	}

	queryArgs := append(args, criterios.Size, criterios.First)
	query := fmt.Sprintf(Select<<.NameUpper>>Filtered, whereClause, orden, len(args)+1, len(args)+2)

	var rowsList []Row<<.NameUpper>>
	var total int

	err = r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, queryArgs...)
		if err != nil {
			return fmt.Errorf("error ejecutando consulta: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var row Row<<.NameUpper>>
			var count int
			if err := rows.Scan(&row.ID, &row.Nombre, &row.Codigo, &row.Descripcion, &row.IsActive, &row.FechaI, &row.FechaM, &count); err != nil {
				return err
			}
			rowsList = append(rowsList, row)
			total = count
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, err
	}

	entities := make([]<<.Name>>.<<.NameUpper>>, len(rowsList))
	for i, row := range rowsList {
		entities[i] = row.ToEntity()
	}
	return entities, total, nil
}

func (r *<<.NameUpper>>Repository) ObtenerPorID(ctx context.Context, id uuid.UUID) (<<.Name>>.<<.NameUpper>>, error) {
	var row Row<<.NameUpper>>
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, Select<<.NameUpper>>ByID, id)
		if err != nil {
			return err
		}
		row, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Row<<.NameUpper>>])
		return err
	})
	if err != nil {
		return <<.Name>>.<<.NameUpper>>{}, err
	}
	return row.ToEntity(), nil
}

func (r *<<.NameUpper>>Repository) ObtenerTodos(ctx context.Context) ([]<<.Name>>.<<.NameUpper>>, error) {
	var items []<<.Name>>.<<.NameUpper>>
	err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, SelectAll<<.NameUpper>>)
		if err != nil {
			return fmt.Errorf("error ejecutando consulta: %w", err)
		}

		rowsList, err := pgx.CollectRows(rows, pgx.RowToStructByName[Row<<.NameUpper>>])
		if err != nil {
			return fmt.Errorf("error mapeando resultados: %w", err)
		}

		items = make([]<<.Name>>.<<.NameUpper>>, len(rowsList))
		for i, row := range rowsList {
			items[i] = row.ToEntity()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *<<.NameUpper>>Repository) Crear(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		var id uuid.UUID
		return tx.QueryRow(ctx, Insert<<.NameUpper>>, item.Nombre, item.Codigo, item.Descripcion, item.IsActive).Scan(&id)
	})
}

func (r *<<.NameUpper>>Repository) Actualizar(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, Update<<.NameUpper>>, item.Nombre, item.Codigo, item.Descripcion, item.IsActive, item.ID)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("registro no encontrado para actualizar")
		}
		return nil
	})
}

func (r *<<.NameUpper>>Repository) Eliminar(ctx context.Context, id uuid.UUID) error {
	return r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx, Delete<<.NameUpper>>, id)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("registro no encontrado para eliminar")
		}
		return nil
	})
}
`

const tmplServiceTypes = `package service

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/pagination"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

type <<.NameUpper>>RepositoryProvider interface {
	Consultar(ctx context.Context, criterios pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]) ([]<<.Name>>.<<.NameUpper>>, int, error)
	ObtenerPorID(ctx context.Context, id uuid.UUID) (<<.Name>>.<<.NameUpper>>, error)
	ObtenerTodos(ctx context.Context) ([]<<.Name>>.<<.NameUpper>>, error)
	Crear(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error
	Actualizar(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error
	Eliminar(ctx context.Context, id uuid.UUID) error
}

type <<.NameUpper>>ServiceProvider interface {
	Listar(ctx context.Context, criterios pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]) (pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>], error)
	Obtener(ctx context.Context, id uuid.UUID) (<<.Name>>.<<.NameUpper>>, error)
	Crear(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error
	Actualizar(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error
	Eliminar(ctx context.Context, id uuid.UUID) error
	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
}

type <<.NameUpper>>Service struct {
	repo     <<.NameUpper>>RepositoryProvider
	validate *validator.Validate
	logger   *logs.Logger
	archivos *archivos.ManejadorArchivos
}

func New<<.NameUpper>>Service(repo <<.NameUpper>>RepositoryProvider, validate *validator.Validate, logger *logs.Logger, archivos *archivos.ManejadorArchivos) *<<.NameUpper>>Service {
	return &<<.NameUpper>>Service{
		repo:     repo,
		validate: validate,
		logger:   logger,
		archivos: archivos,
	}
}
`

const tmplServiceMain = `package service

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/pagination"
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

func (s *<<.NameUpper>>Service) Listar(ctx context.Context, criterios pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]) (pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>], error) {
	data, total, err := s.repo.Consultar(ctx, criterios)
	if err != nil {
		s.logger.ErrorContext(ctx, "Error listando <<.Plural>>", err)
		return criterios, fmt.Errorf("error al consultar <<.Plural>>: %w", err)
	}

	criterios.Data = data
	criterios.Count = total
	return criterios, nil
}

func (s *<<.NameUpper>>Service) Obtener(ctx context.Context, id uuid.UUID) (<<.Name>>.<<.NameUpper>>, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *<<.NameUpper>>Service) Crear(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error {
	if err := s.validate.Struct(item); err != nil {
		return err
	}
	return s.repo.Crear(ctx, item)
}

func (s *<<.NameUpper>>Service) Actualizar(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error {
	if err := s.validate.Struct(item); err != nil {
		return err
	}
	return s.repo.Actualizar(ctx, item)
}

func (s *<<.NameUpper>>Service) Eliminar(ctx context.Context, id uuid.UUID) error {
	return s.repo.Eliminar(ctx, id)
}
`

const tmplServiceTest = `package service

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/pagination"
	"context"
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
)

type mock<<.NameUpper>>Repo struct {
	consultarFn    func(ctx context.Context, criterios pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]) ([]<<.Name>>.<<.NameUpper>>, int, error)
	obtenerPorIDFn func(ctx context.Context, id uuid.UUID) (<<.Name>>.<<.NameUpper>>, error)
	obtenerTodosFn func(ctx context.Context) ([]<<.Name>>.<<.NameUpper>>, error)
	crearFn        func(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error
	actualizarFn   func(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error
	eliminarFn     func(ctx context.Context, id uuid.UUID) error
}

func (m *mock<<.NameUpper>>Repo) Consultar(ctx context.Context, criterios pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]) ([]<<.Name>>.<<.NameUpper>>, int, error) {
	if m.consultarFn != nil {
		return m.consultarFn(ctx, criterios)
	}
	return nil, 0, nil
}

func (m *mock<<.NameUpper>>Repo) ObtenerPorID(ctx context.Context, id uuid.UUID) (<<.Name>>.<<.NameUpper>>, error) {
	if m.obtenerPorIDFn != nil {
		return m.obtenerPorIDFn(ctx, id)
	}
	return <<.Name>>.<<.NameUpper>>{}, nil
}

func (m *mock<<.NameUpper>>Repo) ObtenerTodos(ctx context.Context) ([]<<.Name>>.<<.NameUpper>>, error) {
	if m.obtenerTodosFn != nil {
		return m.obtenerTodosFn(ctx)
	}
	return nil, nil
}

func (m *mock<<.NameUpper>>Repo) Crear(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error {
	if m.crearFn != nil {
		return m.crearFn(ctx, item)
	}
	return nil
}

func (m *mock<<.NameUpper>>Repo) Actualizar(ctx context.Context, item <<.Name>>.<<.NameUpper>>) error {
	if m.actualizarFn != nil {
		return m.actualizarFn(ctx, item)
	}
	return nil
}

func (m *mock<<.NameUpper>>Repo) Eliminar(ctx context.Context, id uuid.UUID) error {
	if m.eliminarFn != nil {
		return m.eliminarFn(ctx, id)
	}
	return nil
}

func Test<<.NameUpper>>Service_Crear(t *testing.T) {
	tests := []struct {
		name    string
		item    <<.Name>>.<<.NameUpper>>
		repoErr error
		wantErr bool
	}{
		{
			name:    "creates valid record",
			item:    <<.Name>>.<<.NameUpper>>{Nombre: "Registro Uno", Codigo: "REG-01", IsActive: true},
			wantErr: false,
		},
		{
			name:    "rejects record with empty nombre",
			item:    <<.Name>>.<<.NameUpper>>{Nombre: "", Codigo: "REG-01"},
			wantErr: true,
		},
		{
			name:    "propagates repository error",
			item:    <<.Name>>.<<.NameUpper>>{Nombre: "Registro Dos", Codigo: "REG-02"},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock<<.NameUpper>>Repo{
				crearFn: func(context.Context, <<.Name>>.<<.NameUpper>>) error {
					return tt.repoErr
				},
			}
			svc := New<<.NameUpper>>Service(repo, validator.New(), nil, nil)
			err := svc.Crear(context.Background(), tt.item)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Crear() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
`

const tmplServiceReports = `package service

import (
	"<<.Module>>/<<.Name>>"
	reporte "<<.Module>>/shared/reportes"
	"context"
	"fmt"
)

var (
	<<.Name>>ReportHeaders = []string{"Nombre", "Código", "Descripción", "Estatus"}
	<<.Name>>ReportFields  = []string{"Nombre", "Codigo", "Descripcion", "IsActive"}
)

func (s *<<.NameUpper>>Service) ReportePdf(ctx context.Context) (string, error) {
	data, err := s.repo.ObtenerTodos(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo obteniendo registros para PDF <<.Name>>", err, "op", "ReportePdf")
		return "", fmt.Errorf("error obteniendo registros: %w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("no hay registros disponibles para exportar")
	}

	renderFuncs := map[string]func(any) string{
		"IsActive": func(v any) string {
			if item, ok := v.(<<.Name>>.<<.NameUpper>>); ok && item.IsActive {
				return "Activo"
			}
			if item, ok := v.(*<<.Name>>.<<.NameUpper>>); ok && item != nil && item.IsActive {
				return "Activo"
			}
			return "Inactivo"
		},
	}

	url, err := reporte.GenerarReportePDF(
		ctx,
		s.archivos,
		data,
		<<.Name>>ReportHeaders,
		<<.Name>>ReportFields,
		"reporte_<<.Name>>",
		"Reporte de <<.PluralUpper>>",
		renderFuncs,
		true,
	)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación PDF <<.Name>>", err, "op", "ReportePdf")
		return "", fmt.Errorf("error generando PDF: %w", err)
	}

	return url, nil
}

func (s *<<.NameUpper>>Service) ReporteExcel(ctx context.Context) (string, error) {
	data, err := s.repo.ObtenerTodos(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo obteniendo registros para Excel <<.Name>>", err, "op", "ReporteExcel")
		return "", fmt.Errorf("error obteniendo registros: %w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("no hay registros disponibles para exportar")
	}

	renderFuncs := map[string]func(any) string{
		"IsActive": func(v any) string {
			if item, ok := v.(<<.Name>>.<<.NameUpper>>); ok && item.IsActive {
				return "Activo"
			}
			if item, ok := v.(*<<.Name>>.<<.NameUpper>>); ok && item != nil && item.IsActive {
				return "Activo"
			}
			return "Inactivo"
		},
	}

	url, err := reporte.GenerarReporteExcel(
		ctx,
		s.archivos,
		data,
		<<.Name>>ReportHeaders,
		<<.Name>>ReportFields,
		"reporte_<<.Name>>",
		"Reporte de <<.PluralUpper>>",
		renderFuncs,
		true,
	)
	if err != nil {
		s.logger.ErrorContext(ctx, "Fallo generación Excel <<.Name>>", err, "op", "ReporteExcel")
		return "", fmt.Errorf("error generando Excel: %w", err)
	}

	return url, nil
}
`

const tmplHandlerTypes = `package handler

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/<<.Name>>/service"
	"<<.Module>>/shared/errores"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/issemym-repos/logs"
)

type <<.NameUpper>>Handler struct {
	svc             service.<<.NameUpper>>ServiceProvider
	logger          *logs.Logger
	validate        *validator.Validate
	sqlErrorHandler *errores.SQLErrorHandler
}

func New<<.NameUpper>>Handler(svc service.<<.NameUpper>>ServiceProvider, logger *logs.Logger, validate *validator.Validate) *<<.NameUpper>>Handler {
	return &<<.NameUpper>>Handler{
		svc:             svc,
		logger:          logger,
		validate:        validate,
		sqlErrorHandler: errores.NewSQLErrorHandler(logger),
	}
}

func (h *<<.NameUpper>>Handler) RegisterRoutes(r chi.Router) {
	rutas := <<.Name>>.Rutas
	r.Get(rutas.Prefix, h.Index)
	r.Route(rutas.Prefix, func(r chi.Router) {
		r.Get("/", h.Index)
		r.Get("/tabla", h.SearchTable)
		r.Get("/modales/nuevo", h.ModalCreate)
		r.Post("/", h.Create)
		r.Get("/modales/editar/{id}", h.ModalUpdate)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Route("/reportes", func(r chi.Router) {
			r.Get("/pdf", h.DownloadPDF)
			r.Get("/excel", h.DownloadExcel)
			r.Get("/notificacion-exito", h.SuccessFileMessage)
		})
	})
}

func (h *<<.NameUpper>>Handler) parseForm(r *http.Request) (<<.Name>>.<<.NameUpper>>, error) {
	if err := r.ParseForm(); err != nil {
		return <<.Name>>.<<.NameUpper>>{}, fmt.Errorf("error parseando formulario: %w", err)
	}

	return <<.Name>>.<<.NameUpper>>{
		Nombre:      r.FormValue("nombre"),
		Codigo:      r.FormValue("codigo"),
		Descripcion: r.FormValue("descripcion"),
		IsActive:    r.FormValue("activo") == "true" || r.FormValue("activo") == "on" || r.FormValue("activo") == "1",
	}, nil
}
`

const tmplHandlerRegister = `package handler

import (
	"<<.Module>>/app"
	"<<.Module>>/shared/navigation"
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/<<.Name>>/repository"
	"<<.Module>>/<<.Name>>/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el repositorio, servicio, handler y monta las rutas de <<.NameUpper>>.
func Register(a *app.App, r chi.Router) {
	validate := validator.New()
	repo := repository.New<<.NameUpper>>Repository(a.Pools)
	svc := service.New<<.NameUpper>>Service(repo, validate, a.Logger, a.ManejadorArchivos)
	h := New<<.NameUpper>>Handler(svc, a.Logger, validate)
	h.RegisterRoutes(r)

	navigation.RegistrarItemMenu(navigation.ItemMenu{
		Label:  "<<.PluralUpper>>",
		URL:    <<.Name>>.Rutas.Prefix,
		Icon:   "folder",
		Codigo: "<<.Name>>",
	})
}
`

const tmplHandlerCRUD = `package handler

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/shared/errores"
	"<<.Module>>/shared/validacion"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
)

func (h *<<.NameUpper>>Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	item, err := h.parseForm(r)
	if err != nil {
		h.MandarErrorModalCreate(ctx, w, "Datos invalidos", nil, item)
		return
	}

	if err := h.svc.Crear(ctx, item); err != nil {
		valErrs := validacion.ValidationErrorsMap(err)
		h.MandarErrorModalCreate(ctx, w, "Error al guardar registro", valErrs, item)
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroCreado, <<.Name>>.Rutas.Tabla)
}

func (h *<<.NameUpper>>Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		h.MandarErrorGeneric(w, r, "ID invalido")
		return
	}

	item, err := h.parseForm(r)
	if err != nil {
		h.MandarErrorModalUpdate(ctx, w, "Datos invalidos", nil, &item)
		return
	}
	item.ID = id

	if err := h.svc.Actualizar(ctx, item); err != nil {
		valErrs := validacion.ValidationErrorsMap(err)
		h.MandarErrorModalUpdate(ctx, w, "Error al actualizar registro", valErrs, &item)
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroActualizado, <<.Name>>.Rutas.Tabla)
}

func (h *<<.NameUpper>>Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		h.MandarErrorGeneric(w, r, "ID invalido")
		return
	}

	if err := h.svc.Eliminar(ctx, id); err != nil {
		h.MandarErrorGeneric(w, r, "Error al eliminar registro")
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroEliminado, <<.Name>>.Rutas.Tabla)
}
`

const tmplHandlerViews = `package handler

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/<<.Name>>/view/components"
	"<<.Module>>/<<.Name>>/view/layouts"
	"<<.Module>>/pagination"
	"<<.Module>>/shared/utils"
	"<<.Module>>/shared/views/components/compartidos"
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
)

func (h *<<.NameUpper>>Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.FromQuery[<<.Name>>.<<.NameUpper>>](r.URL.Query())
	data, err := h.svc.Listar(ctx, criterios)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error listando <<.Plural>>", err)
	}

	if r.Header.Get("HX-Request") == "true" {
		utils.RenderIgnore(ctx, w, components.TableV2(data, <<.Name>>.Rutas.Tabla, false))
		return
	}
	utils.RenderIgnore(ctx, w, layouts.View<<.NameUpper>>V2("<<.PluralUpper>>", data, <<.Name>>.Rutas.Tabla))
}

func (h *<<.NameUpper>>Handler) SearchTable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	criterios := pagination.FromQuery[<<.Name>>.<<.NameUpper>>](r.URL.Query())
	data, err := h.svc.Listar(ctx, criterios)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error buscando en tabla <<.Plural>>", err)
	}
	utils.RenderIgnore(ctx, w, components.TableV2(data, <<.Name>>.Rutas.Tabla, false))
}

func (h *<<.NameUpper>>Handler) ModalCreate(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, components.FormCreate<<.NameUpper>>V2(<<.Name>>.<<.NameUpper>>{IsActive: true}, nil))
}

func (h *<<.NameUpper>>Handler) ModalUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "ID invalido", http.StatusBadRequest)
		return
	}

	item, err := h.svc.Obtener(ctx, id)
	if err != nil {
		http.Error(w, "No encontrado", http.StatusNotFound)
		return
	}

	utils.RenderIgnore(ctx, w, components.FormUpdate<<.NameUpper>>V2(&item, nil))
}

func (h *<<.NameUpper>>Handler) MandarSuccess(ctx context.Context, w http.ResponseWriter, msg string, targetTable string) {
	var sb strings.Builder
	sb.WriteString("<div id=\"modal\" hx-swap-oob=\"true\"></div>")
	criterios := pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]{Size: 10}
	data, _ := h.svc.Listar(ctx, criterios)
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, targetTable, true))
	w.Header().Set("Content-Type", "text/html")
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("success", msg))
	_, _ = w.Write([]byte(sb.String()))
}

func (h *<<.NameUpper>>Handler) MandarErrorModalCreate(ctx context.Context, w http.ResponseWriter, msg string, errs map[string]string, item <<.Name>>.<<.NameUpper>>) {
	var sb strings.Builder
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(ctx, &sb, components.FormCreate<<.NameUpper>>V2(item, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *<<.NameUpper>>Handler) MandarErrorModalUpdate(ctx context.Context, w http.ResponseWriter, msg string, errs map[string]string, item *<<.Name>>.<<.NameUpper>>) {
	var sb strings.Builder
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))
	utils.RenderIgnore(ctx, &sb, components.FormUpdate<<.NameUpper>>V2(item, errs))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = w.Write([]byte(sb.String()))
}

func (h *<<.NameUpper>>Handler) MandarErrorGeneric(w http.ResponseWriter, r *http.Request, msg string) {
	ctx := r.Context()
	var sb strings.Builder
	criterios := pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]{Size: 10}
	data, _ := h.svc.Listar(ctx, criterios)
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, <<.Name>>.Rutas.Tabla, true))
	utils.RenderIgnore(ctx, &sb, compartidos.Mensaje("error", msg))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(sb.String()))
}
`

const tmplHandlerFiles = `package handler

import (
	shared "<<.Module>>/shared/errores"
	"<<.Module>>/shared/utils"
	"<<.Module>>/shared/views/components/compartidos"
	"net/http"
)

func (h *<<.NameUpper>>Handler) DownloadPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, err := h.svc.ReportePdf(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error al generar PDF de <<.Name>>", err)
		http.Error(w, shared.TraducirError(err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *<<.NameUpper>>Handler) DownloadExcel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	url, err := h.svc.ReporteExcel(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Error al generar Excel de <<.Name>>", err)
		http.Error(w, shared.TraducirError(err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (h *<<.NameUpper>>Handler) SuccessFileMessage(w http.ResponseWriter, r *http.Request) {
	utils.RenderIgnore(r.Context(), w, compartidos.Mensaje("success", shared.MensajeOperacionExitosaArchivoGenerado))
}
`

const tmplViewTable = `package components

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/pagination"
	"<<.Module>>/shared/convertidores"
	"<<.Module>>/shared/views/components/v2"
)

templ TableV2(criterios pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>], baseURL string, oob bool) {
	{{
		rutas := <<.Name>>.Rutas
		buildSortURL := func(field string) string {
			return criterios.WithSort(field).BuildURL(rutas.Tabla, 1)
		}

		headers := []v2.TableHeader{
			{Label: "Nombre", Key: "nombre", Sortable: true, SortURL: buildSortURL("nombre"), SortIcon: criterios.SortIcon("nombre"), FilterVal: criterios.GetFilterValue("nombre")},
			{Label: "Codigo", Key: "codigo", Sortable: true, SortURL: buildSortURL("codigo"), SortIcon: criterios.SortIcon("codigo"), FilterVal: criterios.GetFilterValue("codigo")},
			{Label: "Descripcion", Key: "descripcion", Sortable: false},
			{Label: "Estatus", Key: "is_active", Sortable: true, SortURL: buildSortURL("is_active"), SortIcon: criterios.SortIcon("is_active"), FilterVal: criterios.GetFilterValue("is_active")},
		}

		tableProps := v2.DataTableV2Props{
			Title:       "Catalogo de <<.PluralUpper>>",
			Subtitle:    "Gestion Institucional",
			Headers:     headers,
			CreateURL:   rutas.Nuevo,
			SearchURL:   rutas.Tabla,
			Target:      "#tabla<<.Name>>",
			ModalTarget: "#modal",
			OOB:         oob,
			SortedBy:    criterios.SortedBy,
			Direction:      criterios.Direction,
			Page:           criterios.Index,
			Size:           criterios.Size,
			Total:          criterios.Count,
			PDFExportURL:   rutas.Export.PDF,
			ExcelExportURL: rutas.Export.Excel,
			ToastURL:       rutas.Export.Notificacion,
		}
	}}
	@v2.DataTableV2(tableProps) {
		for _, item := range criterios.Data {
			@v2.DataRowV2(v2.DataRowV2Props{
				Cells: []string{
					item.Nombre,
					item.Codigo,
					item.Descripcion,
					convertidores.Traducir(item.IsActive),
				},
				EditURL:     rutas.GetEditar(item.ID),
				DeleteURL:   rutas.GetBorrar(item.ID),
				Target:      tableProps.Target,
				ModalTarget: tableProps.ModalTarget,
			})
		}
	}
}
`

const tmplViewForm = `package components

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/shared/convertidores"
	"<<.Module>>/shared/views/components/v2"
)

templ FormCreate<<.NameUpper>>V2(item <<.Name>>.<<.NameUpper>>, errors map[string]string) {
	@v2.ModalV2("Nuevo <<.NameUpper>>") {
		<form hx-post={ <<.Name>>.Rutas.Crear } hx-target="#modal" hx-swap="innerHTML" class="space-y-6">
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				@v2.InputFieldV2(v2.InputPropsV2{
					Label:       "Nombre",
					Name:        "nombre",
					Placeholder: "Ingrese nombre",
					Value:       item.Nombre,
					Required:    true,
					Errors:      errors,
				})
				@v2.InputFieldV2(v2.InputPropsV2{
					Label:       "Codigo",
					Name:        "codigo",
					Placeholder: "Ej. COD-001",
					Value:       item.Codigo,
					Required:    true,
					Errors:      errors,
				})
				@v2.SelectFieldV2(v2.SelectProps{
					Label:         "Estatus",
					Name:          "activo",
					Options:       convertidores.OptionFromEstatusGenerico(),
					SelectedValue: convertidores.Traducir(item.IsActive),
					Required:      true,
					Errors:        errors,
				})
				<div class="md:col-span-2">
					@v2.TextAreaFieldV2(v2.TextAreaPropsV2{
						Label:       "Descripcion",
						Name:        "descripcion",
						Placeholder: "Detalles adicionales",
						Value:       item.Descripcion,
						Errors:      errors,
					})
				</div>
			</div>
			@v2.ModalFooterV2("Guardar")
		</form>
	}
}

templ FormUpdate<<.NameUpper>>V2(item *<<.Name>>.<<.NameUpper>>, errors map[string]string) {
	@v2.ModalV2("Editar <<.NameUpper>>") {
		<form hx-put={ <<.Name>>.Rutas.GetUpdate(item.ID) } hx-target="#modal" hx-swap="innerHTML" class="space-y-6">
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				@v2.InputFieldV2(v2.InputPropsV2{
					Label:       "Nombre",
					Name:        "nombre",
					Placeholder: "Ingrese nombre",
					Value:       item.Nombre,
					Required:    true,
					Errors:      errors,
				})
				@v2.InputFieldV2(v2.InputPropsV2{
					Label:       "Codigo",
					Name:        "codigo",
					Placeholder: "Ej. COD-001",
					Value:       item.Codigo,
					Required:    true,
					Errors:      errors,
				})
				@v2.SelectFieldV2(v2.SelectProps{
					Label:         "Estatus",
					Name:          "activo",
					Options:       convertidores.OptionFromEstatusGenerico(),
					SelectedValue: convertidores.Traducir(item.IsActive),
					Required:      true,
					Errors:        errors,
				})
				<div class="md:col-span-2">
					@v2.TextAreaFieldV2(v2.TextAreaPropsV2{
						Label:       "Descripcion",
						Name:        "descripcion",
						Placeholder: "Detalles adicionales",
						Value:       item.Descripcion,
						Errors:      errors,
					})
				</div>
			</div>
			@v2.ModalFooterV2("Actualizar")
		</form>
	}
}
`

const tmplViewLayout = `package layouts

import (
	"<<.Module>>/<<.Name>>"
	"<<.Module>>/<<.Name>>/view/components"
	"<<.Module>>/pagination"
	"<<.Module>>/shared/pagina"
	"<<.Module>>/shared/views/layouts"
)

templ View<<.NameUpper>>V2(titulo string, data pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>], baseURL string) {
	@layouts.BaseConNav(titulo, pagina.Pagina("<<.Name>>")) {
		@components.TableV2(data, baseURL, false)
		<div id="modal"></div>
	}
}
`

const tmplMigrationUp = `-- <<.MigrNum>>_create_<<.Name>>_table.up.sql
-- Tabla y politicas para el modulo <<.NameUpper>>

CREATE TABLE IF NOT EXISTS <<.Table>> (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre      VARCHAR(150) NOT NULL,
    codigo      VARCHAR(50) NOT NULL,
    descripcion TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    tenant_id   UUID,
    razon_i     VARCHAR(40),
    razon_m     VARCHAR(40),
    fecha_i     TIMESTAMPTZ DEFAULT NOW(),
    fecha_m     TIMESTAMPTZ
);

-- Triggers de auditoria y fecha de modificacion
DROP TRIGGER IF EXISTS tr_set_fecha_m_<<.Name>> ON <<.Table>>;
CREATE TRIGGER tr_set_fecha_m_<<.Name>>
BEFORE UPDATE ON <<.Table>>
FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();

DROP TRIGGER IF EXISTS tr_auditoria_<<.Name>> ON <<.Table>>;
CREATE TRIGGER tr_auditoria_<<.Name>>
AFTER INSERT OR UPDATE OR DELETE ON <<.Table>>
FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('id');

-- Indice de aislamiento multi-tenant
CREATE INDEX IF NOT EXISTS idx_<<.Name>>_tenant_id ON <<.Table>>(tenant_id);

-- Row Level Security (RLS)
ALTER TABLE <<.Table>> ENABLE ROW LEVEL SECURITY;
ALTER TABLE <<.Table>> FORCE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = '<<.Schema>>' AND tablename = '<<.Name>>' AND policyname = 'tenant_isolation_<<.Name>>'
    ) THEN
        CREATE POLICY tenant_isolation_<<.Name>> ON <<.Table>>
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true) OR current_setting('app.tenant_id', true) = 'local')
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true) OR current_setting('app.tenant_id', true) = 'local');
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_runtime_tenant') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON <<.Table>> TO app_runtime_tenant;
    END IF;
END $$;

-- Registro de recurso y politicas ABAC
INSERT INTO sesion.recurso (nombre, descripcion, razon_i) VALUES
('<<.Name>>', 'Modulo <<.NameUpper>>', 'Carga Inicial')
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO sesion.politica (nombre, descripcion, perfil, recurso, accion, reglas, razon_i) VALUES
('Acceso <<.NameUpper>> Superadmin', 'Acceso al modulo <<.Name>>', 'superadmin', '<<.Name>>', 'consultar', '{"path_prefix": "<<.Prefix>>", "routes": [{"method": "*", "prefix": "<<.Prefix>>"}, {"method": "GET", "path": "<<.Prefix>>", "accion": "consultar"}]}', 'Carga Inicial'),
('Acceso <<.NameUpper>> Admin', 'Acceso al modulo <<.Name>>', 'admin', '<<.Name>>', 'consultar', '{"path_prefix": "<<.Prefix>>", "routes": [{"method": "*", "prefix": "<<.Prefix>>"}, {"method": "GET", "path": "<<.Prefix>>", "accion": "consultar"}]}', 'Carga Inicial'),
('Acceso <<.NameUpper>> Operador', 'Acceso al modulo <<.Name>>', 'operador', '<<.Name>>', 'consultar', '{"path_prefix": "<<.Prefix>>", "routes": [{"method": "*", "prefix": "<<.Prefix>>"}, {"method": "GET", "path": "<<.Prefix>>", "accion": "consultar"}]}', 'Carga Inicial')
ON CONFLICT (nombre) DO NOTHING;
`

const tmplMigrationDown = `-- <<.MigrNum>>_create_<<.Name>>_table.down.sql

DELETE FROM sesion.politica WHERE recurso = '<<.Name>>';
DELETE FROM sesion.recurso WHERE nombre = '<<.Name>>';

DROP TRIGGER IF EXISTS tr_auditoria_<<.Name>> ON <<.Table>>;
DROP TRIGGER IF EXISTS tr_set_fecha_m_<<.Name>> ON <<.Table>>;
DROP TABLE IF EXISTS <<.Table>>;
`
