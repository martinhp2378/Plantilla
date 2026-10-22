package sesionutil

import (
	"testing"

	"github.com/issemym-repos/sesiones"
)

func TestPerfilSesionParaFirmar(t *testing.T) {
	tests := []struct {
		name    string
		perfil  string
		want    sesiones.Perfil
		wantErr bool
	}{
		{name: "superadmin", perfil: PerfilSuperAdmin, want: sesiones.PerfilSuperAdmin},
		{name: "admin", perfil: PerfilAdmin, want: sesiones.PerfilMedico},
		{name: "operativo", perfil: PerfilOperativo, want: sesiones.PerfilRecepcion},
		{name: "ejecutivo", perfil: PerfilEjecutivo, want: sesiones.PerfilAdminClinica},
		{name: "anonimo", perfil: PerfilAnonimo, want: sesiones.PerfilAnonimo},
		{name: "coordinacion", perfil: PerfilCoordinacion, want: sesiones.PerfilAnonimo},
		{name: "unknown", perfil: "medico", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PerfilSesionParaFirmar(tt.perfil)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("PerfilSesionParaFirmar(%q) = %q, want %q", tt.perfil, got, tt.want)
			}
		})
	}
}

func TestPerfilAplicacionDesdeSesion(t *testing.T) {
	tests := []struct {
		name   string
		perfil sesiones.Perfil
		want   string
	}{
		{name: "superadmin", perfil: sesiones.PerfilSuperAdmin, want: PerfilSuperAdmin},
		{name: "medico alias", perfil: sesiones.PerfilMedico, want: PerfilAdmin},
		{name: "recepcion alias", perfil: sesiones.PerfilRecepcion, want: PerfilOperativo},
		{name: "adminclinica alias", perfil: sesiones.PerfilAdminClinica, want: PerfilEjecutivo},
		{name: "anonimo", perfil: sesiones.PerfilAnonimo, want: PerfilAnonimo},
		{name: "unknown passthrough", perfil: sesiones.Perfil("externo"), want: "externo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PerfilAplicacionDesdeSesion(tt.perfil)
			if got != tt.want {
				t.Fatalf("PerfilAplicacionDesdeSesion(%q) = %q, want %q", tt.perfil, got, tt.want)
			}
		})
	}
}

func TestImagenPerfilValida(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "usa placeholder cuando viene vacia", input: "", want: SinImagenPerfil},
		{name: "usa placeholder cuando solo viene extension", input: ".png", want: SinImagenPerfil},
		{name: "acepta imagen png valida", input: "avatar.png", want: "avatar.png"},
		{name: "rechaza extension distinta a png", input: "avatar.jpg", want: SinImagenPerfil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ImagenPerfilValida(tt.input)
			if got != tt.want {
				t.Fatalf("ImagenPerfilValida(%q) = %q, esperado %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNombreCompleto(t *testing.T) {
	tests := []struct {
		name    string
		usuario Usuario
		want    string
	}{
		{
			name: "arma nombre completo con espacios internos",
			usuario: Usuario{
				Nombre:    "Juan",
				Apellido1: "Perez",
				Apellido2: "Lopez",
			},
			want: "Juan Perez Lopez",
		},
		{
			name: "retorna usuario invalido cuando no hay datos",
			usuario: Usuario{
				Nombre:    "   ",
				Apellido1: "",
				Apellido2: "",
			},
			want: "Usuario Inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.usuario.NombreCompleto()
			if got != tt.want {
				t.Fatalf("NombreCompleto() = %q, esperado %q", got, tt.want)
			}
		})
	}
}

func TestUsuarioInvalido(t *testing.T) {
	u := UsuarioInvalido()

	if u == nil {
		t.Fatal("UsuarioInvalido() devolvio nil")
	}
	if u.NombreCompleto() != "Usuario Inválido" {
		t.Fatalf("NombreCompleto usuario invalido = %q, esperado %q", u.NombreCompleto(), "Usuario Inválido")
	}
}
