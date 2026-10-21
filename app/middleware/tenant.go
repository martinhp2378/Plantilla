package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/securitypolicy"
	"PlantillaInstitucional/shared/tenantcontext"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

func TenantMiddleware(a *app.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, source, err := requestAuthority(r, a.TenantPolicy)
			if err != nil || !allowedScheme(r, a.TenantPolicy.AllowedSchemes) {
				http.Error(w, "misdirected request", http.StatusMisdirectedRequest)
				return
			}
			rule, err := a.TenantPolicy.RuleForHost(host)
			if err != nil {
				http.Error(w, "misdirected request", http.StatusMisdirectedRequest)
				return
			}
			ctx := tenantcontext.WithTenant(r.Context(), tenantcontext.TenantContext{
				ID: rule.TenantID, Schema: rule.Schema, CanonicalHost: host, Source: source,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requestAuthority(r *http.Request, policy securitypolicy.TenantPolicy) (string, tenantcontext.Source, error) {
	if err := policy.Validate(); err != nil {
		return "", "", err
	}
	if r.Header.Get("Forwarded") != "" || strings.Contains(r.Header.Get("X-Forwarded-For"), ",") {
		return "", "", errors.New("authority_multihop")
	}
	switch policy.DeploymentProfile {
	case securitypolicy.GCPCloudRunALB:
		if r.Header.Get("X-Forwarded-Host") != "" {
			return "", "", errors.New("authority_conflicting_headers")
		}
		if err := validateForwardedProto(r, policy.AllowedSchemes, false); err != nil {
			return "", "", err
		}
	case securitypolicy.OnPremNginxSingleHop:
		if !trustedRemote(r.RemoteAddr, policy.NginxTrustedCIDRs) {
			return "", "", errors.New("authority_untrusted_peer")
		}
		if err := validateForwardedProto(r, policy.AllowedSchemes, true); err != nil {
			return "", "", err
		}
		forwardedHost, err := singleForwardedHost(r)
		if err != nil {
			return "", "", err
		}
		if _, err := canonicalHost(r.Host); err != nil {
			return "", "", errors.New("authority_invalid")
		}
		return forwardedHost, tenantcontext.TrustedForwardedHost, nil
	case securitypolicy.LocalDirect:
		if hasForwardedAuthorityHeaders(r) || r.Header.Get("X-Forwarded-For") != "" {
			return "", "", errors.New("authority_conflicting_headers")
		}
	default:
		return "", "", errors.New("authority_invalid")
	}
	host, err := canonicalHost(r.Host)
	if err != nil {
		return "", "", errors.New("authority_invalid")
	}
	return host, tenantcontext.DirectHost, nil
}

func hasForwardedAuthorityHeaders(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-Host") != "" || r.Header.Get("X-Forwarded-Proto") != ""
}

func validateForwardedProto(r *http.Request, allowed []string, required bool) error {
	values := r.Header.Values("X-Forwarded-Proto")
	if len(values) == 0 && !required {
		return nil
	}
	if len(values) != 1 || values[0] == "" {
		return errors.New("authority_conflicting_headers")
	}
	for _, scheme := range allowed {
		if strings.EqualFold(values[0], scheme) {
			return nil
		}
	}
	return errors.New("authority_invalid")
}

func singleForwardedHost(r *http.Request) (string, error) {
	values := r.Header.Values("X-Forwarded-Host")
	if len(values) == 0 || values[0] == "" {
		return "", errors.New("authority_invalid")
	}
	if len(values) != 1 {
		return "", errors.New("authority_conflicting_headers")
	}
	if strings.Contains(values[0], ",") {
		return "", errors.New("authority_multihop")
	}
	host, err := canonicalHost(values[0])
	if err != nil {
		return "", errors.New("authority_invalid")
	}
	return host, nil
}

func allowedScheme(r *http.Request, allowed []string) bool {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = r.URL.Scheme
	}
	if scheme == "" {
		scheme = "http"
		if r.TLS != nil {
			scheme = "https"
		}
	}
	for _, candidate := range allowed {
		if strings.EqualFold(scheme, candidate) {
			return true
		}
	}
	return false
}

func trustedRemote(remote string, prefixes []netip.Prefix) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		return false
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	for _, prefix := range prefixes {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func canonicalHost(authority string) (string, error) {
	if authority == "" || strings.ContainsAny(authority, "@, ") || strings.Contains(authority, "/") {
		return "", http.ErrNotSupported
	}
	host := authority
	if strings.Contains(authority, ":") {
		var err error
		host, _, err = net.SplitHostPort(authority)
		if err != nil {
			return "", err
		}
	}
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "" || net.ParseIP(host) != nil || strings.Contains(host, "..") {
		return "", http.ErrNotSupported
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return "", http.ErrNotSupported
		}
		for _, rune := range label {
			if (rune < 'a' || rune > 'z') && (rune < '0' || rune > '9') && rune != '-' {
				return "", http.ErrNotSupported
			}
		}
	}
	return host, nil
}
