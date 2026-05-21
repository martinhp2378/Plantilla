package handlers

import (
	"context"

	"PlantillaInstitucional/login"

	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
	"github.com/issemym-repos/sesiones"
)

type RepositorioLogin interface {
	ValidarEmailPassword(ctx context.Context,
		email string,
		password string) (*login.DatosSesion, error)
}

type HandlerLogin struct {
	RepositorioLogin
	Logger               *logs.Logger
	ConstructorSesion    *sesiones.ConstructorSesion
	ConstructorSesionJWT *sesiones.ConstructorSesionJWT
	Archivos             *archivos.ManejadorArchivos
}

func NuevoHandlerLogin(repo RepositorioLogin, logger *logs.Logger, cs *sesiones.ConstructorSesion, csj *sesiones.ConstructorSesionJWT) *HandlerLogin {
	return &HandlerLogin{
		RepositorioLogin:     repo,
		Logger:               logger,
		ConstructorSesion:    cs,
		ConstructorSesionJWT: csj,
	}
}
