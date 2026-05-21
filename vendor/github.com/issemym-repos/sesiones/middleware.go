package sesiones

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
)

type llaveContexto string

const LlaveContexto llaveContexto = "sesion"

// Middleware es el middleware recomendado que se debería colocar antes de
// las rutas de la aplicación.
// Siempre crea una instancia de *Sesion y la agrega al contexto del Request
// con la llave [LlaveContexto].
// Esta instancia puede ser modificada y al final de las peticiones, será
// actualizada en el Store (Redis o Postgres).
func (cs *ConstructorSesion) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var sesion *Sesion
		var err error
		var sesionExpirada bool

		// extraer el uuid de sesion de la cookie
		cookie, err := r.Cookie(cs.config.CookieNombre)

		if err == nil {
			uuidSesionStr := cookie.Value
			if cs.config.CookieCodificar {
				decodificada, err := base64.URLEncoding.DecodeString(uuidSesionStr)
				if err != nil {
					uuidSesionStr = ""
				}

				uuidSesionStr = string(decodificada)
			}
			uuidSesion, err := uuid.FromString(uuidSesionStr)
			if err == nil {
				sesion, err = cs.SesionDeUuid(ctx, uuidSesion)
				if err == nil {
					if cs.YaExpiro(sesion) {
						// Se elimina la sesión para evitar fugas
						_ = cs.EliminarSesion(ctx, sesion.UuidSesion)
						sesion.CerrarSesion()
						sesionExpirada = true
					} else {
						// Para evitar ataques de fijación de sesión
						if cs.debeRegenerar(sesion) {
							// También se elimina la sesión vieja
							oldUuid := sesion.UuidSesion
							cs.Regenerar(sesion)
							_ = cs.EliminarSesion(ctx, oldUuid)
						}
					}
				}
			}
		}

		if sesion == nil || err != nil {
			// intentar buscar de nuevo la sesion pero ahora por huella
			huella := generarHuella(r)
			sesion, err = cs.SesionDeHuella(ctx, huella)
			if err == nil {
				sesion.CerrarSesion() // asegurarse que es Sesion anónima
			}
		}

		if sesion == nil || err != nil {
			sesion = cs.NuevaSesion(r)
		}

		// Tocar primero para actualizar FechaExpiracion antes de colocar la cookie
		cs.Tocar(sesion)
		cs.colocarCookie(w, sesion)

		if sesionExpirada {
			// avisar al usuario con un mensaje flash?
			cs.AgregarFlash(sesion, FlashError, "Su sesión ha expirado", "Favor de firmarse otra vez.")
		}

		ctx = context.WithValue(ctx, LlaveContexto, sesion)

		// Variables de contexto para instrumentación (las ocupa la librería de logs)
		if sesion.EstaFirmado() {
			ctx = context.WithValue(ctx, logs.LlaveUuidSesion, sesion.UuidSesion)
			ctx = context.WithValue(ctx, logs.LlaveUuidPersona, *sesion.UuidPersona)
		} else {
			ctx = context.WithValue(ctx, logs.LlaveUuidSesion, sesion.UuidSesion)
		}

		next.ServeHTTP(w, r.WithContext(ctx))

		err = cs.ActualizarOInsertarSesion(context.Background(), sesion)
		if err != nil {
			panic(err)
		}
	})
}

func RecuperarSesion(ctx context.Context) *Sesion {
	return ctx.Value(LlaveContexto).(*Sesion)
}
