package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"PlantillaInstitucional/app"
	"PlantillaInstitucional/configuracion/policy"
	"PlantillaInstitucional/pagination"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
	"github.com/issemym-repos/sesiones"
)

type mockPolicyService struct {
	allowed bool
	err     error
}

func (m *mockPolicyService) IsRouteAllowed(ctx context.Context, profile string, path string) (bool, error) {
	return m.allowed, m.err
}

func (m *mockPolicyService) IsProfileActive(ctx context.Context, profile string) (bool, error) {
	return true, nil
}

func (m *mockPolicyService) ListarPoliticas(ctx context.Context, criterios pagination.CriteriosPagina[policy.Politica]) (pagination.CriteriosPagina[policy.Politica], error) {
	return criterios, nil
}

func (m *mockPolicyService) ObtenerPolitica(ctx context.Context, nombre string) (policy.Politica, error) {
	return policy.Politica{}, nil
}

func (m *mockPolicyService) CrearPolitica(ctx context.Context, p policy.Politica) error {
	return nil
}

func (m *mockPolicyService) ActualizarPolitica(ctx context.Context, p policy.Politica, nombreOriginal string) error {
	return nil
}

func (m *mockPolicyService) EliminarPolitica(ctx context.Context, nombre string) error {
	return nil
}

func (m *mockPolicyService) CatalogosPolitica(ctx context.Context) ([]string, []string, []string, error) {
	return nil, nil, nil, nil
}

func TestPolicyMiddleware(t *testing.T) {
	logger := logs.NuevoLogger(&logs.Config{Nivel: "DEBUG"})
	
	tests := []struct {
		name           string
		path           string
		signedIn       bool
		profile        string
		allowed        bool
		isHTMX         bool
		expectedStatus int
	}{
		{
			name:           "Excluded route allowed without session",
			path:           "/health",
			signedIn:       false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Protected route allowed if policy permits",
			path:           "/dashboard",
			signedIn:       true,
			profile:        "admin",
			allowed:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Protected route denied if policy forbids",
			path:           "/dashboard",
			signedIn:       true,
			profile:        "user",
			allowed:        false,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "HTMX request denied with 403",
			path:           "/dashboard",
			signedIn:       true,
			profile:        "user",
			isHTMX:         true,
			allowed:        false,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock app
			mockSvc := &mockPolicyService{allowed: tt.allowed}
			a := &app.App{
				Logger: logger,
				Services: app.Services{
					Policy: mockSvc,
				},
			}
			
			handler := PolicyMiddleware(a)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.isHTMX {
				req.Header.Set("HX-Request", "true")
			}

			// Mock session
			ses := &sesiones.Sesion{}
			if tt.signedIn {
				ses.Perfil = sesiones.Perfil(tt.profile)
				u := uuid.Must(uuid.NewV4())
				ses.UuidPersona = &u
			} else {
				ses.Perfil = sesiones.PerfilAnonimo
			}
			ctx := context.WithValue(req.Context(), sesiones.LlaveContexto, ses)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
