package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/contextkey"
	"context"
	"net/http"
	"strings"
)

func TenantMiddleware(a *app.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Si el multi-tenant está deshabilitado, forzamos siempre public de la forma más segura.
			if a.Env != nil && !a.Env.HabilitarMultiTenant {
				ctx := context.WithValue(r.Context(), contextkey.TenantSchemaKey, "public")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			host := r.Host
			// Remove port if present
			if strings.Contains(host, ":") {
				host = strings.Split(host, ":")[0]
			}

			tenant, err := a.Repositories.Tenants.GetByHost(host)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if tenant == nil {
				// Fallback to public schema or return error
				// For this template, let's assume public if not found
				ctx := context.WithValue(r.Context(), contextkey.TenantSchemaKey, "public")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			ctx := context.WithValue(r.Context(), contextkey.TenantSchemaKey, tenant.Schema)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
