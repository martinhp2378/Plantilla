package securitypolicy

import (
	"net/netip"
	"strings"
	"testing"
	"time"
)

func validStartupPolicy() TenantPolicy {
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	return TenantPolicy{
		EnforcementEnabled: true, DeploymentProfile: LocalDirect, EnvironmentClass: EnvironmentLocal, AllowedSchemes: []string{"https"}, GlobalAccessPolicyID: "global-v1",
		Clock: func() time.Time { return now },
		Catalog: AuthoritativeCatalogSnapshot{Source: "catalog", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour), Entries: []TenantCatalogEntry{{
			Host: "tenant.example.test", TenantID: "tenant-a", Schema: "tenant_a", Lifecycle: TenantActive,
			InstitutionOwner: "institution", PlatformOwner: "platform", InstitutionApproved: true, PlatformApproved: true,
			Source: "catalog", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour), EvidenceApproved: true,
		}}},
	}
}

func TestTenantPolicyValidateRejectsCatalogOutsideValidityWindow(t *testing.T) {
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*AuthoritativeCatalogSnapshot)
	}{
		{"expired snapshot", func(c *AuthoritativeCatalogSnapshot) { c.ValidUntil = now }},
		{"not-yet-valid snapshot", func(c *AuthoritativeCatalogSnapshot) { c.ValidFrom = now.Add(time.Minute) }},
		{"expired entry", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].ValidUntil = now }},
		{"not-yet-valid entry", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].ValidFrom = now.Add(time.Minute) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := validStartupPolicy()
			tt.mutate(&policy.Catalog)
			if err := policy.ValidateAt(now); err == nil || err.Error() != StartupCatalogInvalid {
				t.Fatalf("ValidateAt() = %v, want %s", err, StartupCatalogInvalid)
			}
		})
	}
}

func TestTenantPolicyValidateRejectsUnsafeStartupProfiles(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		mutate     func(*TenantPolicy)
	}{
		{"absent", StartupProfileInvalid, func(p *TenantPolicy) { p.DeploymentProfile = "" }},
		{"unknown", StartupProfileInvalid, func(p *TenantPolicy) { p.DeploymentProfile = "unknown" }},
		{"legacy selector", StartupProfileInvalid, func(p *TenantPolicy) { p.ForwardedAuthorityMode = TrustedProxyOnly }},
		{"mixed profile", StartupProfileInvalid, func(p *TenantPolicy) {
			p.DeploymentProfile = LocalDirect
			p.NginxTrustedCIDRs = []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}
		}},
		{"invalid nginx cidr", StartupProfileInvalid, func(p *TenantPolicy) {
			p.DeploymentProfile = OnPremNginxSingleHop
			p.NginxTrustedCIDRs = []netip.Prefix{{}}
		}},
		{"missing catalog", StartupCatalogMissing, func(p *TenantPolicy) { p.Catalog = AuthoritativeCatalogSnapshot{} }},
		{"missing policy", StartupPolicyIncomplete, func(p *TenantPolicy) { p.GlobalAccessPolicyID = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := validStartupPolicy()
			tc.mutate(&policy)
			err := policy.Validate()
			if err == nil || err.Error() != tc.want {
				t.Fatalf("Validate() = %v, want %s", err, tc.want)
			}
			if strings.Contains(err.Error(), "tenant-a") {
				t.Fatalf("diagnosis leaked tenant: %v", err)
			}
		})
	}
}

func TestTenantPolicyValidateRestrictsLocalDirectToExplicitNonProductionEnvironment(t *testing.T) {
	for _, tc := range []struct {
		name    string
		env     EnvironmentClass
		wantErr string
	}{
		{"explicit local is accepted", EnvironmentLocal, ""},
		{"explicit non-production is accepted", EnvironmentNonProduction, ""},
		{"production is rejected", EnvironmentProduction, StartupProfileInvalid},
		{"missing environment is rejected", "", StartupProfileInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := validStartupPolicy()
			policy.EnvironmentClass = tc.env
			err := policy.Validate()
			if tc.wantErr == "" && err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
			if tc.wantErr != "" && (err == nil || err.Error() != tc.wantErr) {
				t.Fatalf("Validate() = %v, want %s", err, tc.wantErr)
			}
		})
	}
}

func TestTenantPolicyValidateRejectsUnsafeCatalog(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*AuthoritativeCatalogSnapshot)
	}{
		{"duplicate host", func(c *AuthoritativeCatalogSnapshot) { c.Entries = append(c.Entries, c.Entries[0]) }},
		{"duplicate alias", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].Aliases = []string{"tenant.example.test"} }},
		{"duplicate schema", func(c *AuthoritativeCatalogSnapshot) {
			e := c.Entries[0]
			e.Host, e.TenantID = "other.example.test", "tenant-b"
			c.Entries = append(c.Entries, e)
		}},
		{"duplicate tenant", func(c *AuthoritativeCatalogSnapshot) {
			e := c.Entries[0]
			e.Host, e.Schema = "other.example.test", "tenant_b"
			c.Entries = append(c.Entries, e)
		}},
		{"invalid canonical host", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].Host = "Tenant.example.test." }},
		{"public schema", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].Schema = "public" }},
		{"missing ownership", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].PlatformOwner = "" }},
		{"missing approval", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].InstitutionApproved = false }},
		{"missing source", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].Source = "" }},
		{"missing version", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].Version = "" }},
		{"invalid window", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].ValidUntil = c.Entries[0].ValidFrom }},
		{"invalid lifecycle", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].Lifecycle = "invalid" }},
		{"contradictory evidence", func(c *AuthoritativeCatalogSnapshot) { c.Entries[0].EvidenceApproved = false }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy := validStartupPolicy()
			tc.mutate(&policy.Catalog)
			err := policy.Validate()
			if err == nil || err.Error() != StartupCatalogInvalid {
				t.Fatalf("Validate() = %v, want %s", err, StartupCatalogInvalid)
			}
			if strings.Contains(err.Error(), "tenant-a") {
				t.Fatalf("diagnosis leaked tenant: %v", err)
			}
		})
	}
}
