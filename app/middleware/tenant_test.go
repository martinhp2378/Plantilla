package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/securitypolicy"
	"PlantillaInstitucional/shared/tenantcontext"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func tenantPolicy() securitypolicy.TenantPolicy {
	return securitypolicy.TenantPolicy{EnforcementEnabled: true, CanonicalHosts: []securitypolicy.CanonicalHostRule{{Host: "tenant1.localhost", TenantID: "tenant1", Schema: "tenant1", Active: true}}, AllowedSchemes: []string{"https"}, GlobalAccessPolicyID: "global-readers-v1"}
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
func TestTenantMiddlewareUsesForwardedAuthorityOnlyFromTrustedProxy(t *testing.T) {
	policy := tenantPolicy()
	policy.ForwardedAuthorityMode, policy.TrustedProxyCIDRs = securitypolicy.TrustedProxyOnly, []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}
	var source tenantcontext.Source
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { source, _ = tenantFromRequest(r) })
	req := httptest.NewRequest(http.MethodGet, "https://edge.invalid", nil)
	req.RemoteAddr, req.Header = "127.0.0.1:1234", http.Header{"X-Forwarded-Host": []string{"tenant1.localhost"}}
	rr := httptest.NewRecorder()
	TenantMiddleware(&app.App{TenantPolicy: policy})(next).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || source != tenantcontext.TrustedForwardedHost {
		t.Fatalf("status/source = %d/%q", rr.Code, source)
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
func tenantFromRequest(r *http.Request) (tenantcontext.Source, error) {
	tenant, err := tenantcontext.TenantFromContext(r.Context())
	return tenant.Source, err
}
