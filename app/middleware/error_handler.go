package middleware

import (
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/layouts"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegistrarHandler404 configura la página de error 404 institucional para rutas no encontradas.
func RegistrarHandler404(r chi.Router) {
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		utils.RenderIgnore(r.Context(), w, layouts.Pagina404())
	})
}
