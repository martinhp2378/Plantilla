package securitypolicy

import (
	"errors"
	"net/netip"
	"strings"
	"time"
)

type DeploymentTrustProfile string

const (
	GCPCloudRunALB          DeploymentTrustProfile = "gcp_cloud_run_alb"
	OnPremNginxSingleHop    DeploymentTrustProfile = "onprem_nginx_single_hop"
	LocalDirect             DeploymentTrustProfile = "local_direct"
	StartupProfileInvalid                          = "startup_profile_invalid"
	StartupCatalogInvalid                          = "startup_catalog_invalid"
	StartupCatalogMissing                          = "startup_catalog_missing"
	StartupPolicyIncomplete                        = "startup_policy_incomplete"
)

type TenantLifecycle string

const (
	TenantProvisioning   TenantLifecycle = "provisioning"
	TenantActive         TenantLifecycle = "active"
	TenantSuspended      TenantLifecycle = "suspended"
	TenantDecommissioned TenantLifecycle = "decommissioned"
)

type TenantCatalogEntry struct {
	Host, TenantID, Schema, InstitutionOwner, PlatformOwner, Source, Version string
	Aliases                                                                  []string
	Lifecycle                                                                TenantLifecycle
	InstitutionApproved, PlatformApproved, EvidenceApproved                  bool
	ValidFrom, ValidUntil                                                    time.Time
}

type AuthoritativeCatalogSnapshot struct {
	Source, Version       string
	ValidFrom, ValidUntil time.Time
	Entries               []TenantCatalogEntry
}

func (c AuthoritativeCatalogSnapshot) Validate() error {
	if c.Source == "" || c.Version == "" || c.ValidFrom.IsZero() || c.ValidUntil.IsZero() || !c.ValidFrom.Before(c.ValidUntil) || len(c.Entries) == 0 {
		return errors.New(StartupCatalogMissing)
	}
	seenHosts, seenSchemas, seenTenants := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	for _, entry := range c.Entries {
		if !validCatalogHost(entry.Host) || entry.TenantID == "" || entry.Schema == "public" || !schemaPattern.MatchString(entry.Schema) ||
			entry.InstitutionOwner == "" || entry.PlatformOwner == "" || !entry.InstitutionApproved || !entry.PlatformApproved || !entry.EvidenceApproved ||
			entry.Source == "" || entry.Version == "" || entry.ValidFrom.IsZero() || entry.ValidUntil.IsZero() || !entry.ValidFrom.Before(entry.ValidUntil) || !validLifecycle(entry.Lifecycle) {
			return errors.New(StartupCatalogInvalid)
		}
		for _, host := range append([]string{entry.Host}, entry.Aliases...) {
			if !validCatalogHost(host) {
				return errors.New(StartupCatalogInvalid)
			}
			if _, exists := seenHosts[host]; exists {
				return errors.New(StartupCatalogInvalid)
			}
			seenHosts[host] = struct{}{}
		}
		if _, exists := seenSchemas[entry.Schema]; exists {
			return errors.New(StartupCatalogInvalid)
		}
		if _, exists := seenTenants[entry.TenantID]; exists {
			return errors.New(StartupCatalogInvalid)
		}
		seenSchemas[entry.Schema], seenTenants[entry.TenantID] = struct{}{}, struct{}{}
	}
	return nil
}

func validCatalogHost(host string) bool {
	if host == "" || host != strings.ToLower(host) || strings.ContainsAny(host, ":/@* ") || strings.HasSuffix(host, ".") || strings.Contains(host, "..") {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, rune := range label {
			if !(rune >= 'a' && rune <= 'z' || rune >= '0' && rune <= '9' || rune == '-') {
				return false
			}
		}
	}
	return true
}

func validLifecycle(lifecycle TenantLifecycle) bool {
	return lifecycle == TenantProvisioning || lifecycle == TenantActive || lifecycle == TenantSuspended || lifecycle == TenantDecommissioned
}

func validNginxCIDRs(prefixes []netip.Prefix) bool {
	if len(prefixes) == 0 {
		return false
	}
	for _, prefix := range prefixes {
		if !prefix.IsValid() {
			return false
		}
	}
	return true
}
