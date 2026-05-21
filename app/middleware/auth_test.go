package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"PlantillaInstitucional/shared/contextkey"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/sesiones"
)

func TestMiddlewareRequiereLogin(t *testing.T) {
	tests := []struct {
		name           string
		signedIn       bool
		expectedStatus int
	}{
		{
			name:           "Redirects if not signed in",
			signedIn:       false,
			expectedStatus: http.StatusFound,
		},
		{
			name:           "Allows if signed in",
			signedIn:       true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := MiddlewareRequiereLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			
			// Mock session in context
			ses := &sesiones.Sesion{}
			if tt.signedIn {
				// Assuming PerfilAnonimo is unauthenticated, any other is authenticated
				ses.Perfil = "superadmin"
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

func TestAuditContextMiddleware(t *testing.T) {
	handler := AuditContextMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(contextkey.AuditUUIDKey)
		if val == nil {
			t.Error("expected AuditUUIDKey in context, got nil")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	
	// Mock session with UUID
	u := uuid.Must(uuid.NewV4())
	ses := &sesiones.Sesion{
		UuidPersona: &u,
	}
	ctx := context.WithValue(req.Context(), sesiones.LlaveContexto, ses)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
