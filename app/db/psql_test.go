package bd

import (
	"context"
	"errors"
	"strings"
	"testing"

	contextkeys "PlantillaInstitucional/shared/contextkey"
	"PlantillaInstitucional/shared/tenantcontext"
)

func TestTenantSchemaRequiresImmutableTenantContext(t *testing.T) {
	valid := tenantcontext.WithTenant(context.Background(), tenantcontext.TenantContext{
		ID: "tenant-a", Schema: "tenant_a", CanonicalHost: "tenant.example.test",
	})
	legacy := context.WithValue(context.Background(), contextkeys.TenantSchemaKey, "tenant_a")

	for _, tc := range []struct {
		name    string
		ctx     context.Context
		want    string
		wantErr error
	}{
		{"missing context fails closed", context.Background(), "", tenantcontext.ErrMissingTenantContext},
		{"legacy context is incompatible", legacy, "", tenantcontext.ErrMissingTenantContext},
		{"validated immutable context selects tenant schema", valid, "tenant_a", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getTenantSchema(tc.ctx)
			if !errors.Is(err, tc.wantErr) || got != tc.want {
				t.Fatalf("getTenantSchema() = %q, %v; want %q, %v", got, err, tc.want, tc.wantErr)
			}
		})
	}
}


func TestBuildTenantSearchPathRejectsPublicFallback(t *testing.T) {
	for _, schema := range []string{"", "public"} {
		if _, err := buildTenantSearchPath(schema); err == nil {
			t.Fatalf("buildTenantSearchPath(%q) succeeded", schema)
		}
	}

	searchPath, err := buildTenantSearchPath("tenant_a")
	if err != nil || strings.Contains(searchPath, "public") {
		t.Fatalf("buildTenantSearchPath() = %q, %v; want tenant-only path without public", searchPath, err)
	}
}
