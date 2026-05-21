package login

import (
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/sesiones"
)

type DatosSesion struct {
	UuidPersona uuid.UUID
	Email       string
	Perfil      sesiones.Perfil
}
