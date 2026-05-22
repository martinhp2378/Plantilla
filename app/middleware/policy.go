package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/pagina"
	"net/http"
	"strings"

	"github.com/issemym-repos/sesiones"
)

// PolicyMiddleware verifica que el usuario autenticado tenga permisos para acceder a la ruta solicitada.
func PolicyMiddleware(a *app.App) func(http.Handler) http.Handler {
	// Rutas que no requieren validación de políticas (además de las públicas como /login)
	rutasExcluidas := []string{
		"/health",
		"/static",
		pagina.GlobalRutas.Login.Base + "salir",
		"/sesion",
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// 1. Omitir rutas excluidas
			for _, ex := range rutasExcluidas {
				if strings.HasPrefix(path, ex) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 2. Obtener Sesión
			ses := sesiones.RecuperarSesion(r.Context())
			if !ses.EstaFirmado() {
				// Si no está firmado, el middleware de sesiones ya debería haber manejado la redirección 
				// o estamos en una ruta pública que no listamos en excluidas por error.
				// Dejamos pasar para que el handler decida.
				next.ServeHTTP(w, r)
				return
			}

			// 3. Validar Política
			perfil := string(ses.Perfil)
			
			// Si es superadmin y no queremos ser tan estrictos con cada sub-ruta, podríamos dejarlo pasar,
			// pero lo ideal es que TODO esté en la tabla de políticas.
			
			permitido, err := a.Services.Policy.IsRouteAllowed(r.Context(), perfil, path)
			if err != nil {
				a.Logger.ErrorContext(r.Context(), "error al validar política de acceso", err, "path", path, "perfil", perfil)
				http.Error(w, "Error interno de seguridad", http.StatusInternalServerError)
				return
			}

			if !permitido {
				a.Logger.WarnContext(r.Context(), "acceso denegado por política", "path", path, "perfil", perfil)
				
				if esSolicitudHTMX(r) {
					w.Header().Set("HX-Redirect", pagina.GlobalRutas.Login.Prefix) // O a una página de "Sin Permisos"
					w.WriteHeader(http.StatusForbidden)
					return
				}
				
				http.Redirect(w, r, pagina.GlobalRutas.Login.Prefix, http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func esSolicitudHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}
