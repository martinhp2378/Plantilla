package handler

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/crud/repository"
	"PlantillaInstitucional/crud/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el repositorio, servicio, handler y monta las rutas del módulo Crud.
func Register(a *app.App, r chi.Router) {
	validate := validator.New()
	repo := repository.NewCrudRepository(a.Pools)
	svc := service.NewCrudService(repo, validate, a.Logger, a.ManejadorArchivos)
	h := NewCrudHandler(svc, a.Logger, validate, a.ManejadorArchivos)
	h.RegisterRoutes(r)
}
