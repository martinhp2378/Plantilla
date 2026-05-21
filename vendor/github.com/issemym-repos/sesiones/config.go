package sesiones

import (
	"net/http"
	"time"
)

type ConfigSesion struct {
	CookieNombre     string
	CookiePath       string
	CookieDominio    string
	CookieEsSecure   bool
	CookieSoloHTTP   bool
	CookieMismoSitio http.SameSite
	CookieCodificar  bool

	DuracionInactiva time.Duration
	DuracionMaxima   time.Duration

	FlashTTL             time.Duration
	MaximosMensajesFlash int

	HabilitarLimitePeticiones  bool
	PeticionesMaximasPorMinuto int

	IntervaloRegeneracion time.Duration

	HabilitarFingerprinting bool

	// En caso de ser falso, una sesión expira a las 23:59:59 del mismo día
	// que se creó si su DuracionMaxima permite que se pase al día siguiente.
	SesionDuraMasDeUnDia bool

	// Configuracion de Redis opcional.
	// Hace que la 'tabla' de `sesion.sesion` se inserte y obtenga de Redis
	// en lugar de postgres.
	// TODO - Colocar tambien las politicas en Redis?
	HabilitarRedis bool
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
}

func ConfigDefault() *ConfigSesion {
	return &ConfigSesion{
		CookieNombre:               "id_sesion",
		CookiePath:                 "/",
		CookieDominio:              "",
		CookieEsSecure:             true,
		CookieSoloHTTP:             true,
		CookieMismoSitio:           http.SameSiteStrictMode,
		CookieCodificar:            true,
		DuracionInactiva:           30 * time.Minute,
		DuracionMaxima:             12 * time.Hour,
		FlashTTL:                   5 * time.Minute,
		MaximosMensajesFlash:       10,
		HabilitarLimitePeticiones:  true,
		PeticionesMaximasPorMinuto: 120,
		IntervaloRegeneracion:      4 * time.Hour,
		HabilitarFingerprinting:    false,
		SesionDuraMasDeUnDia:       true,
		HabilitarRedis:             false,
		RedisAddr:                  "localhost:6379",
		RedisPassword:              "",
		RedisDB:                    0,
	}
}
