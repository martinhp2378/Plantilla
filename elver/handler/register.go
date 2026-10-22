package handler

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/navigation"
	"PlantillaInstitucional/elver"
	"PlantillaInstitucional/elver/repository"
	"PlantillaInstitucional/elver/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el repositorio, servicio, handler y monta las rutas de Elver.
func Register(a *app.App, r chi.Router) {
	validate := validator.New()
	repo := repository.NewElverRepository(a.Pools)
	svc := service.NewElverService(repo, validate, a.Logger, a.ManejadorArchivos)
	h := NewElverHandler(svc, a.Logger, validate)
	h.RegisterRoutes(r)

	navigation.RegistrarItemMenu(navigation.ItemMenu{
		Label:  "Elveres",
		URL:    elver.Rutas.Prefix,
		Icon:   "folder",
		Codigo: "elver",
	})
}
