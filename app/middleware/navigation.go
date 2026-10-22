package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/navigation"
	"PlantillaInstitucional/shared/sesionutil"
	"net/http"

	"github.com/issemym-repos/sesiones"
)

// NavigationMiddleware supplies presentation permissions without changing auth or sessions.
func NavigationMiddleware(a *app.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := navigation.WithPaths(r.Context(), nil)
			usuario := sesionutil.UsuarioDeContexto(ctx)
			if usuario.Perfil != "" {
				session := sesiones.RecuperarSesion(ctx)
				if session != nil && session.EstaFirmado() {
					if reader, ok := a.Services.Policy.(navigation.Reader); ok {
						paths, err := reader.NavigationPaths(ctx, usuario.Perfil)
						if err == nil {
							ctx = navigation.WithPaths(ctx, paths)
						}
					}
				}
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
