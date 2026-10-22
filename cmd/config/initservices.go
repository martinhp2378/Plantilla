package config

import (
	"PlantillaInstitucional/app"
	policy_service "PlantillaInstitucional/configuracion/policy/service"
	persona_repo "PlantillaInstitucional/institucion/repository"
	persona_service "PlantillaInstitucional/institucion/service"
	sharedEvents "PlantillaInstitucional/shared/events"

	"github.com/go-playground/validator/v10"
)

func InitServices(a *app.App) {
	validate := validator.New()

	// 1. Dispatcher
	dispatcher := sharedEvents.NewDispatcher()
	a.EventDispatcher = dispatcher

	// 2. Core Services
	a.Services.Policy = policy_service.NewService(a.Repositories.Policy)
	personaRepo := persona_repo.NewPersonaRepository(a.Pools)
	a.Services.Persona = *persona_service.NewPersonaService(personaRepo, validate, a.Logger, a.ManejadorArchivos)
}
