package sesiones

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/logs"
)

// MiddlewareJWT es el middleware para manejar autenticación basada en JWT.
// Extrae el token JWT del header Authorization, lo valida, carga la sesión
// desde el store y la agrega al contexto del request.
// A diferencia del middleware de cookies, este NO crea sesiones anónimas automáticamente.
func (cj *ConstructorSesionJWT) MiddlewareJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString, err := extraerTokenDeHeader(r)
		if err != nil {
			responderError(w, http.StatusUnauthorized, "Token no encontrado", err)
			return
		}

		claims, err := cj.ValidarToken(tokenString)
		if err != nil {
			responderError(w, http.StatusUnauthorized, "Token inválido", err)
			return
		}

		if claims.TipoToken != TipoTokenAccess {
			responderError(w, http.StatusUnauthorized, "Tipo de token inválido", nil)
			return
		}

		uuidSesion, err := uuid.FromString(claims.UuidSesion)
		if err != nil {
			responderError(w, http.StatusUnauthorized, "UUID de sesión inválido", err)
			return
		}

		// Cargar sesión completa desde el store (PostgreSQL o Redis)
		sesion, err := cj.SesionDeUuid(ctx, uuidSesion)
		if err != nil {
			// Sesión no encontrada en el store
			responderError(w, http.StatusUnauthorized, "Sesión no encontrada", err)
			return
		}

		// Validar que la sesión sea de tipo JWT
		if sesion.TipoSesion != TipoSesionJWT {
			responderError(w, http.StatusUnauthorized, "Tipo de sesión inválido", nil)
			return
		}

		// Verificar que la sesión no haya expirado en la base de datos
		if cj.YaExpiro(sesion) {
			// Eliminar sesión para evitar fugas
			_ = cj.EliminarSesion(ctx, sesion.UuidSesion)
			responderError(w, http.StatusUnauthorized, "Sesión expirada", nil)
			return
		}

		// TODO: Validar fingerprint si está habilitado en la configuración
		// if cj.config.ValidarHuella {
		// 	huellaActual := generarHuella(r)
		// 	if huellaActual != sesion.Huella {
		// 		responderError(w, http.StatusUnauthorized, "Huella de dispositivo no coincide", nil)
		// 		return
		// 	}
		// }

		cj.Tocar(sesion)

		ctx = context.WithValue(ctx, LlaveContexto, sesion)

		// Variables de contexto para instrumentación (las ocupa la librería de logs)
		if sesion.EstaFirmado() {
			ctx = context.WithValue(ctx, logs.LlaveUuidSesion, sesion.UuidSesion)
			ctx = context.WithValue(ctx, logs.LlaveUuidPersona, *sesion.UuidPersona)
		} else {
			ctx = context.WithValue(ctx, logs.LlaveUuidSesion, sesion.UuidSesion)
		}

		next.ServeHTTP(w, r.WithContext(ctx))

		// Se usa context.Background() para que no se cancele si el cliente cerró la conexión
		err = cj.ActualizarSesion(context.Background(), sesion)
		if err != nil {
			// Log error pero no panic (a diferencia del middleware de cookies)
			// porque en APIs es mejor no interrumpir el flujo
			// TODO: usar librería de logs aquí
			slog.Error("Error al actualizar sesión JWT", "error", err)
		}
	})
}

// responderError es una función helper para enviar respuestas de error HTTP
// en formato JSON.
func responderError(w http.ResponseWriter, statusCode int, mensaje string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	respuesta := map[string]string{
		"error": mensaje,
	}

	if err != nil {
		respuesta["detalle"] = err.Error()
	}

	json.NewEncoder(w).Encode(respuesta)
}
