package middleware

import (
	"PlantillaInstitucional/shared/contextkey"
	"PlantillaInstitucional/shared/sesionutil"
	"context"
	"net/http"

	"github.com/issemym-repos/sesiones"
)

// MiddlewareRequiereLogin redirige al login si el usuario no tiene una sesión activa.
func MiddlewareRequiereLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sesion := sesiones.RecuperarSesion(r.Context())
		if !sesion.EstaFirmado() {
			// Podríamos agregar un mensaje flash aquí si la librería lo soporta en este punto
			http.Redirect(w, r, "/login", http.StatusFound)
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
		
		// Inyectamos en la llave esperada por el repositorio (shared/db/db.go)
		ctx = context.WithValue(ctx, contextkey.AuditUUIDKey, uuid)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
