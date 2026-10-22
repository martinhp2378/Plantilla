package handler

import (
	"PlantillaInstitucional/app"

	"github.com/go-chi/chi/v5"
)

// Register inicializa el handler y monta las rutas del módulo Políticas.
func Register(a *app.App, r chi.Router) {
	h := NewPolicyHandler(a.Services.Policy, a.Logger)
	h.RegisterRoutes(r)
}
