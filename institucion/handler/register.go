package handler

import (
	"PlantillaInstitucional/app"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el handler y monta las rutas del módulo Institución (Personas).
func Register(a *app.App, r chi.Router) {
	validate := validator.New()
	h := NewPersonaHandler(&a.Services.Persona, a.Logger, validate, a.ManejadorArchivos)
	h.RegisterRoutes(r)
}
