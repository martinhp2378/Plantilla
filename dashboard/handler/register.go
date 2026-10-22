package handler

import (
	"PlantillaInstitucional/app"

	"github.com/go-chi/chi/v5"
)

// Register inicializa el handler y monta las rutas del Dashboard.
func Register(a *app.App, r chi.Router) {
	h := NewDashboardHandler(a.Logger)
	h.RegisterRoutes(r)
}
