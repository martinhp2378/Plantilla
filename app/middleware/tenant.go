package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/securitypolicy"
	"PlantillaInstitucional/shared/tenantcontext"
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
	if policy.ForwardedAuthorityMode == securitypolicy.TrustedProxyOnly && trustedRemote(r.RemoteAddr, policy.TrustedProxyCIDRs) {
		if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
			host, err := canonicalHost(forwarded)
			return host, tenantcontext.TrustedForwardedHost, err
		}
	}
	host, err := canonicalHost(r.Host)
	return host, tenantcontext.DirectHost, err
}

func allowedScheme(r *http.Request, allowed []string) bool {
	scheme := r.URL.Scheme
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
			if !(rune >= 'a' && rune <= 'z' || rune >= '0' && rune <= '9' || rune == '-') {
				return "", http.ErrNotSupported
			}
		}
	}
	return host, nil
}
