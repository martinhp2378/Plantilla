package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/contextkey"
	"PlantillaInstitucional/shared/tenants"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTenantMiddleware(t *testing.T) {
	repo := tenants.NewStaticRepository()
	a := &app.App{
		Repositories: app.Repositories{
			Tenants: repo,
		},
	}
	mw := TenantMiddleware(a)

	tests := []struct {
		name           string
		host           string
		expectedSchema string
	}{
		{
			name:           "Valid tenant 1",
			host:           "tenant1.localhost",
			expectedSchema: "tenant1",
		},
		{
			name:           "Valid tenant 2",
			host:           "tenant2.localhost",
			expectedSchema: "tenant2",
		},
		{
			name:           "Unknown tenant defaults to public",
			host:           "unknown.localhost",
			expectedSchema: "public",
		},
		{
			name:           "No subdomain defaults to public",
			host:           "localhost",
			expectedSchema: "public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://"+tt.host, nil)
			rr := httptest.NewRecorder()

			var capturedSchema string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				val := r.Context().Value(contextkey.TenantSchemaKey)
				if val != nil {
					capturedSchema = val.(string)
				}
			})

			mw(next).ServeHTTP(rr, req)

			if capturedSchema != tt.expectedSchema {
				t.Errorf("expected schema %s, got %s", tt.expectedSchema, capturedSchema)
			}
		})
	}
}
