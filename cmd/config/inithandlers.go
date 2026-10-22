package config

import (
	"PlantillaInstitucional/app"
	login_handler "PlantillaInstitucional/login/handlers"
)

func InitHandlers(a *app.App) {
	a.Handlers.Login = login_handler.NuevoHandlerLogin(a.Repositories.Login, a.Repositories.Login, a.ServicioEmail, a.Logger, a.ConstructorSesion, a.ConstructorSesionJWT)
}
