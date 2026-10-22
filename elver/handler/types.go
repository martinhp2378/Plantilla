package handler

import (
	"PlantillaInstitucional/elver"
	"PlantillaInstitucional/elver/service"
	"PlantillaInstitucional/shared/errores"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/issemym-repos/logs"
)

type ElverHandler struct {
	svc             service.ElverServiceProvider
	logger          *logs.Logger
	validate        *validator.Validate
	sqlErrorHandler *errores.SQLErrorHandler
}

func NewElverHandler(svc service.ElverServiceProvider, logger *logs.Logger, validate *validator.Validate) *ElverHandler {
	return &ElverHandler{
		svc:             svc,
		logger:          logger,
		validate:        validate,
		sqlErrorHandler: errores.NewSQLErrorHandler(logger),
	}
}

func (h *ElverHandler) RegisterRoutes(r chi.Router) {
	rutas := elver.Rutas
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

func (h *ElverHandler) parseForm(r *http.Request) (elver.Elver, error) {
	if err := r.ParseForm(); err != nil {
		return elver.Elver{}, fmt.Errorf("error parseando formulario: %w", err)
	}

	return elver.Elver{
		Nombre:      r.FormValue("nombre"),
		Codigo:      r.FormValue("codigo"),
		Descripcion: r.FormValue("descripcion"),
		IsActive:    r.FormValue("activo") == "true" || r.FormValue("activo") == "on" || r.FormValue("activo") == "1",
	}, nil
}
