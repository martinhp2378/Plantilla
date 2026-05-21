package sesionutil

import "testing"

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
