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
	DeploymentProfile      DeploymentTrustProfile
	NginxTrustedCIDRs      []netip.Prefix
	Catalog                AuthoritativeCatalogSnapshot
	CanonicalHosts         []CanonicalHostRule
	TrustedProxyCIDRs      []netip.Prefix
	ForwardedAuthorityMode ForwardedAuthorityMode
	AllowedSchemes         []string
	GlobalAccessPolicyID   string
}

func (p TenantPolicy) Validate() error {
	if !p.EnforcementEnabled {
		return errors.New(StartupPolicyIncomplete)
	}
	if p.DeploymentProfile != GCPCloudRunALB && p.DeploymentProfile != OnPremNginxSingleHop && p.DeploymentProfile != LocalDirect ||
		p.ForwardedAuthorityMode != "" || len(p.TrustedProxyCIDRs) != 0 ||
		(p.DeploymentProfile == OnPremNginxSingleHop && !validNginxCIDRs(p.NginxTrustedCIDRs)) ||
		(p.DeploymentProfile != OnPremNginxSingleHop && len(p.NginxTrustedCIDRs) != 0) {
		return errors.New(StartupProfileInvalid)
	}
	if len(p.AllowedSchemes) == 0 || p.GlobalAccessPolicyID == "" {
		return errors.New(StartupPolicyIncomplete)
	}
	if err := p.Catalog.Validate(); err != nil {
		return err
	}
	activeCatalog := map[[3]string]struct{}{}
	for _, entry := range p.Catalog.Entries {
		if entry.Lifecycle == TenantActive {
			activeCatalog[[3]string{entry.Host, entry.TenantID, entry.Schema}] = struct{}{}
		}
	}
	seenHosts, seenSchemas := map[string]struct{}{}, map[string]struct{}{}
	for _, rule := range p.CanonicalHosts {
		host := strings.TrimSuffix(strings.ToLower(rule.Host), ".")
		if host == "" || host != rule.Host || rule.TenantID == "" || rule.Schema == "public" || !schemaPattern.MatchString(rule.Schema) {
			return errors.New("tenant_policy_invalid_host_rule")
		}
		if _, ok := activeCatalog[[3]string{host, rule.TenantID, rule.Schema}]; rule.Active && !ok {
			return errors.New(StartupCatalogInvalid)
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
				return CanonicalHostRule{}, errors.New("tenant_inactive")
			}
			return rule, nil
		}
	}
	return CanonicalHostRule{}, errors.New("tenant_unknown")
}
