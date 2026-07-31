package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/securitypolicy"
	"PlantillaInstitucional/shared/tenantcontext"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

func tenantPolicy() securitypolicy.TenantPolicy {
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	return securitypolicy.TenantPolicy{EnforcementEnabled: true, DeploymentProfile: securitypolicy.LocalDirect, CanonicalHosts: []securitypolicy.CanonicalHostRule{{Host: "tenant1.localhost", TenantID: "tenant1", Schema: "tenant1", Active: true}}, AllowedSchemes: []string{"https"}, GlobalAccessPolicyID: "global-readers-v1", Catalog: securitypolicy.AuthoritativeCatalogSnapshot{Source: "test", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour), Entries: []securitypolicy.TenantCatalogEntry{{Host: "tenant1.localhost", TenantID: "tenant1", Schema: "tenant1", Lifecycle: securitypolicy.TenantActive, InstitutionOwner: "institution", PlatformOwner: "platform", InstitutionApproved: true, PlatformApproved: true, EvidenceApproved: true, Source: "test", Version: "v1", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour)}}}}
}
func TestTenantMiddlewareResolvesOnlyCanonicalActiveAuthority(t *testing.T) {
	a, called := &app.App{TenantPolicy: tenantPolicy()}, false
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		tenant, err := tenantcontext.TenantFromContext(r.Context())
		if err != nil || tenant.Schema != "tenant1" || tenant.CanonicalHost != "tenant1.localhost" {
			t.Fatalf("tenant context/error = %#v/%v", tenant, err)
		}
	})
	rr := httptest.NewRecorder()
	TenantMiddleware(a)(next).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "https://tenant1.localhost.:8443", nil))
	if rr.Code != http.StatusOK || !called {
		t.Fatalf("status/called = %d/%t, want 200/true", rr.Code, called)
	}
}
func TestTenantMiddlewareDeniesDisallowedSchemeBeforeNextHandler(t *testing.T) {
	called := false
	rr := httptest.NewRecorder()
	TenantMiddleware(&app.App{TenantPolicy: tenantPolicy()})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://tenant1.localhost", nil))
	if rr.Code != http.StatusMisdirectedRequest || called {
		t.Fatalf("status/called = %d/%t, want 421/false", rr.Code, called)
	}
}
func TestTenantMiddlewareDeniesLegacyForwardedAuthoritySelector(t *testing.T) {
	policy := tenantPolicy()
	policy.ForwardedAuthorityMode, policy.TrustedProxyCIDRs = securitypolicy.TrustedProxyOnly, []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
	req := httptest.NewRequest(http.MethodGet, "https://edge.invalid", nil)
	req.RemoteAddr, req.Header = "127.0.0.1:1234", http.Header{"X-Forwarded-Host": []string{"tenant1.localhost"}}
	rr := httptest.NewRecorder()
	TenantMiddleware(&app.App{TenantPolicy: policy})(next).ServeHTTP(rr, req)
	if rr.Code != http.StatusMisdirectedRequest || called {
		t.Fatalf("status/called = %d/%t", rr.Code, called)
	}
}
func TestTenantMiddlewareDeniesUntrustedAuthorityBeforeNextHandler(t *testing.T) {
	called := false
	req := httptest.NewRequest(http.MethodGet, "https://unknown.localhost", nil)
	req.Header.Set("X-Forwarded-Host", "tenant1.localhost")
	rr := httptest.NewRecorder()
	TenantMiddleware(&app.App{TenantPolicy: tenantPolicy()})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(rr, req)
	if rr.Code != http.StatusMisdirectedRequest || called {
		t.Fatalf("status/called = %d/%t, want 421/false", rr.Code, called)
	}
}

func TestRequestAuthorityEnforcesProfileHeaderContracts(t *testing.T) {
	tests := []struct {
		name    string
		profile securitypolicy.DeploymentTrustProfile
		remote  string
		host    string
		headers http.Header
		want    string
		code    string
	}{
		{"gcp uses host", securitypolicy.GCPCloudRunALB, "198.51.100.1:443", "tenant1.localhost", nil, "tenant1.localhost", ""},
		{"gcp rejects xfh", securitypolicy.GCPCloudRunALB, "198.51.100.1:443", "tenant1.localhost", http.Header{"X-Forwarded-Host": {"tenant1.localhost"}}, "", "authority_conflicting_headers"},
		{"gcp rejects forwarded", securitypolicy.GCPCloudRunALB, "198.51.100.1:443", "tenant1.localhost", http.Header{"Forwarded": {"host=tenant1.localhost"}}, "", "authority_multihop"},
		{"gcp ignores single xff", securitypolicy.GCPCloudRunALB, "198.51.100.1:443", "tenant1.localhost", http.Header{"X-Forwarded-For": {"203.0.113.1"}}, "tenant1.localhost", ""},
		{"gcp rejects xff multihop", securitypolicy.GCPCloudRunALB, "198.51.100.1:443", "tenant1.localhost", http.Header{"X-Forwarded-For": {"203.0.113.1, 198.51.100.1"}}, "", "authority_multihop"},
		{"nginx uses exactly one trusted hop", securitypolicy.OnPremNginxSingleHop, "10.0.0.2:443", "upstream.internal", http.Header{"X-Forwarded-Host": {"tenant1.localhost"}, "X-Forwarded-Proto": {"https"}}, "tenant1.localhost", ""},
		{"nginx rejects untrusted peer", securitypolicy.OnPremNginxSingleHop, "198.51.100.1:443", "upstream.internal", http.Header{"X-Forwarded-Host": {"tenant1.localhost"}, "X-Forwarded-Proto": {"https"}}, "", "authority_untrusted_peer"},
		{"nginx rejects missing xfh", securitypolicy.OnPremNginxSingleHop, "10.0.0.2:443", "upstream.internal", http.Header{"X-Forwarded-Proto": {"https"}}, "", "authority_invalid"},
		{"nginx rejects duplicate xfh", securitypolicy.OnPremNginxSingleHop, "10.0.0.2:443", "upstream.internal", http.Header{"X-Forwarded-Host": {"tenant1.localhost", "tenant1.localhost"}, "X-Forwarded-Proto": {"https"}}, "", "authority_conflicting_headers"},
		{"nginx rejects duplicate proto", securitypolicy.OnPremNginxSingleHop, "10.0.0.2:443", "upstream.internal", http.Header{"X-Forwarded-Host": {"tenant1.localhost"}, "X-Forwarded-Proto": {"https", "https"}}, "", "authority_conflicting_headers"},
		{"nginx rejects multihop xfh", securitypolicy.OnPremNginxSingleHop, "10.0.0.2:443", "upstream.internal", http.Header{"X-Forwarded-Host": {"tenant1.localhost, attacker.localhost"}, "X-Forwarded-Proto": {"https"}}, "", "authority_multihop"},
		{"nginx rejects missing host", securitypolicy.OnPremNginxSingleHop, "10.0.0.2:443", "", http.Header{"X-Forwarded-Host": {"tenant1.localhost"}, "X-Forwarded-Proto": {"https"}}, "", "authority_invalid"},
		{"local uses direct host", securitypolicy.LocalDirect, "127.0.0.1:443", "tenant1.localhost", nil, "tenant1.localhost", ""},
		{"local rejects forwarded shortcut", securitypolicy.LocalDirect, "127.0.0.1:443", "tenant1.localhost", http.Header{"X-Forwarded-Proto": {"https"}}, "", "authority_conflicting_headers"},
		{"local rejects xff", securitypolicy.LocalDirect, "127.0.0.1:443", "tenant1.localhost", http.Header{"X-Forwarded-For": {"203.0.113.1"}}, "", "authority_conflicting_headers"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := tenantPolicy()
			policy.DeploymentProfile = tt.profile
			if tt.profile == securitypolicy.OnPremNginxSingleHop {
				policy.NginxTrustedCIDRs = []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
			}
			req := httptest.NewRequest(http.MethodGet, "https://"+tt.host, nil)
			req.Host, req.RemoteAddr, req.Header = tt.host, tt.remote, tt.headers
			host, _, err := requestAuthority(req, policy)
			if tt.code != "" {
				if err == nil || err.Error() != tt.code {
					t.Fatalf("error = %v, want %q", err, tt.code)
				}
				return
			}
			if err != nil || host != tt.want {
				t.Fatalf("host/error = %q/%v, want %q/nil", host, err, tt.want)
			}
		})
	}
}

func TestTenantMiddlewareDeniesAuthorityBeforeDownstream(t *testing.T) {
	policy := tenantPolicy()
	policy.DeploymentProfile = securitypolicy.GCPCloudRunALB
	called := false
	req := httptest.NewRequest(http.MethodGet, "https://tenant1.localhost", nil)
	req.Header.Set("X-Forwarded-Host", "tenant1.localhost")
	rr := httptest.NewRecorder()
	TenantMiddleware(&app.App{TenantPolicy: policy})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(rr, req)
	if rr.Code != http.StatusMisdirectedRequest || called {
		t.Fatalf("status/called = %d/%t, want 421/false", rr.Code, called)
	}
}

func TestTenantMiddlewareAllowsTerminatedTLS(t *testing.T) {
	tests := []struct {
		name    string
		profile securitypolicy.DeploymentTrustProfile
		remote  string
		headers http.Header
	}{
		{"gcp", securitypolicy.GCPCloudRunALB, "198.51.100.1:8080", http.Header{"X-Forwarded-Proto": {"https"}}},
		{"nginx", securitypolicy.OnPremNginxSingleHop, "10.0.0.2:8080", http.Header{"X-Forwarded-Host": {"tenant1.localhost"}, "X-Forwarded-Proto": {"https"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := tenantPolicy()
			policy.DeploymentProfile = tt.profile
			if tt.profile == securitypolicy.OnPremNginxSingleHop {
				policy.NginxTrustedCIDRs = []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
			}
			req := httptest.NewRequest(http.MethodGet, "http://tenant1.localhost", nil)
			req.RemoteAddr, req.Header = tt.remote, tt.headers
			called, rr := false, httptest.NewRecorder()
			TenantMiddleware(&app.App{TenantPolicy: policy})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(rr, req)
			if rr.Code != http.StatusOK || !called {
				t.Fatalf("status/called = %d/%t, want 200/true", rr.Code, called)
			}
		})
	}
}
func tenantFromRequest(r *http.Request) (tenantcontext.Source, error) {
	tenant, err := tenantcontext.TenantFromContext(r.Context())
	return tenant.Source, err
}
