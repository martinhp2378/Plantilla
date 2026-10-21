package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issemym-repos/logs"
)

func TestDashboardHandler_Index(t *testing.T) {
	logger := logs.NuevoLogger(&logs.Config{Nivel: "DEBUG"})
	h := NewDashboardHandler(logger)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()

	h.Index(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("se esperaba status 200, obtenido %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Dashboard") {
		t.Errorf("se esperaba que el body contuviera 'Dashboard', pero no lo contiene")
	}
	if !strings.Contains(body, "tendencia-actividad-chart-container") {
		t.Errorf("se esperaba que el body contuviera el contenedor de gráfica de tendencia")
	}
	if !strings.Contains(body, "distribucion-estado-chart-container") {
		t.Errorf("se esperaba que el body contuviera el contenedor de gráfica de distribución")
	}
	if !strings.Contains(body, "Registros Totales") {
		t.Errorf("se esperaba que el body contuviera la tarjeta KPI 'Registros Totales'")
	}
}

func TestDashboardHandler_RegisterRoutes(t *testing.T) {
	logger := logs.NuevoLogger(&logs.Config{Nivel: "DEBUG"})
	h := NewDashboardHandler(logger)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	h.Index(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 al llamar Index, obtenido %d", rec.Code)
	}
}
