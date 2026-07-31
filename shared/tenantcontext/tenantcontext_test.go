package tenantcontext

import (
	"context"
	"errors"
	"testing"
)

func TestTenantContextRoundTrip(t *testing.T) {
	want := TenantContext{ID: "tenant-a", Schema: "tenant_a", CanonicalHost: "a.example.test", Source: DirectHost}
	got, err := TenantFromContext(WithTenant(context.Background(), want))
	if err != nil {
		t.Fatalf("TenantFromContext() error = %v", err)
	}
	if got != want {
		t.Fatalf("TenantFromContext() = %#v, want %#v", got, want)
	}
}

func TestTenantFromContextRejectsMissingOrUnsafeContext(t *testing.T) {
	if _, err := TenantFromContext(context.Background()); !errors.Is(err, ErrMissingTenantContext) {
		t.Fatalf("TenantFromContext() error = %v, want ErrMissingTenantContext", err)
	}
	ctx := WithTenant(context.Background(), TenantContext{ID: "tenant-a", Schema: "public", CanonicalHost: "a.example.test"})
	if _, err := TenantFromContext(ctx); !errors.Is(err, ErrUnsafeTenantContext) {
		t.Fatalf("TenantFromContext() error = %v, want ErrUnsafeTenantContext", err)
	}
}
