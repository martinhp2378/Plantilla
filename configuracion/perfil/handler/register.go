package handler

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/configuracion/perfil/repository"
	"PlantillaInstitucional/configuracion/perfil/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el repositorio, servicio, handler y monta las rutas de Perfiles.
func Register(a *app.App, r chi.Router) {
	validate := validator.New()
	repo := repository.NewPerfilRepository(a.Pools)
	svc := service.NewPerfilService(repo, validate, a.Logger, a.ManejadorArchivos)
	h := NewPerfilHandler(svc, a.Logger, validate)
	h.RegisterRoutes(r)
}
