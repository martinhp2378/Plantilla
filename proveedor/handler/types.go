package handler

import (
	"PlantillaInstitucional/proveedor"
	"PlantillaInstitucional/proveedor/service"
	"PlantillaInstitucional/shared/errores"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/issemym-repos/logs"
)

type ProveedorHandler struct {
	svc             service.ProveedorServiceProvider
	logger          *logs.Logger
	validate        *validator.Validate
	sqlErrorHandler *errores.SQLErrorHandler
}

func NewProveedorHandler(svc service.ProveedorServiceProvider, logger *logs.Logger, validate *validator.Validate) *ProveedorHandler {
	return &ProveedorHandler{
		svc:             svc,
		logger:          logger,
		validate:        validate,
		sqlErrorHandler: errores.NewSQLErrorHandler(logger),
	}
}

func (h *ProveedorHandler) RegisterRoutes(r chi.Router) {
	rutas := proveedor.Rutas
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

func (h *ProveedorHandler) parseForm(r *http.Request) (proveedor.Proveedor, error) {
	if err := r.ParseForm(); err != nil {
		return proveedor.Proveedor{}, fmt.Errorf("error parseando formulario: %w", err)
	}

	return proveedor.Proveedor{
		Nombre:      r.FormValue("nombre"),
		Codigo:      r.FormValue("codigo"),
		Descripcion: r.FormValue("descripcion"),
		IsActive:    r.FormValue("activo") == "true" || r.FormValue("activo") == "on" || r.FormValue("activo") == "1",
	}, nil
}
