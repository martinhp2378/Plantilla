package securitypolicy

import (
	"errors"
	"net/netip"
	"regexp"
	"strings"
)

var schemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type ForwardedAuthorityMode string

const (
	ForwardedAuthorityDisabled ForwardedAuthorityMode = "disabled"
	TrustedProxyOnly           ForwardedAuthorityMode = "trusted_proxy_only"
)

type CanonicalHostRule struct {
	Host     string
	TenantID string
	Schema   string
	Active   bool
}
type TenantPolicy struct {
	EnforcementEnabled     bool
	CanonicalHosts         []CanonicalHostRule
	TrustedProxyCIDRs      []netip.Prefix
	ForwardedAuthorityMode ForwardedAuthorityMode
	AllowedSchemes         []string
	GlobalAccessPolicyID   string
}

func (p TenantPolicy) Validate() error {
	if !p.EnforcementEnabled {
		return errors.New("tenant_policy_enforcement_disabled")
	}
	if len(p.CanonicalHosts) == 0 {
		return errors.New("tenant_policy_missing_canonical_hosts")
	}
	if len(p.AllowedSchemes) == 0 {
		return errors.New("tenant_policy_missing_allowed_schemes")
	}
	if p.GlobalAccessPolicyID == "" {
		return errors.New("tenant_policy_missing_global_access_policy")
	}
	if p.ForwardedAuthorityMode == "" {
		p.ForwardedAuthorityMode = ForwardedAuthorityDisabled
	}
	if p.ForwardedAuthorityMode != ForwardedAuthorityDisabled && p.ForwardedAuthorityMode != TrustedProxyOnly {
		return errors.New("tenant_policy_invalid_forwarded_authority_mode")
	}
	if p.ForwardedAuthorityMode == TrustedProxyOnly && len(p.TrustedProxyCIDRs) == 0 {
		return errors.New("tenant_policy_missing_trusted_proxy_cidrs")
	}
	seenHosts, seenSchemas := map[string]struct{}{}, map[string]struct{}{}
	for _, rule := range p.CanonicalHosts {
		host := strings.TrimSuffix(strings.ToLower(rule.Host), ".")
		if host == "" || host != rule.Host || rule.TenantID == "" || rule.Schema == "public" || !schemaPattern.MatchString(rule.Schema) {
			return errors.New("tenant_policy_invalid_host_rule")
		}
		if _, ok := seenHosts[host]; ok {
			return errors.New("tenant_policy_duplicate_host")
		}
		if _, ok := seenSchemas[rule.Schema]; ok {
			return errors.New("tenant_policy_duplicate_schema")
		}
		seenHosts[host], seenSchemas[rule.Schema] = struct{}{}, struct{}{}
	}
	return nil
}
func (p TenantPolicy) RuleForHost(host string) (CanonicalHostRule, error) {
	if err := p.Validate(); err != nil {
		return CanonicalHostRule{}, err
	}
	for _, rule := range p.CanonicalHosts {
		if rule.Host == host {
			if !rule.Active {
				return CanonicalHostRule{}, errors.New("tenant_policy_inactive_tenant")
			}
			return rule, nil
		}
	}
	return CanonicalHostRule{}, errors.New("tenant_policy_unknown_host")
}
