package securitypolicy

import (
	"strings"
	"testing"
)

func TestTenantPolicyValidateAcceptsCanonicalActivePolicy(t *testing.T) {
	policy := validStartupPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestTenantPolicyValidateRejectsUnsafeConfigurationWithoutLeakingValues(t *testing.T) {
	policy := validStartupPolicy()
	policy.Catalog.Entries[0].Schema = "public"
	err := policy.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want unsafe policy rejection")
	}
	if strings.Contains(err.Error(), "tenant-a") || strings.Contains(err.Error(), "tenant-b") {
		t.Fatalf("Validate() leaked policy values: %v", err)
	}
}

func TestTenantPolicyValidateRejectsActiveRoutesOutsideActiveCatalog(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*TenantPolicy)
	}{
		{"absent host", func(p *TenantPolicy) { p.CanonicalHosts[0].Host = "other.example.test" }},
		{"mismatched tenant", func(p *TenantPolicy) { p.CanonicalHosts[0].TenantID = "tenant-b" }},
		{"mismatched schema", func(p *TenantPolicy) { p.CanonicalHosts[0].Schema = "tenant_b" }},
		{"suspended catalog entry", func(p *TenantPolicy) { p.Catalog.Entries[0].Lifecycle = TenantSuspended }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := validStartupPolicy()
			policy.CanonicalHosts = []CanonicalHostRule{{Host: "tenant.example.test", TenantID: "tenant-a", Schema: "tenant_a", Active: true}}
			tc.mutate(&policy)
			if err := policy.Validate(); err == nil || err.Error() != StartupCatalogInvalid {
				t.Fatalf("Validate() = %v, want %s", err, StartupCatalogInvalid)
			}
		})
	}
}
