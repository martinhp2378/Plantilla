package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/pagina"
	"PlantillaInstitucional/shared/sesionutil"
	"PlantillaInstitucional/shared/tenantcontext"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/layouts"
	"context"
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

			// Validar Tenant de la sesión contra el request
			u := sesionutil.UsuarioDeContexto(r.Context())
			if tc, err := tenantcontext.TenantFromContext(r.Context()); err == nil && tc.Schema != "" {
				tenantSesion := u.TenantSchema
				if tenantSesion == "" {
					tenantSesion = "public"
				}
				if tenantSesion != tc.Schema {
					a.Logger.WarnContext(r.Context(), "bloqueo por tenant de sesión inválido", "tenant_sesion", tenantSesion, "tenant_request", tc.Schema)
					http.Error(w, "Acceso denegado: tenant inválido", http.StatusForbidden)
					return
				}
			}

			// 3. Validar Política
			perfil := u.Perfil
			if perfil == "" {
				perfil = string(ses.Perfil)
			}

			var (
				permitido bool
				err       error
			)
			if withMethod, ok := a.Services.Policy.(interface {
				IsRouteAllowedWithMethod(ctx context.Context, profile string, method string, path string) (bool, error)
			}); ok {
				permitido, err = withMethod.IsRouteAllowedWithMethod(r.Context(), perfil, r.Method, path)
			} else {
				permitido, err = a.Services.Policy.IsRouteAllowed(r.Context(), perfil, path)
			}
			if err != nil {
				a.Logger.ErrorContext(r.Context(), "error al validar política de acceso", err, "path", path, "perfil", perfil)
				http.Error(w, "Error interno de seguridad", http.StatusInternalServerError)
				return
			}

			if !permitido {
				a.Logger.WarnContext(r.Context(), "acceso denegado por política", "path", path, "perfil", perfil)

				solicitudHTMX := esSolicitudHTMX(r)
				if solicitudHTMX {
					w.Header().Set("HX-Retarget", "body")
				}
				w.WriteHeader(http.StatusForbidden)

				if solicitudHTMX {
					// Si es HTMX, podemos renderizar el error en el contenedor principal o mostrar alerta
					utils.RenderIgnore(r.Context(), w, layouts.Pagina403())
					return
				}

				utils.RenderIgnore(r.Context(), w, layouts.Pagina403())
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func esSolicitudHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}
