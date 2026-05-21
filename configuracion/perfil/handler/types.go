package handler

import (
	"context"

	"PlantillaInstitucional/configuracion/perfil"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/errores"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/issemym-repos/logs"
)

// PerfilServiceProvider define el contrato del servicio de perfiles.
type PerfilServiceProvider interface {
	ConsultarPerfiles(ctx context.Context, criterios pagination.CriteriosPagina[perfil.Perfil]) (pagination.CriteriosPagina[perfil.Perfil], error)
	ConsultarPerfilPorNombre(ctx context.Context, nombre string) (perfil.Perfil, error)
	ObtenerPerfil(ctx context.Context, nombre string) (perfil.Perfil, error)
	ListarPerfiles(ctx context.Context, criterios pagination.CriteriosPagina[perfil.Perfil]) (pagination.CriteriosPagina[perfil.Perfil], error)
	CrearPerfil(ctx context.Context, p perfil.Perfil) error
	ActualizarPerfil(ctx context.Context, p perfil.Perfil, nombreOld string) error
	EliminarPerfil(ctx context.Context, nombre string) error
	ReportePdf(ctx context.Context) (string, error)
	ReporteExcel(ctx context.Context) (string, error)
}

// PerfilHandler es la implementación del handler de perfiles.
type PerfilHandler struct {
	validate        *validator.Validate
	sqlErrorHandler *shared.SQLErrorHandler
	service         PerfilServiceProvider
	logger          *logs.Logger
}

func NewPerfilHandler(svc PerfilServiceProvider, logger *logs.Logger, validate *validator.Validate) *PerfilHandler {
	return &PerfilHandler{
		service:         svc,
		logger:          logger,
		validate:        validate,
		sqlErrorHandler: shared.NewSQLErrorHandler(logger),
	}
}

// --- Helpers de Parseo de Requests (Privados) ---

func (h *PerfilHandler) RegisterRoutes(r chi.Router) {
	r.Route("/configuracion/perfiles", func(r chi.Router) {
		r.Get("/", h.Index)
		r.Get("/tabla", h.SearchTable)
		r.Get("/nuevo", h.ModalCreate)
		r.Post("/", h.Create)
		r.Get("/{id}", h.ModalUpdate)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Get("/pdf", h.DownloadPDF)
		r.Get("/excel", h.DownloadExcel)
	})
}

func (h *PerfilHandler) parseFormToPerfil(r *http.Request) (perfil.Perfil, error) {
	if err := r.ParseForm(); err != nil {
		return perfil.Perfil{}, fmt.Errorf("error parseando form: %w", err)
	}

	p := perfil.Perfil{
		Nombre: r.FormValue("nombre"),
	}

	return p, nil
}

func (h *PerfilHandler) parseNombreParam(r *http.Request) (string, error) {
	NombreParam := chi.URLParam(r, "id")
	if NombreParam == "" {
		return "", fmt.Errorf("id requerido")
	}

	return NombreParam, nil
}
