package handler

import (
	"PlantillaInstitucional/app"
	usuario_events "PlantillaInstitucional/configuracion/usuario/events"
	usuario_listeners "PlantillaInstitucional/configuracion/usuario/listeners"
	"PlantillaInstitucional/configuracion/usuario/repository"
	"PlantillaInstitucional/configuracion/usuario/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Register inicializa el repositorio, servicio, listeners de eventos, handler y monta las rutas de Usuario.
func Register(a *app.App, r chi.Router, personaProviders ...PersonaServiceProvider) {
	validate := validator.New()
	repo := repository.NewUsuarioRepository(a.Pools)
	svc := service.NewUsuarioService(repo, validate, a.Logger, a.ManejadorArchivos, a.EventDispatcher)

	if a.EventDispatcher != nil && a.ServicioEmail != nil {
		userEmailListener := usuario_listeners.NewUserEmailListener(a.ServicioEmail)
		a.EventDispatcher.Register(usuario_events.SendUserActivationEvent{}.Name(), userEmailListener)
	}

	var personaProvider PersonaServiceProvider
	if len(personaProviders) > 0 {
		personaProvider = personaProviders[0]
	}

	h := NewUsuarioHandler(svc, personaProvider, a.EventDispatcher, a.Logger, validate)
	h.RegisterRoutes(r)
}
