package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/navigation"
	"PlantillaInstitucional/shared/sesionutil"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
	"github.com/issemym-repos/sesiones"
)

type mockNavigationPolicyService struct {
	mockPolicyService
	paths []string
}

func (m *mockNavigationPolicyService) NavigationPaths(ctx context.Context, profile string) ([]string, error) {
	return m.paths, nil
}

func TestNavigationMiddleware(t *testing.T) {
	logger := logs.NuevoLogger(&logs.Config{Nivel: "DEBUG"})
	policyMock := &mockNavigationPolicyService{
		paths: []string{"/dashboard", "/institucion/personas"},
	}

	appInstance := &app.App{
		Logger: logger,
	}
	appInstance.Services.Policy = policyMock

	regTipos := sesiones.NuevoRegistroTipos()
	_, _ = regTipos.RegistrarTipo(sesionutil.Usuario{})

	constructor := sesiones.NuevoConstructor(nil, &sesiones.ConfigSesion{}, regTipos, nil)
	middlewareFn := NavigationMiddleware(appInstance)

	t.Run("supplies paths for authenticated user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		sesion := constructor.NuevaSesion(req)
		u := uuid.Must(uuid.NewV4())
		sesion.UuidPersona = &u
		sesion.Perfil = sesiones.PerfilSuperAdmin
		_ = constructor.AgregarInstancia(sesion, sesionutil.Usuario{Perfil: "superadmin"})

		ctx := context.WithValue(req.Context(), sesiones.LlaveContexto, sesion)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()

		var checked bool
		handler := middlewareFn(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !navigation.Allowed(r.Context(), "/dashboard") {
				t.Error("expected /dashboard to be allowed in navigation")
			}
			if !navigation.Allowed(r.Context(), "/institucion/personas") {
				t.Error("expected /institucion/personas to be allowed in navigation")
			}
			if navigation.Allowed(r.Context(), "/configuracion/politicas") {
				t.Error("expected /configuracion/politicas to NOT be allowed in navigation")
			}
			checked = true
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)
		if !checked {
			t.Fatal("handler was not called")
		}
	})

	t.Run("fails closed for unauthenticated request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()

		var checked bool
		handler := middlewareFn(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if navigation.Allowed(r.Context(), "/dashboard") {
				t.Error("expected /dashboard to NOT be allowed for anonymous")
			}
			checked = true
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)
		if !checked {
			t.Fatal("handler was not called")
		}
	})
}
