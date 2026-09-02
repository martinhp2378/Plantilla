package middleware_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appmiddleware "PlantillaInstitucional/app/middleware"
	reportes "PlantillaInstitucional/shared/reportes"
	"PlantillaInstitucional/shared/tenantcontext"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/sesiones"
)

func TestPrivateReportDownloadRequiresOwningAuthenticatedTenant(t *testing.T) {
	manager, err := archivos.NuevoManejador(&archivos.Config{Ambiente: archivos.AmbienteLocal, DirArchivosLocal: t.TempDir()})
	if err != nil {
		t.Fatalf("NuevoManejador() error = %v", err)
	}
	key := "reportes/tenant_one/report-id.pdf"
	if err := manager.ArchivoPrivado(key).Guardar(context.Background(), bytes.NewBufferString("private report")); err != nil {
		t.Fatalf("save private fixture: %v", err)
	}
	router := chi.NewRouter()
	router.With(appmiddleware.MiddlewareRequiereLogin).Get("/private-reports/*", reportes.PrivateDownloadHandler(manager))

	tests := []struct {
		name       string
		signedIn   bool
		schema     string
		wantStatus int
	}{
		{name: "anonymous denied", schema: "tenant_one", wantStatus: http.StatusFound},
		{name: "owning tenant allowed", signedIn: true, schema: "tenant_one", wantStatus: http.StatusOK},
		{name: "cross tenant denied", signedIn: true, schema: "tenant_two", wantStatus: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/private-reports/"+key, nil)
			session := &sesiones.Sesion{Perfil: sesiones.PerfilAnonimo}
			if tt.signedIn {
				id := uuid.Must(uuid.NewV4())
				session.UuidPersona = &id
			}
			ctx := context.WithValue(req.Context(), sesiones.LlaveContexto, session)
			ctx = tenantcontext.WithTenant(ctx, tenantcontext.TenantContext{ID: "tenant-id", Schema: tt.schema, CanonicalHost: tt.schema + ".example"})
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req.WithContext(ctx))
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK && recorder.Body.String() != "private report" {
				t.Fatalf("body = %q", recorder.Body.String())
			}
		})
	}
}
