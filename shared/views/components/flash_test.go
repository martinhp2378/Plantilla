package components

import (
	"context"
	"strings"
	"testing"

	"github.com/issemym-repos/sesiones"
)

func TestNormalizarVarianteFlash(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "exito a success", in: "exito", want: "success"},
		{name: "success queda success", in: "success", want: "success"},
		{name: "error queda error", in: "error", want: "error"},
		{name: "advertencia a warning", in: "advertencia", want: "warning"},
		{name: "desconocido a info", in: "otro", want: "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizarVarianteFlash(tt.in)
			if got != tt.want {
				t.Fatalf("normalizarVarianteFlash(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRenderFlashMessage_RenderizaToastSinRutaNotificacion(t *testing.T) {
	msg := sesiones.MensajeFlash{Tipo: "error", Contenido: "Sin permisos"}

	var sb strings.Builder
	err := renderFlashMessage(msg).Render(context.Background(), &sb)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	out := sb.String()
	if !strings.Contains(out, "#toast-container") {
		t.Fatalf("output should target toast container, got: %s", out)
	}
	if strings.Contains(out, "/notificacion?") {
		t.Fatalf("output should not include /notificacion route: %s", out)
	}
}
