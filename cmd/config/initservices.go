package config

import (
	"PlantillaInstitucional/app"
	perfil_service "PlantillaInstitucional/configuracion/perfil/service"
	policy_service "PlantillaInstitucional/configuracion/policy/service"
	usuario_events "PlantillaInstitucional/configuracion/usuario/events"
	usuario_listeners "PlantillaInstitucional/configuracion/usuario/listeners"
	usuario_service "PlantillaInstitucional/configuracion/usuario/service"
	persona_service "PlantillaInstitucional/institucion/service"
	parametros_service "PlantillaInstitucional/parametros/service"
	sharedEvents "PlantillaInstitucional/shared/events"

	"github.com/go-playground/validator/v10"
)

func InitServices(a *app.App) {
	validate := validator.New()

	// 1. Dispatcher
	dispatcher := sharedEvents.NewDispatcher()
	a.EventDispatcher = dispatcher

	// 2. Core Services
	a.Services.Parametro = parametros_service.NewService(a.Logger, a.Repositories.Parametro, validate, a.ManejadorArchivos)

	// Inject Parameters into Email Service
	a.ServicioEmail.SetParametersProvider(a.Services.Parametro)

	a.Services.Usuario = usuario_service.NewUsuarioService(a.Repositories.Usuario, validate, a.Logger, a.ManejadorArchivos, dispatcher)
	a.Services.Perfil = perfil_service.NewPerfilService(a.Repositories.Perfil, validate, a.Logger, a.ManejadorArchivos)
	a.Services.Policy = policy_service.NewService(a.Repositories.Policy)
	a.Services.Persona = *persona_service.NewPersonaService(a.Repositories.Persona, validate, a.Logger, a.ManejadorArchivos)

	// 3. Listeners
	userEmailListener := usuario_listeners.NewUserEmailListener(a.ServicioEmail)

	// 4. Register Listeners
	dispatcher.Register(usuario_events.SendUserActivationEvent{}.Name(), userEmailListener)
}
