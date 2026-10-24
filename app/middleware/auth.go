package middleware

import (
	"PlantillaInstitucional/shared/contextkey"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/sesionutil"
	"context"
	"net/http"
	"strings"

	"github.com/issemym-repos/sesiones"
)

func MiddlewareRequiereLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sesion := sesiones.RecuperarSesion(r.Context())

		if !sesion.EstaFirmado() {
			if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
				w.Header().Set("HX-Redirect", pagina.GlobalRutas.Login.Prefix)
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, pagina.GlobalRutas.Login.Prefix, http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuditContextMiddleware inyecta el UUID del usuario autenticado en el contexto
// para que esté disponible en las operaciones de base de datos para auditoría.
func AuditContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		uuid := sesionutil.ObtenerUsuarioParaBd(ctx)

		ctx = context.WithValue(ctx, contextkey.AuditUUIDKey, uuid)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
