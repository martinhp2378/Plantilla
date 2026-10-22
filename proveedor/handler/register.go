package handler

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/proveedor/repository"
	"PlantillaInstitucional/proveedor/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el repositorio, servicio, handler y monta las rutas de Proveedor.
func Register(a *app.App, r chi.Router) {
	validate := validator.New()
	repo := repository.NewProveedorRepository(a.Pools)
	svc := service.NewProveedorService(repo, validate, a.Logger)
	h := NewProveedorHandler(svc, a.Logger, validate)
	h.RegisterRoutes(r)
}
