package entorno

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/issemym-repos/logs"
	"github.com/issemym-repos/sesiones"
	"github.com/joho/godotenv"
)

const (
	AmbienteLocal      = "LOCAL"
	AmbienteDesarrollo = "DESARROLLO"
	AmbienteProduccion = "PRODUCCION"
)

type Env struct {
	// General
	Servicio   string `json:"servicio"`
	IdProyecto string `json:"idProyecto"`
	Version    string `json:"version"`
	Ambiente   string `json:"ambiente"`

	Timeout time.Duration `json:"timeout"`

	// Puerto del server, debe ser 8080 para Cloud Run
	ServerListen string `json:"serverListen"`

	// Base de datos
	PqWriteDb           string `json:"pqWriteDb"`
	PqWriteUser         string `json:"pqWriteUser"`
	PqWritePass         string
	PqWriteHost         string        `json:"pqWriteHost"`
	PqWritePort         string        `json:"pqWritePort"`
	PqWriteSSLMode      string        `json:"pqWriteSSLMode"`
	PqWriteMaxConns     int32         `json:"pqWriteMaxConns"`
	PqWriteMinConns     int32         `json:"pqWriteMinConns"`
	PqWriteMaxConnIdle  time.Duration `json:"pqWriteMaxConnIdle"`
	PqWriteMaxConnLife  time.Duration `json:"pqWriteMaxConnLife"`
	PqWriteQueryTimeout time.Duration `json:"pqWriteQueryTimeout"`

	// Base de datos read replica (usar los mismos valores que write para
	// ambiente local)
	PqReadDb           string `json:"pqReadDb"`
	PqReadUser         string `json:"pqReadUser"`
	PqReadPass         string
	PqReadHost         string        `json:"pqReadHost"`
	PqReadPort         string        `json:"pqReadPort"`
	PqReadSSLMode      string        `json:"pqReadSSLMode"`
	PqReadMaxConns     int32         `json:"pqReadMaxConns"`
	PqReadMinConns     int32         `json:"pqReadMinConns"`
	PqReadMaxConnIdle  time.Duration `json:"pqReadMaxConnIdle"`
	PqReadMaxConnLife  time.Duration `json:"pqReadMaxConnLife"`
	PqReadQueryTimeout time.Duration `json:"pqReadQueryTimeout"`

	// Redis, dejar el host vacío si no se requiere una conexión
	RedisHost     string `json:"redisHost"`
	RedisPort     string `json:"redisPort"`
	RedisPassword string `json:"redisPassword"`
	RedisDB       int    `json:"redisDB"`

	// Logger
	LogsConColores bool `json:"logsConColores"`
	AgregarSource  bool `json:"agregarSource"`
	Nivel          logs.Severidad
	// Instrumentación BD
	NombreSpanCorto    bool `json:"nombreSpanCorto"`
	QuitarArgsQuery    bool `json:"quitarArgsQuery"`
	QuitarPrefijoQuery bool `json:"quitarPrefijoQuery"`
	QuitarStatement    bool `json:"quitarStatement"`
	// Instrumentación HTTP
	TasaSampling                 float64       `json:"tasaSampling"`
	IntervaloMetricas            time.Duration `json:"intervaloMetricas"`
	HabilitarMetricasAutomaticas bool          `json:"habilitarMetricasAutomaticas"`
	DeshabilitarTracing          bool          `json:"deshabilitarTracing"`
	DeshabilitarMetricas         bool          `json:"deshabilitarMetricas"`
	UsarExportadorConsola        bool          `json:"usarExportadorConsola"`

	// Multi-tenant
	HabilitarMultiTenant bool `json:"habilitarMultiTenant"`

	// Sesiones
	CookieNombre               string        `json:"cookieNombre"`
	CookiePath                 string        `json:"cookiePath"`
	CookieDominio              string        `json:"cookieDominio"`
	CookieEsSecure             bool          `json:"cookieEsSecure"`
	CookieSoloHTTP             bool          `json:"cookieSoloHTTP"`
	CookieMismoSitio           http.SameSite `json:"cookieMismoSitio"`
	CookieCodificar            bool          `json:"cookieCodificar"`
	DuracionInactiva           time.Duration `json:"duracionInactiva"`
	DuracionMaxima             time.Duration `json:"duracionMaxima"`
	FlashTTL                   time.Duration `json:"flashTTL"`
	MaximosMensajesFlash       int           `json:"maximosMensajesFlash"`
	HabilitarLimitePeticiones  bool          `json:"habilitarLimitePeticiones"`
	PeticionesMaximasPorMinuto int           `json:"peticionesMaximasPorMinuto"`
	IntervaloRegeneracion      time.Duration `json:"intervaloRegeneracion"`
	HabilitarFingerprinting    bool          `json:"habilitarFingerprinting"`
	SesionDuraMasDeUnDia       bool          `json:"sesionDuraMasDeUnDia"`
	HabilitarRedis             bool          `json:"habilitarRedis"`

	// Sesiones JWT
	// Nota: El issuer se debe colocar igual al Servicio
	JWTSecretKey           []byte
	JWTAlgorithm           string        `json:"JWTAlgorithm"`
	JWTAccessTokenExpiry   time.Duration `json:"JWTAccessTokenExpiry"`
	JWTRefreshTokenExpiry  time.Duration `json:"JWTRefreshTokenExpiry"`
	JWTEnableRefreshTokens bool          `json:"JWTEnableRefreshTokens"`
	JWTAudience            []string      `json:"JWTAudience"`
	JWTAllowedClockSkew    time.Duration `json:"JWTAllowedClockSkew"`

	// Archivos

	DirArchivosLocal       string        `json:"dirArchivosLocal"`
	RutaPubLocal           string        `json:"rutaPubLocal"`
	RutaPrivLocal          string        `json:"rutaPrivLocal"`
	CubetaPublicaGCP       string        `json:"cubetaPublicaGCP"`
	CubetaPrivadaGCP       string        `json:"cubetaPrivadaGCP"`
	UrlCDN                 string        `json:"urlCDN"`
	NombreURLMapCDN        string        `json:"nombreURLMapCDN"`
	ExpiracionURLsFirmados time.Duration `json:"expiracionURLsFirmados"`
	CacheControl           string        `json:"cacheControl"`

	// Email
	SMTPHost     string `json:"smtpHost"`
	SMTPPort     string `json:"smtpPort"`
	SMTPUser     string `json:"smtpUser"`
	SMTPPassword string `json:"smtpPassword"`
	SMTPFrom     string `json:"smtpFrom"`
	SMTPFromName string `json:"smtpFromName"`
	SMTPUseTLS   bool   `json:"smtpUseTLS"`

	// Variables particulares de esta app
	VariablePersonal string `json:"variabePersonal"`
}

func GetEnv(envPaths ...string) (*Env, error) {
	// Intentar cargar del path, si falla puede que sea porque está en
	// Cloud Run (ambiente no Local) o simplemente el archivo no existe
	err := godotenv.Load(envPaths...)
	if err != nil {
		slog.Warn("Error al cargar archivo con variables de entorno: " + err.Error())
	}

	env := &Env{}

	ambiente := os.Getenv("AMBIENTE")
	if !slices.Contains([]string{AmbienteLocal, AmbienteDesarrollo, AmbienteProduccion}, ambiente) {
		return nil, fmt.Errorf("ambiente inválido, se esperaba %s, %s o %s",
			AmbienteLocal, AmbienteDesarrollo, AmbienteProduccion)
	}
	env.Ambiente = ambiente

	if Servicio := os.Getenv("SERVICIO"); Servicio != "" {
		env.Servicio = Servicio
	} else {
		return nil, fmt.Errorf("el valor %s no puede estar vacío", "SERVICIO")
	}

	if IdProyecto := os.Getenv("ID_PROYECTO"); IdProyecto != "" {
		env.IdProyecto = IdProyecto
	} else {
		return nil, fmt.Errorf("el valor %s no puede estar vacío", "ID_PROYECTO")
	}

	if Version := os.Getenv("VERSION"); Version != "" {
		env.Version = Version
	} else {
		return nil, fmt.Errorf("el valor %s no puede estar vacío", "VERSION")
	}

	if timeoutStr := os.Getenv("TIMEOUT"); timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un duration", "TIMEOUT")
		}
		env.Timeout = timeout
	} else {
		env.Timeout = 120 * time.Second
	}

	serverListen := os.Getenv("LISTEN_ADDR")
	if !strings.Contains(serverListen, ":") {
		return nil, fmt.Errorf("el puerto del server debe tener el formato :8080")
	} else {
		env.ServerListen = serverListen
	}

	env.PqWriteDb = os.Getenv("PQ_WRITE_DB")
	env.PqWriteUser = os.Getenv("PQ_WRITE_USER")
	env.PqWritePass = os.Getenv("PQ_WRITE_PASSWORD")
	env.PqWriteHost = os.Getenv("PQ_WRITE_HOST")
	env.PqWritePort = os.Getenv("PQ_WRITE_PORT")
	env.PqWriteSSLMode = os.Getenv("PQ_WRITE_SSL_MODE")

	writeMaxConns, err := strconv.ParseInt(os.Getenv("PQ_WRITE_MAX_CONNS"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_WRITE_MAX_CONNS")
	}
	writeMinConns, err := strconv.ParseInt(os.Getenv("PQ_WRITE_MIN_CONNS"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_WRITE_MIN_CONNS")
	}
	writeMaxConnIdle, err := strconv.ParseInt(os.Getenv("PQ_WRITE_MAX_CONN_IDLE"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_WRITE_MAX_CONN_IDLE")
	}
	writeMaxConnLife, err := strconv.ParseInt(os.Getenv("PQ_WRITE_MAX_CONN_LIFE"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_WRITE_MAX_CONN_LIFE")
	}
	writeQueryTimeout, err := strconv.ParseInt(os.Getenv("PQ_WRITE_QUERY_TIMEOUT"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_WRITE_QUERY_TIMEOUT")
	}

	env.PqReadDb = os.Getenv("PQ_READ_DB")
	env.PqReadUser = os.Getenv("PQ_READ_USER")
	env.PqReadPass = os.Getenv("PQ_READ_PASSWORD")
	env.PqReadHost = os.Getenv("PQ_READ_HOST")
	env.PqReadPort = os.Getenv("PQ_READ_PORT")
	env.PqReadSSLMode = os.Getenv("PQ_READ_SSL_MODE")

	readMaxConns, err := strconv.ParseInt(os.Getenv("PQ_READ_MAX_CONNS"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_READ_MAX_CONNS")
	}
	readMinConns, err := strconv.ParseInt(os.Getenv("PQ_READ_MIN_CONNS"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_READ_MIN_CONNS")
	}
	readMaxConnIdle, err := strconv.ParseInt(os.Getenv("PQ_READ_MAX_CONN_IDLE"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_READ_MAX_CONN_IDLE")
	}
	readMaxConnLife, err := strconv.ParseInt(os.Getenv("PQ_READ_MAX_CONN_LIFE"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_READ_MAX_CONN_LIFE")
	}
	readQueryTimeout, err := strconv.ParseInt(os.Getenv("PQ_READ_QUERY_TIMEOUT"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error al parsear %s, se esperaba valor Int",
			"PQ_READ_QUERY_TIMEOUT")
	}

	env.PqWriteMaxConns = int32(writeMaxConns)
	env.PqWriteMinConns = int32(writeMinConns)
	env.PqWriteMaxConnIdle = time.Duration(writeMaxConnIdle) * time.Minute
	env.PqWriteMaxConnLife = time.Duration(writeMaxConnLife) * time.Minute
	env.PqWriteQueryTimeout = time.Duration(writeQueryTimeout) * time.Second

	env.PqReadMaxConns = int32(readMaxConns)
	env.PqReadMinConns = int32(readMinConns)
	env.PqReadMaxConnIdle = time.Duration(readMaxConnIdle) * time.Minute
	env.PqReadMaxConnLife = time.Duration(readMaxConnLife) * time.Minute
	env.PqReadQueryTimeout = time.Duration(readQueryTimeout) * time.Second

	// Redis
	env.RedisHost = os.Getenv("REDIS_HOST")
	env.RedisPort = os.Getenv("REDIS_PORT")
	env.RedisPassword = os.Getenv("REDIS_PASSWORD")
	redisDbStr := os.Getenv("REDIS_DB")
	if redisDbStr == "" {
		env.RedisDB = 0
	} else {
		redisDb, err := strconv.Atoi(redisDbStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un entero", "REDIS_DB")
		}
		env.RedisDB = redisDb
	}

	// Valores de logs, todos estos son opcionales
	if logsConColoresStr := os.Getenv("LOGS_CON_COLORES"); logsConColoresStr != "" {
		logsConColores, err := strconv.ParseBool(logsConColoresStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "LOGS_CON_COLORES")
		}
		env.LogsConColores = logsConColores
	}
	if agregarSourceStr := os.Getenv("AGREGAR_SOURCE"); agregarSourceStr != "" {
		agregarSource, err := strconv.ParseBool(agregarSourceStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "AGREGAR_SOURCE")
		}
		env.AgregarSource = agregarSource
	}
	if nombreSpanCortoStr := os.Getenv("NOMBRE_SPAN_CORTO"); nombreSpanCortoStr != "" {
		nombreSpanCorto, err := strconv.ParseBool(nombreSpanCortoStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "NOMBRE_SPAN_CORTO")
		}
		env.NombreSpanCorto = nombreSpanCorto
	}
	if quitarArgsQueryStr := os.Getenv("QUITAR_ARGS_QUERY"); quitarArgsQueryStr != "" {
		quitarArgsQuery, err := strconv.ParseBool(quitarArgsQueryStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "QUITAR_ARGS_QUERY")
		}
		env.QuitarArgsQuery = quitarArgsQuery
	}
	if quitarPrefijoQueryStr := os.Getenv("QUITAR_PREFIJO_QUERY"); quitarPrefijoQueryStr != "" {
		quitarPrefijoQuery, err := strconv.ParseBool(quitarPrefijoQueryStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "QUITAR_PREFIJO_QUERY")
		}
		env.QuitarPrefijoQuery = quitarPrefijoQuery
	}
	if quitarStatementStr := os.Getenv("QUITAR_STATEMENT"); quitarStatementStr != "" {
		quitarStatement, err := strconv.ParseBool(quitarStatementStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "QUITAR_STATEMENT")
		}
		env.QuitarStatement = quitarStatement
	}
	if habilitarMetricasAutomaticasStr := os.Getenv("HABILITAR_METRICAS_AUTOMATICAS"); habilitarMetricasAutomaticasStr != "" {
		habilitarMetricasAutomaticas, err := strconv.ParseBool(habilitarMetricasAutomaticasStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "HABILITAR_METRICAS_AUTOMATICAS")
		}
		env.HabilitarMetricasAutomaticas = habilitarMetricasAutomaticas
	}
	if deshabilitarTracingStr := os.Getenv("DESHABILITAR_TRACING"); deshabilitarTracingStr != "" {
		deshabilitarTracing, err := strconv.ParseBool(deshabilitarTracingStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "DESHABILITAR_TRACING")
		}
		env.DeshabilitarTracing = deshabilitarTracing
	}
	if deshabilitarMetricasStr := os.Getenv("DESHABILITAR_METRICAS"); deshabilitarMetricasStr != "" {
		deshabilitarMetricas, err := strconv.ParseBool(deshabilitarMetricasStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "DESHABILITAR_METRICAS")
		}
		env.DeshabilitarMetricas = deshabilitarMetricas
	}
	if usarExportadorConsolaStr := os.Getenv("USAR_EXPORTADOR_CONSOLA"); usarExportadorConsolaStr != "" {
		usarExportadorConsola, err := strconv.ParseBool(usarExportadorConsolaStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "USAR_EXPORTADOR_CONSOLA")
		}
		env.UsarExportadorConsola = usarExportadorConsola
	}

	if nivel := logs.Severidad(os.Getenv("NIVEL")); nivel != "" {
		env.Nivel = nivel
		nivelesValidos := []logs.Severidad{
			logs.SeveridadDefault, logs.SeveridadDebug, logs.SeveridadInfo, logs.SeveridadNotice,
			logs.SeveridadWarning, logs.SeveridadError, logs.SeveridadCritical,
			logs.SeveridadAlert, logs.SeveridadEmergency,
		}
		if !slices.Contains(nivelesValidos, nivel) {
			return nil, fmt.Errorf("el nivel de severidad es inválido, se esperaba %s, %s, %s, %s, %s, %s, %s, %s, o %s",
				logs.SeveridadDefault, logs.SeveridadDebug, logs.SeveridadInfo,
				logs.SeveridadNotice, logs.SeveridadWarning,
				logs.SeveridadError, logs.SeveridadCritical,
				logs.SeveridadAlert, logs.SeveridadEmergency)
		}
	}

	if tasaSamplingStr := os.Getenv("TASA_SAMPLING"); tasaSamplingStr != "" {
		tasaSampling, err := strconv.ParseFloat(tasaSamplingStr, 64)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un float", "TASA_SAMPLING")
		}
		env.TasaSampling = tasaSampling
	}

	if intervaloMetricasStr := os.Getenv("INTERVALO_METRICAS"); intervaloMetricasStr != "" {
		intervaloMetricas, err := time.ParseDuration(intervaloMetricasStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un duration", "INTERVALO_METRICAS")
		}
		env.IntervaloMetricas = intervaloMetricas
	}

	// Valores de sesion, los valores son opcionales aunque la librería no
	// siempre coloca valores por default sino pretende que se use
	// `sesiones.ConfigDefault()`

	// Usar valores default primero
	configDefault := sesiones.ConfigDefault()
	env.CookieNombre = configDefault.CookieNombre
	env.CookiePath = configDefault.CookiePath
	env.CookieDominio = configDefault.CookieDominio
	env.CookieEsSecure = false
	env.CookieSoloHTTP = configDefault.CookieSoloHTTP
	env.CookieMismoSitio = configDefault.CookieMismoSitio
	env.CookieCodificar = configDefault.CookieCodificar
	env.DuracionInactiva = configDefault.DuracionInactiva
	env.DuracionMaxima = configDefault.DuracionMaxima
	env.FlashTTL = configDefault.FlashTTL
	env.MaximosMensajesFlash = configDefault.MaximosMensajesFlash
	env.HabilitarLimitePeticiones = configDefault.HabilitarLimitePeticiones
	env.PeticionesMaximasPorMinuto = configDefault.PeticionesMaximasPorMinuto
	env.IntervaloRegeneracion = configDefault.IntervaloRegeneracion
	env.HabilitarFingerprinting = configDefault.HabilitarFingerprinting
	env.SesionDuraMasDeUnDia = configDefault.SesionDuraMasDeUnDia
	env.HabilitarRedis = configDefault.HabilitarRedis

	if cookieNombre := os.Getenv("COOKIE_NOMBRE"); cookieNombre != "" {
		env.CookieNombre = cookieNombre
	}
	if cookiePath := os.Getenv("COOKIE_PATH"); cookiePath != "" {
		env.CookiePath = cookiePath
	}

	if duracionInactivaStr := os.Getenv("DURACION_INACTIVA"); duracionInactivaStr != "" {
		duracionInactiva, err := time.ParseDuration(duracionInactivaStr)
		if err != nil {
			return nil, fmt.Errorf("mal formato de %s, se esperaba un duration",
				"DURACION_INACTIVA")
		}
		env.DuracionInactiva = duracionInactiva
	}
	if duracionMaximaStr := os.Getenv("DURACION_MAXIMA"); duracionMaximaStr != "" {
		duracionMaxima, err := time.ParseDuration(duracionMaximaStr)
		if err != nil {
			return nil, fmt.Errorf("mal formato de %s, se esperaba un duration",
				"DURACION_MAXIMA")
		}
		env.DuracionMaxima = duracionMaxima
	}
	if flashTTLStr := os.Getenv("FLASH_TTL"); flashTTLStr != "" {
		flashTTL, err := time.ParseDuration(flashTTLStr)
		if err != nil {
			return nil, fmt.Errorf("mal formato de %s, se esperaba un duration",
				"FLASH_TTL")
		}
		env.FlashTTL = flashTTL
	}
	if maximosMensajesFlashStr := os.Getenv("MAXIMOS_MENSAJES_FLASH"); maximosMensajesFlashStr != "" {
		maximosMensajesFlash, err := strconv.Atoi(maximosMensajesFlashStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un int", "MAXIMOS_MENSAJES_FLASH")
		}
		env.MaximosMensajesFlash = maximosMensajesFlash
	}

	if PeticionesMaximasPorMinutoStr := os.Getenv("PETICIONES_MAXIMAS_POR_MINUTO"); PeticionesMaximasPorMinutoStr != "" {
		PeticionesMaximasPorMinuto, err := strconv.Atoi(PeticionesMaximasPorMinutoStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un int", "PETICIONES_MAXIMAS_POR_MINUTO")
		}
		env.PeticionesMaximasPorMinuto = PeticionesMaximasPorMinuto
	}
	if habilitarLimitePeticionesStr := os.Getenv("HABILITAR_LIMITE_PETICIONES"); habilitarLimitePeticionesStr != "" {
		habilitarLimitePeticiones, err := strconv.ParseBool(habilitarLimitePeticionesStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "HABILITAR_LIMITE_PETICIONES")
		}
		env.HabilitarLimitePeticiones = habilitarLimitePeticiones
	}
	if HabilitarFingerprintingStr := os.Getenv("HABILITAR_FINGERPRINTING"); HabilitarFingerprintingStr != "" {
		HabilitarFingerprinting, err := strconv.ParseBool(HabilitarFingerprintingStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "HABILITAR_FINGERPRINTING")
		}
		env.HabilitarFingerprinting = HabilitarFingerprinting
	}
	if SesionDuraMasDeUnDiaStr := os.Getenv("SESION_DURA_MAS_DE_UN_DIA"); SesionDuraMasDeUnDiaStr != "" {
		SesionDuraMasDeUnDia, err := strconv.ParseBool(SesionDuraMasDeUnDiaStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "SESION_DURA_MAS_DE_UN_DIA")
		}
		env.SesionDuraMasDeUnDia = SesionDuraMasDeUnDia
	}
	if HabilitarRedisStr := os.Getenv("HABILITAR_REDIS"); HabilitarRedisStr != "" {
		HabilitarRedis, err := strconv.ParseBool(HabilitarRedisStr)
		if err != nil {
			return nil, fmt.Errorf("error al parsear %s: se esperaba un bool", "HABILITAR_REDIS")
		}
		env.HabilitarRedis = HabilitarRedis
	}
	if IntervaloRegeneracionStr := os.Getenv("INTERVALO_REGENERACION"); IntervaloRegeneracionStr != "" {
		IntervaloRegeneracion, err := time.ParseDuration(IntervaloRegeneracionStr)
		if err != nil {
			return nil, fmt.Errorf("mal formato de %s, se esperaba un duration",
				"INTERVALO_REGENERACION")
		}
		env.IntervaloRegeneracion = IntervaloRegeneracion
	}

	// Config JWT
	configJWT := sesiones.ConfigJWTDefault()
	env.JWTAlgorithm = configJWT.Algorithm

	if JWTSecretKeyStr := os.Getenv("JWT_SECRET_KEY"); JWTSecretKeyStr != "" {
		JWTSecretKey := []byte(JWTSecretKeyStr)
		env.JWTSecretKey = JWTSecretKey
	}
	if JWTAccessTokenExpiryStr := os.Getenv("JWT_ACCESS_TOKEN_EXPIRY"); JWTAccessTokenExpiryStr != "" {
		JWTAccessTokenExpiry, err := time.ParseDuration(JWTAccessTokenExpiryStr)
		if err != nil {
			return nil, fmt.Errorf("mal formato de %s, se esperaba un duration",
				"JWT_ACCESS_TOKEN_EXPIRY")
		}
		env.JWTAccessTokenExpiry = JWTAccessTokenExpiry
	}
	if JWTRefreshTokenExpiryStr := os.Getenv("JWT_REFRESH_TOKEN_EXPIRY"); JWTRefreshTokenExpiryStr != "" {
		JWTRefreshTokenExpiry, err := time.ParseDuration(JWTRefreshTokenExpiryStr)
		if err != nil {
			return nil, fmt.Errorf("mal formato de %s, se esperaba un duration",
				"JWT_REFRESH_TOKEN_EXPIRY")
		}
		env.JWTRefreshTokenExpiry = JWTRefreshTokenExpiry
	}
	if JWTEnableRefreshTokensStr := os.Getenv("JWT_ENABLE_REFRESH_TOKENS"); JWTEnableRefreshTokensStr != "" {
		JWTEnableRefreshTokens, err := strconv.ParseBool(JWTEnableRefreshTokensStr)
		if err != nil {
			return nil, fmt.Errorf("mal formato de %s, se esperaba un duration",
				"JWT_ENABLE_REFRESH_TOKENS")
		}
		env.JWTEnableRefreshTokens = JWTEnableRefreshTokens
	}
	if JWTAudienceStr := os.Getenv("JWT_AUDIENCE"); JWTAudienceStr != "" {
		JWTAudience := strings.Split(JWTAudienceStr, ",")
		if len(JWTAudience) == 0 || slices.Contains(JWTAudience, "") {
			return nil, fmt.Errorf("mal formato de %s, formato correcto: \"aud1,aud2,aud3\" (o 1 solo valor sin comas)",
				"JWT_AUDIENCE")
		}
		env.JWTAudience = JWTAudience
	}
	if JWTAllowedClockSkewStr := os.Getenv("JWT_ALLOWED_CLOCK_SKEW"); JWTAllowedClockSkewStr != "" {
		JWTAllowedClockSkew, err := time.ParseDuration(JWTAllowedClockSkewStr)
		if err != nil {
			return nil, fmt.Errorf("mal formato de %s, se esperaba un duration",
				"JWT_ALLOWED_CLOCK_SKEW")
		}
		env.JWTAllowedClockSkew = JWTAllowedClockSkew
	}

	// Archivos

	env.DirArchivosLocal = getEnvODefault("ARCHIVOS_DIR_LOCAL", "./static")
	env.RutaPubLocal = getEnvODefault("ARCHIVOS_RUTA_PUB_LOCAL", "/static/public")
	env.RutaPrivLocal = getEnvODefault("ARCHIVOS_RUTA_PRIV_LOCAL", "/static/private")
	env.CubetaPublicaGCP = os.Getenv("ARCHIVOS_CUBETA_PUBLICA")
	env.CubetaPrivadaGCP = os.Getenv("ARCHIVOS_CUBETA_PRIVADA")
	env.UrlCDN = os.Getenv("ARCHIVOS_URL_CDN")
	env.ExpiracionURLsFirmados = parseTimeEnv("ARCHIVOS_EXPIRACION_URLs", 15*time.Minute)
	env.CacheControl = os.Getenv("ARCHIVOS_CACHE_CONTROL")

	// Configuración de Email
	env.SMTPHost = os.Getenv("SMTP_HOST")
	env.SMTPPort = os.Getenv("SMTP_PORT")
	env.SMTPUser = os.Getenv("SMTP_USER")
	env.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	env.SMTPFrom = os.Getenv("SMTP_FROM")
	env.SMTPFromName = os.Getenv("SMTP_FROM_NAME")
	env.SMTPUseTLS, _ = strconv.ParseBool(os.Getenv("SMTP_USE_TLS"))

	// Variables propias de la app

	if VariablePersonalStr := os.Getenv("VARIABLE_PERSONAL"); VariablePersonalStr != "" {
		env.VariablePersonal = VariablePersonalStr
	} else {
		env.VariablePersonal = "La VariablePersonal no fue encontrada"
	}

	return env, nil
}

type llaveCtxEnv string

const envKey llaveCtxEnv = "server_env"

func EnvMiddleware(env *Env) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), envKey, env)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func EnvDeContexto(ctx context.Context) *Env {
	if e := ctx.Value(envKey); e != nil {
		if e, ok := e.(*Env); ok {
			return e
		}
	}
	return nil
}

// utilerías para librería de archivos

func getEnvODefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseTimeEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func (e *Env) LogDebug(logger *logs.Logger) {
	jsBs, err := json.Marshal(e)
	if err != nil {
		logger.Error("Error al serializar entorno", err)
	}
	logger.Debug("Variables de entorno", "env", string(jsBs))
}

func ObtenerAmbiente(ctx context.Context) string {
	env := EnvDeContexto(ctx)
	if env == nil {
		return ""
	}
	return env.Ambiente
}

func ObtenerRequestId(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}
