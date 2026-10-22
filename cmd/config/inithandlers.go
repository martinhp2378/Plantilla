package config

import (
	"PlantillaInstitucional/app"
	policy_handler "PlantillaInstitucional/configuracion/policy/handler"
	login_handler "PlantillaInstitucional/login/handlers"
)

func InitHandlers(a *app.App) {
	a.Handlers.Policy = policy_handler.NewPolicyHandler(a.Services.Policy, a.Logger)
	a.Handlers.Login = login_handler.NuevoHandlerLogin(a.Repositories.Login, a.Repositories.Login, a.ServicioEmail, a.Logger, a.ConstructorSesion, a.ConstructorSesionJWT)
}
