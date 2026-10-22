package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type Config struct {
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
	nameFlag := flag.String("name", "", "Module name in singular (e.g. proveedor, articulo)")
	schemaFlag := flag.String("schema", "sistema", "Database schema (default: sistema)")
	titleFlag := flag.String("title", "", "Human readable title (default: auto capitalized)")
	pluralFlag := flag.String("plural", "", "Plural name (default: auto)")
	noMountFlag := flag.Bool("no-mount", false, "Do not auto-mount in cmd/main.go")
	flag.Parse()

	if *nameFlag == "" {
		fmt.Println("Error: el parametro -name es obligatorio. Ej: go run cmd/tools/scaffold/main.go -name proveedor")
		os.Exit(1)
	}

	name := strings.ToLower(strings.TrimSpace(*nameFlag))
	schema := strings.ToLower(strings.TrimSpace(*schemaFlag))

	plural := *pluralFlag
	if plural == "" {
		plural = autoPlural(name)
	}

	nameUpper := capitalize(name)
	pluralUpper := capitalize(plural)
	if *titleFlag != "" {
		pluralUpper = *titleFlag
	}

	prefix := fmt.Sprintf("/%s/%s", schema, plural)
	table := fmt.Sprintf("%s.%s", schema, name)
	migrNum := getNextMigrationNumber()

	cfg := Config{
		Name:        name,
		NameUpper:   nameUpper,
		Plural:      plural,
		PluralUpper: pluralUpper,
		Schema:      schema,
		Prefix:      prefix,
		Table:       table,
		MigrNum:     migrNum,
	}

	fmt.Printf("\n🚀 Generando modulo institucional: %s (%s)...\n", cfg.NameUpper, cfg.Prefix)

	if err := generateModuleFiles(cfg); err != nil {
		fmt.Printf("❌ Error generando archivos: %v\n", err)
		os.Exit(1)
	}

	if err := generateMigrations(cfg); err != nil {
		fmt.Printf("❌ Error generando migraciones: %v\n", err)
		os.Exit(1)
	}

	if !*noMountFlag {
		if err := mountInMain(cfg); err != nil {
			fmt.Printf("⚠️ No se pudo auto-montar en cmd/main.go: %v (podes agregarlo manualmente)\n", err)
		} else {
			fmt.Printf("✅ Montado automaticamente en cmd/main.go: %s_handler.Register(a, r)\n", cfg.Name)
		}
	}

	fmt.Println("⚙️ Ejecutando templ generate y templ fmt...")
	_ = exec.Command("templ", "generate").Run()
	_ = exec.Command("templ", "fmt", ".").Run()

	fmt.Printf("\n✨ Modulo '%s' generado exitosamente!\n", cfg.NameUpper)
	fmt.Printf("📁 Carpeta: ./%s\n", cfg.Name)
	fmt.Printf("📜 Migracion: ./sql/migrations/%s_create_%s_table.up.sql\n", cfg.MigrNum, cfg.Name)
	fmt.Printf("🌐 Endpoint: http://localhost:8080%s\n\n", cfg.Prefix)
}

func autoPlural(s string) string {
	last := s[len(s)-1:]
	if strings.ContainsAny(last, "aeiou") {
		return s + "s"
	}
	return s + "es"
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
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
		filepath.Join(cfg.Name, "handler", "types.go"):                        tmplHandlerTypes,
		filepath.Join(cfg.Name, "handler", "register.go"):                     tmplHandlerRegister,
		filepath.Join(cfg.Name, "handler", "crud.go"):                         tmplHandlerCRUD,
		filepath.Join(cfg.Name, "handler", "views.go"):                        tmplHandlerViews,
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

func mountInMain(cfg Config) error {
	content, err := os.ReadFile(mainFile)
	if err != nil {
		return err
	}
	s := string(content)

	importTarget := fmt.Sprintf("%s_handler \"PlantillaInstitucional/%s/handler\"", cfg.Name, cfg.Name)
	if strings.Contains(s, importTarget) {
		return nil // Ya montado
	}

	// Inyectar import
	importMarker := "import ("
	idx := strings.Index(s, importMarker)
	if idx == -1 {
		return fmt.Errorf("no se encontro bloque import en cmd/main.go")
	}
	insertImportPos := idx + len(importMarker)
	s = s[:insertImportPos] + "\n\t" + importTarget + s[insertImportPos:]

	// Inyectar llamada de Register en el grupo protegido
	registerLine := fmt.Sprintf("\t\t%s_handler.Register(a, r)\n", cfg.Name)
	groupMarker := "miperfil_handler.Register(a, r)"
	if strings.Contains(s, groupMarker) {
		s = strings.Replace(s, groupMarker, groupMarker+"\n"+registerLine, 1)
	} else {
		// Fallback: antes de la llave de cierre de NavigationMiddleware
		navMarker := "r.Use(app_middleware.NavigationMiddleware(a))"
		s = strings.Replace(s, navMarker, navMarker+"\n\n"+registerLine, 1)
	}

	// #nosec G703 -- mainFile is a constant target path in scaffolding CLI
	return os.WriteFile(mainFile, []byte(s), 0600)
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

import "PlantillaInstitucional/shared/pagina"

// Rutas contiene la configuracion canonica de rutas para el modulo <<.NameUpper>>.
var Rutas = pagina.NuevaRuta("<<.Prefix>>")
`

const tmplRepoTypes = `package repository

import (
	bd "PlantillaInstitucional/app/db"
	"PlantillaInstitucional/<<.Name>>"
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

	Insert<<.NameUpper>> = ` + "`" + `
		INSERT INTO <<.Table>> (nombre, codigo, descripcion, is_active, tenant_id, razon_i)
		VALUES ($1, $2, $3, $4, current_setting('app.tenant_id', true)::uuid, 'Creacion')
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
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/db"
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
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/pagination"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
)

type <<.NameUpper>>RepositoryProvider interface {
	Consultar(ctx context.Context, criterios pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]) ([]<<.Name>>.<<.NameUpper>>, int, error)
	ObtenerPorID(ctx context.Context, id uuid.UUID) (<<.Name>>.<<.NameUpper>>, error)
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
}

type <<.NameUpper>>Service struct {
	repo     <<.NameUpper>>RepositoryProvider
	validate *validator.Validate
	logger   *logs.Logger
}

func New<<.NameUpper>>Service(repo <<.NameUpper>>RepositoryProvider, validate *validator.Validate, logger *logs.Logger) *<<.NameUpper>>Service {
	return &<<.NameUpper>>Service{
		repo:     repo,
		validate: validate,
		logger:   logger,
	}
}
`

const tmplServiceMain = `package service

import (
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/pagination"
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

const tmplHandlerTypes = `package handler

import (
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/<<.Name>>/service"
	"PlantillaInstitucional/shared/errores"
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
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/navigation"
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/<<.Name>>/repository"
	"PlantillaInstitucional/<<.Name>>/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el repositorio, servicio, handler y monta las rutas de <<.NameUpper>>.
func Register(a *app.App, r chi.Router) {
	validate := validator.New()
	repo := repository.New<<.NameUpper>>Repository(a.Pools)
	svc := service.New<<.NameUpper>>Service(repo, validate, a.Logger)
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
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/shared/errores"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
)

func (h *<<.NameUpper>>Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	item, err := h.parseForm(r)
	if err != nil {
		h.MandarErrorModal(ctx, w, "Datos invalidos", err)
		return
	}

	if err := h.svc.Crear(ctx, item); err != nil {
		h.MandarErrorModal(ctx, w, "Error al guardar registro", err)
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroCreado, <<.Name>>.Rutas.Tabla)
}

func (h *<<.NameUpper>>Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		h.MandarErrorModal(ctx, w, "ID invalido", err)
		return
	}

	item, err := h.parseForm(r)
	if err != nil {
		h.MandarErrorModal(ctx, w, "Datos invalidos", err)
		return
	}
	item.ID = id

	if err := h.svc.Actualizar(ctx, item); err != nil {
		h.MandarErrorModal(ctx, w, "Error al actualizar registro", err)
		return
	}

	h.MandarSuccess(ctx, w, errores.MensajeOperacionExitosaRegistroActualizado, <<.Name>>.Rutas.Tabla)
}

func (h *<<.NameUpper>>Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "ID invalido", http.StatusBadRequest)
		return
	}

	if err := h.svc.Eliminar(ctx, id); err != nil {
		http.Error(w, "Error al eliminar registro", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
`

const tmplHandlerViews = `package handler

import (
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/<<.Name>>/view/components"
	"PlantillaInstitucional/<<.Name>>/view/layouts"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/utils"
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
	w.Header().Set("HX-Trigger", "cerrarModal")
	criterios := pagination.CriteriosPagina[<<.Name>>.<<.NameUpper>>]{Size: 10}
	data, _ := h.svc.Listar(ctx, criterios)
	var sb strings.Builder
	utils.RenderIgnore(ctx, &sb, components.TableV2(data, targetTable, true))
	_, _ = w.Write([]byte(sb.String()))
}

func (h *<<.NameUpper>>Handler) MandarErrorModal(ctx context.Context, w http.ResponseWriter, msg string, err error) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	errMap := map[string]string{"general": err.Error()}
	utils.RenderIgnore(ctx, w, components.FormCreate<<.NameUpper>>V2(<<.Name>>.<<.NameUpper>>{}, errMap))
}
`

const tmplViewTable = `package components

import (
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/convertidores"
	"PlantillaInstitucional/shared/views/components/v2"
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
			Direction:   criterios.Direction,
			Page:        criterios.Index,
			Size:        criterios.Size,
			Total:       criterios.Count,
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
				EditURL:   rutas.GetEditar(item.ID),
				DeleteURL: rutas.GetBorrar(item.ID),
			})
		}
	}
}
`

const tmplViewForm = `package components

import (
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/shared/convertidores"
	"PlantillaInstitucional/shared/views/components/v2"
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
			<div class="flex justify-end gap-3 pt-4 border-t border-gray-100">
				<button type="button" @click="open = false" class="px-4 py-2 text-xs font-bold text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg uppercase tracking-wider">Cancelar</button>
				<button type="submit" class="px-6 py-2 text-xs font-bold text-white bg-[var(--color-guinda)] hover:bg-[#521426] rounded-lg uppercase tracking-wider">Guardar</button>
			</div>
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
			<div class="flex justify-end gap-3 pt-4 border-t border-gray-100">
				<button type="button" @click="open = false" class="px-4 py-2 text-xs font-bold text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg uppercase tracking-wider">Cancelar</button>
				<button type="submit" class="px-6 py-2 text-xs font-bold text-white bg-[var(--color-guinda)] hover:bg-[#521426] rounded-lg uppercase tracking-wider">Actualizar</button>
			</div>
		</form>
	}
}
`

const tmplViewLayout = `package layouts

import (
	"PlantillaInstitucional/<<.Name>>"
	"PlantillaInstitucional/<<.Name>>/view/components"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/views/layouts"
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
            USING (tenant_id::text = current_setting('app.tenant_id', true))
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
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
