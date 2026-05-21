package pagina

import "testing"

func TestPaginaPoliticasURL(t *testing.T) {
	if got := Politicas.URL(); got != "/configuracion/politicas" {
		t.Fatalf("Politicas.URL() = %q, want %q", got, "/configuracion/politicas")
	}
}

func TestPaginaDashboardPlantillaURL(t *testing.T) {
	if got := DashboardPlantilla.URL(); got != "/dashboard_plantilla/configuracion" {
		t.Fatalf("DashboardPlantilla.URL() = %q, want %q", got, "/dashboard_plantilla/configuracion")
	}
}
