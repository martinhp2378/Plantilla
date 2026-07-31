package securitypolicy

import (
	"strings"
	"testing"
)

func TestTenantPolicyValidateAcceptsCanonicalActivePolicy(t *testing.T) {
	policy := TenantPolicy{
		EnforcementEnabled:   true,
		CanonicalHosts:       []CanonicalHostRule{{Host: "a.example.test", TenantID: "tenant-a", Schema: "tenant_a", Active: true}},
		AllowedSchemes:       []string{"https"},
		GlobalAccessPolicyID: "global-readers-v1",
	}
	if err := policy.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestTenantPolicyValidateRejectsUnsafeConfigurationWithoutLeakingValues(t *testing.T) {
	policy := TenantPolicy{
		EnforcementEnabled:     true,
		CanonicalHosts:         []CanonicalHostRule{{Host: "a.example.test", TenantID: "tenant-a", Schema: "public", Active: true}, {Host: "a.example.test", TenantID: "tenant-b", Schema: "tenant_b", Active: true}},
		ForwardedAuthorityMode: TrustedProxyOnly,
	}
	err := policy.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want unsafe policy rejection")
	}
	if strings.Contains(err.Error(), "tenant-a") || strings.Contains(err.Error(), "tenant-b") {
		t.Fatalf("Validate() leaked policy values: %v", err)
	}
}
