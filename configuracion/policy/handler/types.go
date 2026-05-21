package handler

import (
	"PlantillaInstitucional/configuracion/policy"
	"PlantillaInstitucional/pagination"
	shared "PlantillaInstitucional/shared/errores"
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/issemym-repos/logs"
)

type PolicyServiceProvider interface {
	ListarPoliticas(ctx context.Context, criterios pagination.CriteriosPagina[policy.Politica]) (pagination.CriteriosPagina[policy.Politica], error)
	ObtenerPolitica(ctx context.Context, nombre string) (policy.Politica, error)
	CrearPolitica(ctx context.Context, p policy.Politica) error
	ActualizarPolitica(ctx context.Context, p policy.Politica, nombreOriginal string) error
	EliminarPolitica(ctx context.Context, nombre string) error
	CatalogosPolitica(ctx context.Context) ([]string, []string, []string, error)
}

type PolicyHandler struct {
	service         PolicyServiceProvider
	logger          *logs.Logger
	sqlErrorHandler *shared.SQLErrorHandler
}

func NewPolicyHandler(svc PolicyServiceProvider, logger *logs.Logger) *PolicyHandler {
	return &PolicyHandler{service: svc, logger: logger, sqlErrorHandler: shared.NewSQLErrorHandler(logger)}
}

func (h *PolicyHandler) RegisterRoutes(r chi.Router) {
	r.Route("/configuracion/politicas", func(r chi.Router) {
		r.Get("/", h.Index)
		r.Get("/tabla", h.SearchTable)
		r.Get("/nuevo", h.ModalCreate)
		r.Post("/", h.Create)
		r.Get("/{id}", h.ModalUpdate)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *PolicyHandler) parseForm(r *http.Request) (policy.Politica, error) {
	if err := r.ParseForm(); err != nil {
		return policy.Politica{}, fmt.Errorf("error parseando form: %w", err)
	}
	return policy.Politica{
		Nombre:      r.FormValue("nombre"),
		Descripcion: r.FormValue("descripcion"),
		Perfil:      r.FormValue("perfil"),
		Recurso:     r.FormValue("recurso"),
		Accion:      r.FormValue("accion"),
		PathPrefix:  r.FormValue("path_prefix"),
	}, nil
}

func (h *PolicyHandler) parseNombre(r *http.Request) (string, error) {
	nombre := chi.URLParam(r, "id")
	if nombre == "" {
		return "", fmt.Errorf("id requerido")
	}
	return nombre, nil
}
