package pagina

import "testing"

func TestPaginaPoliticasURL(t *testing.T) {
	if got := Politicas.URL(); got != "/configuracion/politicas/" {
		t.Fatalf("Politicas.URL() = %q, want %q", got, "/configuracion/politicas/")
	}
}

