package handler

import (
	"PlantillaInstitucional/crud"
	"PlantillaInstitucional/pagination"
	"PlantillaInstitucional/shared/cargamasiva"
	errores "PlantillaInstitucional/shared/errores"
	"PlantillaInstitucional/shared/pagina"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

type CrudHandler interface {
	RegisterRoutes(r chi.Router)
	Index(w http.ResponseWriter, r *http.Request)
	SearchTable(w http.ResponseWriter, r *http.Request)
	ModalCreate(w http.ResponseWriter, r *http.Request)
	ModalUpdate(w http.ResponseWriter, r *http.Request)
	ModalCargaMasiva(w http.ResponseWriter, r *http.Request)
	DescargarPlantillaCSV(w http.ResponseWriter, r *http.Request)
	CargaMasiva(w http.ResponseWriter, r *http.Request)
	CloseModal(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	DownloadPDF(w http.ResponseWriter, r *http.Request)
	DownloadExcel(w http.ResponseWriter, r *http.Request)
	SuccessFileMessage(w http.ResponseWriter, r *http.Request)
}

type CrudServiceProvider interface {
	ListarCrud(ctx context.Context, criterios pagination.CriteriosPagina[crud.Crud]) (pagination.CriteriosPagina[crud.Crud], error)
	ObtenerCrud(ctx context.Context, id uuid.UUID) (crud.Crud, error)
	CrearCrud(ctx context.Context, c crud.Crud) error
	ActualizarCrud(ctx context.Context, c crud.Crud) error
	EliminarCrud(ctx context.Context, id uuid.UUID) error
	ObtenerTodos(ctx context.Context) ([]crud.Crud, error)
	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
	GenerarPlantillaCSV() ([]byte, error)
	CargarMasivo(ctx context.Context, r io.Reader) (*cargamasiva.ResultadoCarga[crud.Crud], error)
}

type crudHandler struct {
	service         CrudServiceProvider
	logger          *logs.Logger
	validate        *validator.Validate
	sqlErrorHandler *errores.SQLErrorHandler
	archivos        *archivos.ManejadorArchivos
}

func NewCrudHandler(service CrudServiceProvider, logger *logs.Logger, validate *validator.Validate, archivos *archivos.ManejadorArchivos) CrudHandler {
	return &crudHandler{
		service:         service,
		logger:          logger,
		validate:        validate,
		sqlErrorHandler: errores.NewSQLErrorHandler(logger),
		archivos:        archivos,
	}
}

func (h *crudHandler) RegisterRoutes(r chi.Router) {
	rutas := pagina.GlobalRutas.Crud
	r.Route(rutas.Prefix, func(r chi.Router) {
		r.Get("/", h.Index)
		r.Get("/tabla", h.SearchTable)
		r.Get("/modales/nuevo", h.ModalCreate)
		r.Post("/", h.Create)
		r.Get("/modales/editar/{id}", h.ModalUpdate)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Get("/modales/carga-masiva", h.ModalCargaMasiva)
		r.Get("/carga-masiva/plantilla", h.DescargarPlantillaCSV)
		r.Post("/carga-masiva", h.CargaMasiva)

		r.Route("/reportes", func(r chi.Router) {
			r.Get("/pdf", h.DownloadPDF)
			r.Get("/excel", h.DownloadExcel)
			r.Get("/notificacion-exito", h.SuccessFileMessage)
		})
	})
}

func (h *crudHandler) parseFormToCrud(r *http.Request) (crud.Crud, error) {
	if err := r.ParseForm(); err != nil {
		return crud.Crud{}, fmt.Errorf("error parseando form: %w", err)
	}

	entero, _ := strconv.Atoi(r.FormValue("entero"))
	decimal, _ := strconv.ParseFloat(r.FormValue("decimal"), 64)
	boolVal := r.FormValue("booleano") == "true" || r.FormValue("booleano") == "on" || r.FormValue("booleano") == "1"

	color := r.FormValue("color")
	if color == "" {
		color = "#691C32"
	}

	c := crud.Crud{
		TextoCorto: r.FormValue("texto_corto"),
		TextoLargo: r.FormValue("texto_largo"),
		Entero:     entero,
		Decimal:    decimal,
		Booleano:   boolVal,
		Fecha:      r.FormValue("fecha"),
		Hora:       r.FormValue("hora"),
		FechaHora:  r.FormValue("fecha_hora"),
		Seleccion:  r.FormValue("seleccion"),
		Correo:     r.FormValue("correo"),
		Telefono:   r.FormValue("telefono"),
		URL:        r.FormValue("url"),
		Color:      color,
		Razon_i:    r.FormValue("razon_i"),
		Razon_m:    r.FormValue("razon_m"),
	}

	return c, nil
}

func (h *crudHandler) parseIDParam(r *http.Request) (string, error) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		return "", fmt.Errorf("id requerido")
	}
	if _, err := uuid.FromString(idParam); err != nil {
		return "", fmt.Errorf("formato de id inválido")
	}
	return idParam, nil
}
