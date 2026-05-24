package middleware

import (
	"PlantillaInstitucional/app"
	"PlantillaInstitucional/shared/entorno"
	"PlantillaInstitucional/shared/utils"
	"PlantillaInstitucional/shared/views/layouts"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type serviciosMiddleware struct {
	a *app.App
}

func RegistrarMiddlewaresApp(r chi.Router, a *app.App) {
	sm := &serviciosMiddleware{a: a}

	r.Use(SecurityHeadersMiddleware)
	r.Use(CSRFMiddleware)

	r.Use(TenantMiddleware(a))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	
	r.Use(sm.CustomRecoverer)

	r.Use(a.Logger.MiddlewareHTTPLogger)

	r.Use(a.Instrumentacion.MiddlewareOtel())
	r.Use(a.Instrumentacion.MiddlewareAttrsAdicionales)
	r.Use(a.Instrumentacion.MiddlewareMetricas)

	r.Use(a.ManejadorArchivos.Middleware)

	r.Use(a.ConstructorSesion.Middleware)

	r.Use(entorno.EnvMiddleware(a.Env))

	r.Use(AuditContextMiddleware)

	r.Use(middleware.Timeout(a.Env.Timeout))
}

// CustomRecoverer captura panics y muestra la página 500 institucional.
func (sm *serviciosMiddleware) CustomRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				var err error
				switch v := rvr.(type) {
				case error:
					err = v
				case string:
					err = fmt.Errorf("%s", v)
				default:
					err = fmt.Errorf("%v", v)
				}

				sm.a.Logger.ErrorContext(r.Context(), "se recuperó de un panic", err)

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
					acceptHeader := r.Header.Get("Accept")

					// Si acepta HTML, mandamos la página 500 institucional
					if acceptHeader == "" || strings.Contains(acceptHeader, "text/html") || acceptHeader == "*/*" {
						w.Header().Set("Content-Type", "text/html; charset=utf-8")
						utils.RenderIgnore(r.Context(), w, layouts.Pagina500())
					}
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
