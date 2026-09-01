package app

import (
	"PlantillaInstitucional/shared/entorno"
	"PlantillaInstitucional/shared/securitypolicy"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewApp(t *testing.T) {
	// We use a dummy env. NewPool will fail to connect but we can verify the error
	// or mock if we had a mocking library. For now, let's just ensure the logic flows.
	env := &entorno.Env{
		Ambiente:    entorno.AmbienteLocal,
		PqWriteHost: "localhost",
		PqWritePort: "5432",
		PqReadHost:  "localhost",
		PqReadPort:  "5432",
	}

	ctx := context.Background()
	a, err := NewApp(ctx, env, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("expected nil error for simple constructor, got %v", err)
	}
	if err := a.TenantPolicy.Validate(); err != nil {
		t.Errorf("default tenant policy must be valid, got %v", err)
	}
}

func TestNewAppLoadsEnvironmentAppropriateTenantAuthority(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	originalClock := timeNow
	timeNow = func() time.Time { return now }
	t.Cleanup(func() { timeNow = originalClock })

	validConfig := productionTenantAuthorityConfig(t, now)
	tests := []struct {
		name        string
		env         *entorno.Env
		wantErr     string
		wantProfile securitypolicy.DeploymentTrustProfile
		wantSource  string
	}{
		{name: "local uses intentional fixture", env: &entorno.Env{Ambiente: entorno.AmbienteLocal}, wantProfile: securitypolicy.LocalDirect, wantSource: "local-fixture"},
		{name: "production requires authority configuration", env: &entorno.Env{Ambiente: entorno.AmbienteProduccion}, wantErr: securitypolicy.StartupCatalogMissing},
		{name: "production rejects invalid authority configuration", env: &entorno.Env{Ambiente: entorno.AmbienteProduccion, TenantAuthorityConfig: `{"deployment_profile":`}, wantErr: securitypolicy.StartupCatalogInvalid},
		{name: "production accepts validated authority configuration", env: &entorno.Env{Ambiente: entorno.AmbienteProduccion, TenantAuthorityConfig: validConfig}, wantProfile: securitypolicy.GCPCloudRunALB, wantSource: "deployment-catalog"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewApp(context.Background(), tt.env, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("NewApp() error = %v, want %s", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewApp() error = %v, want nil", err)
			}
			if a.TenantPolicy.DeploymentProfile != tt.wantProfile || a.TenantPolicy.Catalog.Source != tt.wantSource {
				t.Fatalf("profile/source = %q/%q, want %q/%q", a.TenantPolicy.DeploymentProfile, a.TenantPolicy.Catalog.Source, tt.wantProfile, tt.wantSource)
			}
		})
	}
}

func productionTenantAuthorityConfig(t *testing.T, now time.Time) string {
	t.Helper()
	config := tenantAuthorityConfig{
		DeploymentProfile: securitypolicy.GCPCloudRunALB,
		AllowedSchemes:    []string{"https"}, GlobalAccessPolicyID: "global-readers-v1",
		CanonicalHosts: []securitypolicy.CanonicalHostRule{{Host: "tenant.example.test", TenantID: "tenant-a", Schema: "tenant_a", Active: true}},
		Catalog: securitypolicy.AuthoritativeCatalogSnapshot{Source: "deployment-catalog", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour), Entries: []securitypolicy.TenantCatalogEntry{{
			Host: "tenant.example.test", TenantID: "tenant-a", Schema: "tenant_a", Lifecycle: securitypolicy.TenantActive,
			InstitutionOwner: "institution", PlatformOwner: "platform", InstitutionApproved: true, PlatformApproved: true, EvidenceApproved: true,
			Source: "deployment-catalog", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour),
		}}},
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}
