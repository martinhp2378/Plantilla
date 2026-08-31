package middleware

import (
	"PlantillaInstitucional/shared/contextkey"
	"context"
	"net/http"

	"github.com/justinas/nosurf"
)

// SecurityHeadersMiddleware añade cabeceras de seguridad estándar a todas las respuestas.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// CSRFMiddleware configura la protección contra CSRF usando nosurf.
func CSRFMiddleware(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inyectamos el token en el contexto para que templ pueda leerlo
		token := nosurf.Token(r)
		ctx := context.WithValue(r.Context(), contextkey.CSRFTokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	}))

	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Error CSRF: token inválido o ausente", http.StatusBadRequest)
	}))

	return csrfHandler
}

// GetCSRFToken recupera el token del contexto.
func GetCSRFToken(ctx context.Context) string {
	if token, ok := ctx.Value(contextkey.CSRFTokenKey).(string); ok {
		return token
	}
	return ""
}
