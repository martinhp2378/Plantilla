package sesiones

import (
	"encoding/base64"
	"net/http"
	"time"
)

func (cs *ConstructorSesion) colocarCookie(w http.ResponseWriter, sesion *Sesion) {
	uuidSesion := sesion.UuidSesion
	valorCookie := uuidSesion.String()
	if cs.config.CookieCodificar {
		valorCookie = base64.URLEncoding.EncodeToString([]byte(valorCookie))
	}
	cookie := &http.Cookie{
		Name:     cs.config.CookieNombre,
		Value:    valorCookie,
		Path:     cs.config.CookiePath,
		Domain:   cs.config.CookieDominio,
		Secure:   cs.config.CookieEsSecure,
		HttpOnly: true,
		SameSite: cs.config.CookieMismoSitio,
	}

	// Colocar MaxAge para que el cookie sea persistente y permanezca tras
	// cerrar el browser y etc.
	if sesion.FechaExpiracion != nil {
		// Usar el tiempo hasta expirar si está firmado
		maxAge := int(time.Until(*sesion.FechaExpiracion).Seconds())
		if maxAge > 0 {
			cookie.MaxAge = maxAge
		} else {
			// Usar el mínimo valor posible si ya expiró
			cookie.MaxAge = 1
		}
	} else {
		// En sesiones anónimas se usa la DuracionInactiva
		cookie.MaxAge = int(cs.config.DuracionInactiva.Seconds())
	}

	http.SetCookie(w, cookie)
}
