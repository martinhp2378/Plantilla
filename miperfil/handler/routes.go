package handler

import (
	"PlantillaInstitucional/miperfil"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registra los endpoints del módulo de Mi Perfil en el enrutador.
func (h *MiPerfilHandler) RegisterRoutes(r chi.Router) {
	rutas := miperfil.Rutas
	r.Route(rutas.Prefix, func(r chi.Router) {
		r.Get("/", h.HandleVerPerfil)
		r.Post("/password", h.HandleCambiarPassword)
		r.Post("/imagen", h.HandleCambiarImagen)
	})
}
