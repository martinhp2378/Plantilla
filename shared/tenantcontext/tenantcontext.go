package tenantcontext

import (
	"context"
	"errors"
	"regexp"
)

var ErrMissingTenantContext = errors.New("missing tenant context")
var ErrUnsafeTenantContext = errors.New("unsafe tenant context")
var schemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type Source string

const DirectHost Source = "direct_host"
const TrustedForwardedHost Source = "trusted_forwarded_host"

type TenantContext struct {
	ID, Schema, CanonicalHost string
	Source                    Source
}
type contextKey struct{}

func WithTenant(ctx context.Context, tenant TenantContext) context.Context {
	return context.WithValue(ctx, contextKey{}, tenant)
}
func TenantFromContext(ctx context.Context) (TenantContext, error) {
	tenant, ok := ctx.Value(contextKey{}).(TenantContext)
	if !ok {
		return TenantContext{}, ErrMissingTenantContext
	}
	if tenant.ID == "" || tenant.CanonicalHost == "" || tenant.Schema == "public" || !schemaPattern.MatchString(tenant.Schema) {
		return TenantContext{}, ErrUnsafeTenantContext
	}
	return tenant, nil
}
