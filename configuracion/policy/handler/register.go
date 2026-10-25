package handler

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/configuracion/policy/repository"
	"PlantillaInstitucional/configuracion/policy/service"

	"github.com/go-chi/chi/v5"
)

// Register inicializa el handler y monta las rutas del módulo Políticas.
func Register(a *app.App, r chi.Router) {
	repo := repository.NewPolicyRepository(a.Pools)
	svc := service.NewService(repo)
	h := NewPolicyHandler(svc, a.Logger)
	h.RegisterRoutes(r)
}
