package handler

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/institucion/repository"
	"PlantillaInstitucional/institucion/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el handler y monta las rutas del módulo Institución (Personas).
func Register(a *app.App, r chi.Router) {
	validate := validator.New()
	repo := repository.NewPersonaRepository(a.Pools)
	svc := service.NewPersonaService(repo, validate, a.Logger, a.ManejadorArchivos)
	h := NewPersonaHandler(svc, a.Logger, validate, a.ManejadorArchivos)
	h.RegisterRoutes(r)
}
