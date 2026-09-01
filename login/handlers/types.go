package handlers

import (
	"context"
	"time"

	"PlantillaInstitucional/login"

	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
	"github.com/issemym-repos/sesiones"
)

type AccountRecoveryProvider interface {
	IssuePasswordReset(ctx context.Context, email string, tokenHash []byte, expiresAt time.Time) (bool, error)
	ConsumePasswordReset(ctx context.Context, tokenHash []byte, passwordHash string) error
}

type ActivationEmailSender interface {
	SendUserActivation(ctx context.Context, to, name, username, activationURL string) error
}

type RepositorioLogin interface {
	ValidarEmailPassword(ctx context.Context,
		email string,
		password string) (*login.DatosSesion, error)
}

type HandlerLogin struct {
	RepositorioLogin
	AccountRecoveryProvider
	activationEmailSender ActivationEmailSender
	resetLimiter          *resetRateLimiter
	Logger                *logs.Logger
	ConstructorSesion     *sesiones.ConstructorSesion
	ConstructorSesionJWT  *sesiones.ConstructorSesionJWT
	Archivos              *archivos.ManejadorArchivos
}

func NuevoHandlerLogin(repo RepositorioLogin, recovery AccountRecoveryProvider, sender ActivationEmailSender, logger *logs.Logger, cs *sesiones.ConstructorSesion, csj *sesiones.ConstructorSesionJWT) *HandlerLogin {
	return &HandlerLogin{
		RepositorioLogin:        repo,
		AccountRecoveryProvider: recovery,
		activationEmailSender:   sender,
		resetLimiter:            newResetRateLimiter(),
		Logger:                  logger,
		ConstructorSesion:       cs,
		ConstructorSesionJWT:    csj,
	}
}
