package handler

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/parametros/repository"
	"PlantillaInstitucional/parametros/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el repositorio, servicio, inyecta el proveedor de parámetros al servicio de email,
// construye el handler y monta las rutas del módulo Parámetros.
func Register(a *app.App, r chi.Router) {
	validate := validator.New()
	repo := repository.NewRepository(a.Pools)
	svc := service.NewService(a.Logger, repo, validate, a.ManejadorArchivos)

	if a.ServicioEmail != nil {
		a.ServicioEmail.SetParametersProvider(svc)
	}

	h := NewHandler(svc, a.Logger)
	h.RegisterRoutes(r)
}
