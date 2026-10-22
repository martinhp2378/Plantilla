package handler

import (
	"PlantillaInstitucional/institucion"
	"PlantillaInstitucional/pagination"
	errores "PlantillaInstitucional/shared/errores"
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
)

// PersonaHandler define las rutas y acciones HTTP disponibles.
type PersonaHandler interface {
	RegisterRoutes(r chi.Router)
	// Vistas y Componentes
	Index(w http.ResponseWriter, r *http.Request)
	SearchTable(w http.ResponseWriter, r *http.Request)

	// Modals
	ModalCreate(w http.ResponseWriter, r *http.Request)
	ModalUpdate(w http.ResponseWriter, r *http.Request)
	CloseModal(w http.ResponseWriter, r *http.Request)

	// CRUD Actions
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)

	// Archivos
	DownloadPDF(w http.ResponseWriter, r *http.Request)
	DownloadExcel(w http.ResponseWriter, r *http.Request)
	SuccessFileMessage(w http.ResponseWriter, r *http.Request)
}

// PersonaServiceProvider define el contrato que el Handler espera de la capa de servicio.
type PersonaServiceProvider interface {
	ListarPersonas(ctx context.Context, criterios pagination.CriteriosPagina[institucion.Persona]) (pagination.CriteriosPagina[institucion.Persona], error)
	ObtenerPersona(ctx context.Context, uuid_persona uuid.UUID) (institucion.Persona, error)
	CrearPersona(ctx context.Context, p institucion.Persona) error
	ActualizarPersona(ctx context.Context, p institucion.Persona) error
	EliminarPersona(ctx context.Context, uuid_persona uuid.UUID) error
	ObtenerTodasCorto(ctx context.Context) ([]institucion.Persona, error)

	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
}

type personaHandler struct {
	validate        *validator.Validate
	sqlErrorHandler *errores.SQLErrorHandler
	service         PersonaServiceProvider
	logger          *logs.Logger
	archivos        *archivos.ManejadorArchivos
}

func NewPersonaHandler(service PersonaServiceProvider, logger *logs.Logger, validate *validator.Validate, archivos *archivos.ManejadorArchivos) PersonaHandler {
	return &personaHandler{
		service:         service,
		logger:          logger,
		validate:        validate,
		sqlErrorHandler: errores.NewSQLErrorHandler(logger),
		archivos:        archivos,
	}
}

// --- Helpers de Parseo de Requests (Privados) ---

func (h *personaHandler) RegisterRoutes(r chi.Router) {
	rutas := institucion.Rutas
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

func (h *personaHandler) parseFormToPersona(r *http.Request) (institucion.Persona, error) {
	if err := r.ParseForm(); err != nil {
		return institucion.Persona{}, fmt.Errorf("error parseando form: %w", err)
	}

	b := institucion.Persona{
		Nombre:       r.FormValue("nombre"),
		Apellido1:    r.FormValue("apellido1"),
		Apellido2:    r.FormValue("apellido2"),
		Rfc:          r.FormValue("rfc"),
		Curp:         r.FormValue("curp"),
		ClaveIssemym: r.FormValue("clave_issemym"),
		Razon_i:      r.FormValue("razon_i"),
	}

	return b, nil
}

func (h *personaHandler) parseIDParam(r *http.Request) (string, error) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		return "", fmt.Errorf("id requerido")
	}
	// Validamos formato brevemente
	if _, err := uuid.FromString(idParam); err != nil {
		return "", fmt.Errorf("formato de id inválido")
	}
	return idParam, nil
}
