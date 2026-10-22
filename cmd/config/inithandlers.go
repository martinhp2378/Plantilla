package config

import (
	"PlantillaInstitucional/app"
	perfil_handler "PlantillaInstitucional/configuracion/perfil/handler"
	policy_handler "PlantillaInstitucional/configuracion/policy/handler"
	login_handler "PlantillaInstitucional/login/handlers"

	"github.com/go-playground/validator/v10"
)

func InitHandlers(a *app.App) {
	validate := validator.New()

	a.Handlers.Perfil = perfil_handler.NewPerfilHandler(a.Services.Perfil, a.Logger, validate)
	a.Handlers.Policy = policy_handler.NewPolicyHandler(a.Services.Policy, a.Logger)
	a.Handlers.Login = login_handler.NuevoHandlerLogin(a.Repositories.Login, a.Repositories.Login, a.ServicioEmail, a.Logger, a.ConstructorSesion, a.ConstructorSesionJWT)
}
