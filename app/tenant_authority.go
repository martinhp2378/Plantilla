package app

import (
	"PlantillaInstitucional/shared/entorno"
	"PlantillaInstitucional/shared/securitypolicy"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"strings"
	"time"
)

type tenantAuthorityConfig struct {
	DeploymentProfile    securitypolicy.DeploymentTrustProfile       `json:"deployment_profile"`
	NginxTrustedCIDRs    []string                                    `json:"nginx_trusted_cidrs,omitempty"`
	AllowedSchemes       []string                                    `json:"allowed_schemes"`
	GlobalAccessPolicyID string                                      `json:"global_access_policy_id"`
	Catalog              securitypolicy.AuthoritativeCatalogSnapshot `json:"catalog"`
	CanonicalHosts       []securitypolicy.CanonicalHostRule          `json:"canonical_hosts"`
}

var timeNow = time.Now

func tenantPolicyFromEnvironment(env *entorno.Env, clock func() time.Time) (securitypolicy.TenantPolicy, error) {
	if env == nil || clock == nil {
		return securitypolicy.TenantPolicy{}, errors.New(securitypolicy.StartupProfileInvalid)
	}

	environmentClass, err := tenantEnvironmentClass(env.Ambiente)
	if err != nil {
		return securitypolicy.TenantPolicy{}, err
	}
	if strings.TrimSpace(env.TenantAuthorityConfig) == "" {
		if environmentClass == securitypolicy.EnvironmentLocal {
			return localTenantPolicy(clock), nil
		}
		return securitypolicy.TenantPolicy{}, errors.New(securitypolicy.StartupCatalogMissing)
	}

	var config tenantAuthorityConfig
	decoder := json.NewDecoder(strings.NewReader(env.TenantAuthorityConfig))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return securitypolicy.TenantPolicy{}, fmt.Errorf("%s: invalid JSON", securitypolicy.StartupCatalogInvalid)
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return securitypolicy.TenantPolicy{}, fmt.Errorf("%s: invalid JSON", securitypolicy.StartupCatalogInvalid)
	}

	trustedCIDRs, err := parseCIDRs(config.NginxTrustedCIDRs)
	if err != nil {
		return securitypolicy.TenantPolicy{}, errors.New(securitypolicy.StartupProfileInvalid)
	}
	policy := securitypolicy.TenantPolicy{
		EnforcementEnabled:   true,
		DeploymentProfile:    config.DeploymentProfile,
		EnvironmentClass:     environmentClass,
		NginxTrustedCIDRs:    trustedCIDRs,
		Catalog:              config.Catalog,
		CanonicalHosts:       config.CanonicalHosts,
		AllowedSchemes:       config.AllowedSchemes,
		GlobalAccessPolicyID: config.GlobalAccessPolicyID,
		Clock:                clock,
	}
	if err := policy.Validate(); err != nil {
		return securitypolicy.TenantPolicy{}, err
	}
	return policy, nil
}

func tenantEnvironmentClass(environment string) (securitypolicy.EnvironmentClass, error) {
	switch environment {
	case entorno.AmbienteLocal:
		return securitypolicy.EnvironmentLocal, nil
	case entorno.AmbienteDesarrollo:
		return securitypolicy.EnvironmentNonProduction, nil
	case entorno.AmbienteProduccion:
		return securitypolicy.EnvironmentProduction, nil
	default:
		return "", errors.New(securitypolicy.StartupProfileInvalid)
	}
}

func localTenantPolicy(clock func() time.Time) securitypolicy.TenantPolicy {
	now := clock().UTC()
	validUntil := now.AddDate(100, 0, 0)
	return securitypolicy.TenantPolicy{
		EnforcementEnabled: true, DeploymentProfile: securitypolicy.LocalDirect, EnvironmentClass: securitypolicy.EnvironmentLocal,
		CanonicalHosts: []securitypolicy.CanonicalHostRule{{Host: "localhost", TenantID: "local", Schema: "local", Active: true}},
		AllowedSchemes: []string{"http", "https"}, GlobalAccessPolicyID: "global-readers-v1", Clock: clock,
		Catalog: securitypolicy.AuthoritativeCatalogSnapshot{Source: "local-fixture", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: validUntil, Entries: []securitypolicy.TenantCatalogEntry{{Host: "localhost", TenantID: "local", Schema: "local", Lifecycle: securitypolicy.TenantActive, InstitutionOwner: "local", PlatformOwner: "local", InstitutionApproved: true, PlatformApproved: true, EvidenceApproved: true, Source: "local-fixture", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: validUntil}}},
	}
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("unexpected trailing JSON")
	}
	return nil
}

func parseCIDRs(values []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}
