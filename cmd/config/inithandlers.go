package config

import (
	"PlantillaInstitucional/app"
	perfil_handler "PlantillaInstitucional/configuracion/perfil/handler"
	policy_handler "PlantillaInstitucional/configuracion/policy/handler"
	usuario_handler "PlantillaInstitucional/configuracion/usuario/handler"
	persona_handler "PlantillaInstitucional/institucion/handler"
	crud_handler "PlantillaInstitucional/crud/handler"
	dashboard_handler "PlantillaInstitucional/dashboard/handler"
	login_handler "PlantillaInstitucional/login/handlers"
	parametros_handler "PlantillaInstitucional/parametros/handler"

	"github.com/go-playground/validator/v10"
)

func InitHandlers(a *app.App) {
	validate := validator.New()

	a.Handlers.Usuario = usuario_handler.NewUsuarioHandler(a.Services.Usuario, &a.Services.Persona, a.EventDispatcher, a.Logger, validate)
	a.Handlers.Perfil = perfil_handler.NewPerfilHandler(a.Services.Perfil, a.Logger, validate)
	a.Handlers.Policy = policy_handler.NewPolicyHandler(a.Services.Policy, a.Logger)
	a.Handlers.Persona = persona_handler.NewPersonaHandler(&a.Services.Persona, a.Logger, validate, a.ManejadorArchivos)
	a.Handlers.Login = login_handler.NuevoHandlerLogin(a.Repositories.Login, a.Repositories.Login, a.ServicioEmail, a.Logger, a.ConstructorSesion, a.ConstructorSesionJWT)
	a.Handlers.Parametro = parametros_handler.NewHandler(a.Services.Parametro, a.Logger)
	a.Handlers.Crud = crud_handler.NewCrudHandler(&a.Services.Crud, a.Logger, validate, a.ManejadorArchivos)
	a.Handlers.Dashboard = dashboard_handler.NewDashboardHandler(a.Logger)
}
