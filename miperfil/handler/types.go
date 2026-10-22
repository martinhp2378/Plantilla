package handler

import (
	"PlantillaInstitucional/miperfil/service"

	"github.com/issemym-repos/logs"
	"github.com/issemym-repos/sesiones"
)

// MiPerfilHandler gestiona las solicitudes HTTP relacionadas con el perfil del usuario.
type MiPerfilHandler struct {
	service           service.ServiceProvider
	logger            *logs.Logger
	constructorSesion *sesiones.ConstructorSesion
}

// NewMiPerfilHandler inicializa el handler.
func NewMiPerfilHandler(service service.ServiceProvider, logger *logs.Logger, constructorSesion *sesiones.ConstructorSesion) *MiPerfilHandler {
	return &MiPerfilHandler{
		service:           service,
		logger:            logger,
		constructorSesion: constructorSesion,
	}
}
