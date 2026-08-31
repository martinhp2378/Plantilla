package bd

import (
	"context"
	"errors"
	"strings"
	"testing"

	contextkeys "PlantillaInstitucional/shared/contextkey"
	"PlantillaInstitucional/shared/tenantcontext"
	"github.com/jackc/pgx/v5/pgxpool"
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

func TestConfigureTenantPoolFailsClosedWithoutValidTenantContext(t *testing.T) {
	config, err := pgxpool.ParseConfig("postgres://user:password@localhost:5432/test")
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	configureTenantPool(config)
	if config.BeforeAcquire == nil {
		t.Fatal("BeforeAcquire must be configured for direct pool consumers")
	}

	unsafeTenant := tenantcontext.WithTenant(context.Background(), tenantcontext.TenantContext{
		ID: "tenant-a", Schema: "public", CanonicalHost: "tenant.example.test",
	})
	for _, ctx := range []context.Context{context.Background(), unsafeTenant} {
		if config.BeforeAcquire(ctx, nil) {
			t.Fatal("BeforeAcquire accepted a direct operation without a valid tenant selection")
		}
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
