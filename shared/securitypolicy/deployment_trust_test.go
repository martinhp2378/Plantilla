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
		EnforcementEnabled: true, DeploymentProfile: LocalDirect, AllowedSchemes: []string{"https"}, GlobalAccessPolicyID: "global-v1",
		Catalog: AuthoritativeCatalogSnapshot{Source: "catalog", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour), Entries: []TenantCatalogEntry{{
			Host: "tenant.example.test", TenantID: "tenant-a", Schema: "tenant_a", Lifecycle: TenantActive,
			InstitutionOwner: "institution", PlatformOwner: "platform", InstitutionApproved: true, PlatformApproved: true,
			Source: "catalog", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour), EvidenceApproved: true,
		}}},
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
