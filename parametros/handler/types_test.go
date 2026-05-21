package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestParametroFromFormCreate(t *testing.T) {
	req := httptest.NewRequest("POST", "/parametros", strings.NewReader(url.Values{
		"clave":       {"MAX_INTENTOS"},
		"valor":       {"5"},
		"descripcion": {"Limite de intentos"},
		"tipo_dato":   {"INT"},
		"is_active":   {"on"},
		"razon_i":     {"Carga inicial"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	got, err := ParametroFromFormCreate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ParametroId.String() == "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("expected generated UUID, got nil UUID")
	}
	if got.Clave != "MAX_INTENTOS" {
		t.Fatalf("unexpected clave: %q", got.Clave)
	}
	if got.Valor != "5" {
		t.Fatalf("unexpected valor: %q", got.Valor)
	}
	if got.Descripcion != "Limite de intentos" {
		t.Fatalf("unexpected descripcion: %q", got.Descripcion)
	}
	if got.TipoDato != "INT" {
		t.Fatalf("unexpected tipo_dato: %q", got.TipoDato)
	}
	if !got.IsActive {
		t.Fatalf("expected is_active=true")
	}
	if got.RazonI != "Carga inicial" {
		t.Fatalf("unexpected razon_i: %q", got.RazonI)
	}
}

func TestParametroFromFormUpdate_Success(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest("POST", "/parametros/"+id, strings.NewReader(url.Values{
		"clave":       {"TIEMPO_SESION"},
		"valor":       {"30"},
		"descripcion": {"Minutos de sesion"},
		"tipo_dato":   {"INT"},
		"is_active":   {"on"},
		"razon_m":     {"Ajuste operativo"},
	}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	req = req.WithContext(contextWithRouteCtx(req, rctx))

	got, err := ParametroFromFormUpdate(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ParametroId.String() != id {
		t.Fatalf("unexpected id: %s", got.ParametroId.String())
	}
	if got.Clave != "TIEMPO_SESION" {
		t.Fatalf("unexpected clave: %q", got.Clave)
	}
	if got.Valor != "30" {
		t.Fatalf("unexpected valor: %q", got.Valor)
	}
	if got.Descripcion != "Minutos de sesion" {
		t.Fatalf("unexpected descripcion: %q", got.Descripcion)
	}
	if got.TipoDato != "INT" {
		t.Fatalf("unexpected tipo_dato: %q", got.TipoDato)
	}
	if !got.IsActive {
		t.Fatalf("expected is_active=true")
	}
	if got.RazonM != "Ajuste operativo" {
		t.Fatalf("unexpected razon_m: %q", got.RazonM)
	}
}

func TestParametroFromFormUpdate_ParseFormError(t *testing.T) {
	req := httptest.NewRequest("POST", "/parametros/550e8400-e29b-41d4-a716-446655440000", strings.NewReader("clave=%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "550e8400-e29b-41d4-a716-446655440000")
	req = req.WithContext(contextWithRouteCtx(req, rctx))

	_, err := ParametroFromFormUpdate(req)
	if err == nil {
		t.Fatalf("expected parse form error")
	}
}

func TestParametroFromFormUpdate_InvalidUUID(t *testing.T) {
	req := httptest.NewRequest("POST", "/parametros/invalid-id", strings.NewReader("clave=K&valor=V"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-id")
	req = req.WithContext(contextWithRouteCtx(req, rctx))

	_, err := ParametroFromFormUpdate(req)
	if err == nil {
		t.Fatalf("expected invalid UUID error")
	}
}

func TestParametroFromFormSimple(t *testing.T) {
	req := httptest.NewRequest("POST", "/parametros/tabla", strings.NewReader("clave=CLAVE&valor=VALOR"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	got, err := ParametroFromFormSimple(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Clave != "CLAVE" {
		t.Fatalf("unexpected clave: %q", got.Clave)
	}
	if got.Valor != "VALOR" {
		t.Fatalf("unexpected valor: %q", got.Valor)
	}
}

func TestParametroFromFormSimple_ParseFormError(t *testing.T) {
	req := httptest.NewRequest("POST", "/parametros/tabla", strings.NewReader("clave=%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err := ParametroFromFormSimple(req)
	if err == nil {
		t.Fatalf("expected parse form error")
	}
}

func contextWithRouteCtx(req *http.Request, rctx *chi.Context) context.Context {
	return context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
}
