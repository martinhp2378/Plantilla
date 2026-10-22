package sesionutil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/sesiones"
)

const SinImagenPerfil = "placeholder.png"

const (
	PerfilSuperAdmin   = "superadmin"
	PerfilAdmin        = "admin"
	PerfilOperativo    = "operativo"
	PerfilEjecutivo    = "ejecutivo"
	PerfilAnonimo      = "anonimo"
	PerfilCoordinacion = "coordinacion"

	DatoExtraPerfilAplicacion = "perfil_app"
	DatoExtraTenantSchema     = "tenant_schema"
)

// PerfilSesionParaFirmar maps canonical application profiles to perfiles accepted by sesiones.
func PerfilSesionParaFirmar(perfilAplicacion string) (sesiones.Perfil, error) {
	switch perfilAplicacion {
	case PerfilSuperAdmin:
		return sesiones.PerfilSuperAdmin, nil
	case PerfilAdmin:
		return sesiones.PerfilMedico, nil
	case PerfilOperativo:
		return sesiones.PerfilRecepcion, nil
	case PerfilEjecutivo:
		return sesiones.PerfilAdminClinica, nil
	case PerfilAnonimo, PerfilCoordinacion:
		return sesiones.PerfilAnonimo, nil
	default:
		return "", fmt.Errorf("perfil de aplicacion desconocido: %s", perfilAplicacion)
	}
}

// PerfilAplicacionDesdeSesion maps perfiles from existing sesiones/JWTs to canonical application profiles.
func PerfilAplicacionDesdeSesion(perfil sesiones.Perfil) string {
	switch perfil {
	case sesiones.PerfilSuperAdmin:
		return PerfilSuperAdmin
	case sesiones.PerfilMedico:
		return PerfilAdmin
	case sesiones.PerfilRecepcion:
		return PerfilOperativo
	case sesiones.PerfilAdminClinica:
		return PerfilEjecutivo
	case sesiones.PerfilAnonimo:
		return PerfilAnonimo
	default:
		return string(perfil)
	}
}

// Usuario es un tipo de dato que acumula algunos datos comunes entre
// diferentes tipos de usuario. Se debe de instanciar y agregar a la sesión a
// la hora de firmar un usuario. Se utiliza en las plantillas templ base.
type Usuario struct {
	UuidPersona         string `json:"uuid_persona"`
	Nombre              string `json:"nombre"`
	Apellido1           string `json:"apellido1"`
	Apellido2           string `json:"apellido2"`
	Curp                string `json:"curp"`
	ClaveIssemym        string `json:"clave_issemym"`
	ArchivoImagenPerfil string `json:"archivo_imagen_perfil"`
	Email               string `json:"email"`
	Perfil              string `json:"perfil"`
	TenantSchema        string `json:"tenant_schema"`
}

func ImagenPerfilValida(nombreImagenPerfil string) string {
	nombreImagenPerfil = strings.TrimSpace(nombreImagenPerfil)
	if nombreImagenPerfil == "" || nombreImagenPerfil == SinImagenPerfil {
		return SinImagenPerfil
	}
	base := strings.TrimSuffix(nombreImagenPerfil, filepath.Ext(nombreImagenPerfil))
	if base == "" {
		return SinImagenPerfil
	}
	ext := strings.ToLower(filepath.Ext(nombreImagenPerfil))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		return nombreImagenPerfil
	default:
		return SinImagenPerfil
	}
}

func (u *Usuario) UrlImagenPerfil(ctx context.Context) string {
	return archivos.Asset(ctx,
		filepath.Join("img", "perfil", ImagenPerfilValida(u.ArchivoImagenPerfil)))
}

func UrlImagenPerfil(ctx context.Context) string {
	// Safe fallback - handle nil session without panic
	u := UsuarioDeContexto(ctx)
	if u == nil {
		return archivos.Asset(ctx, filepath.Join("img", "perfil", SinImagenPerfil))
	}
	return u.UrlImagenPerfil(ctx)
}

func (u *Usuario) NombreCompleto() string {
	nom := strings.TrimSpace(u.Nombre + " " + u.Apellido1 + " " + u.Apellido2)
	if nom == "" {
		return "Usuario Inválido"
	}
	return nom
}

func UsuarioInvalido() *Usuario {
	return &Usuario{}
}

func UsuarioDeContexto(ctx context.Context) (usuario *Usuario) {
	// Recover from any panic inside sesiones.RecuperarSesion (e.g., nil interface assertion)
	defer func() {
		if r := recover(); r != nil {
			// Return safe fallback on any panic
			usuario = UsuarioInvalido()
		}
	}()

	sesion := sesiones.RecuperarSesion(ctx)

	// Handle nil session gracefully - return invalid user instead of panicking
	if sesion == nil {
		return UsuarioInvalido()
	}

	u, err := sesiones.GetStruct[Usuario](sesion)

	if err != nil {
		return UsuarioInvalido()
	}
	return &u
}

func ObtenerUsuarioParaBd(ctx context.Context) string {
	return UsuarioDeContexto(ctx).UuidPersona
}

// RecoverarSesion returns the session from context, or nil if not found.
// It recovers from any panic that might occur inside sesiones.RecuperarSesion
// when the context doesn't have a valid session.
func RecoverarSesion(ctx context.Context) interface{} {
	var result interface{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				result = nil
			}
		}()
		result = sesiones.RecuperarSesion(ctx)
	}()
	return result
}

// PopFlashes retrieves flash messages from the session safely.
// Returns nil if session is not available.
func PopFlashes(ctx context.Context) []sesiones.MensajeFlash {
	sesion := RecoverarSesion(ctx)
	if sesion == nil {
		return nil
	}
	if s, ok := sesion.(*sesiones.Sesion); ok {
		return s.PopFlashes()
	}
	return nil
}
