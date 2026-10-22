package miperfil

import (
	"strings"

	"github.com/gofrs/uuid/v5"
)

// PerfilUsuario consolida los datos personales, de cuenta y de rol del usuario autenticado.
type PerfilUsuario struct {
	UuidPersona  uuid.UUID `json:"uuid_persona"`
	Nombre       string    `json:"nombre"`
	Apellido1    string    `json:"apellido1"`
	Apellido2    string    `json:"apellido2"`
	ClaveIssemym string    `json:"clave_issemym"`
	Curp         string    `json:"curp"`
	Rfc          string    `json:"rfc"`
	Email        string    `json:"email"`
	Perfil       string    `json:"perfil"`
	ImagenPerfil string    `json:"imagen_perfil"`
}

// NombreCompleto retorna el nombre formateado.
func (p PerfilUsuario) NombreCompleto() string {
	partes := []string{}
	if p.Nombre != "" {
		partes = append(partes, p.Nombre)
	}
	if p.Apellido1 != "" {
		partes = append(partes, p.Apellido1)
	}
	if p.Apellido2 != "" {
		partes = append(partes, p.Apellido2)
	}
	return strings.Join(partes, " ")
}

// CambioPasswordForm contiene los datos para la actualización de credenciales.
type CambioPasswordForm struct {
	PasswordActual       string `json:"password_actual"`
	PasswordNueva        string `json:"password_nueva"`
	PasswordConfirmacion string `json:"password_confirmacion"`
}

// MensajeFeedback representa una notificación de éxito o error para la vista.
type MensajeFeedback struct {
	Tipo    string `json:"tipo"` // "success" | "error"
	Mensaje string `json:"mensaje"`
}
