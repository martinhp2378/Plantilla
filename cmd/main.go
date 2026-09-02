package main

import (
	"PlantillaInstitucional/app"
	bd "PlantillaInstitucional/app/db"
	app_middleware "PlantillaInstitucional/app/middleware"
	"PlantillaInstitucional/cmd/config"
	"PlantillaInstitucional/shared/email"
	"PlantillaInstitucional/shared/entorno"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
	"github.com/issemym-repos/sesiones"
	"github.com/redis/go-redis/v9"
)

func main() {
	slog.Info("Iniciando aplicación institucional...")

	env, err := entorno.GetEnv(".env")
	if err != nil {
		slog.Error("error al obtener variables de ambiente", "error", err)
		os.Exit(1)
	}

	if isMigrationCommand(os.Args[1:]) {
		slog.Info("Ejecutando migraciones de base de datos...")
		if err := bd.RunMigrations(env); err != nil {
			slog.Error("error ejecutando migraciones", "error", err)
			os.Exit(1)
		}
		slog.Info("Migraciones de base de datos finalizadas")
		return
	}

	configInstrumentacion := configInstrumentacion(env)
	logger := logs.NuevoLogger(configInstrumentacion)

	instrumentacion, err := logs.NuevaInstrumentacionParaHTTP(configInstrumentacion)
	if err != nil {
		logger.Fatal("Error inicializando instrumentación", err)
	}
	defer func() {
		if err := instrumentacion.Shutdown(context.Background()); err != nil {
			logger.Error("error al hacer shutdown a la instrumentación", err)
		}
	}()

	pools, err := bd.ConnectPostgres(context.Background(), env, instrumentacion)
	if err != nil {
		logger.Fatal("error al conectar a postgres", err)
	}

	if !env.SkipMigrations {
		logger.Info("Ejecutando migraciones de base de datos...")
		if err := bd.RunMigrations(env); err != nil {
			logger.Fatal("error ejecutando migraciones", err)
		}
	} else {
		logger.Info("SKIP_MIGRATIONS habilitado, omitiendo ejecución de migraciones")
	}

	pqWrite := pools.WritePool()
	pqRead := pools.ReadPool()

	var conexionRedis *redis.Client
	if env.RedisHost != "" {
		conexionRedis = redis.NewClient(&redis.Options{
			Addr:     env.RedisHost + ":" + env.RedisPort,
			Password: env.RedisPassword,
			DB:       env.RedisDB,
		})
	}

	cfgArchivos := configArchivos(env)
	manejadorArchivos, err := archivos.NuevoManejador(cfgArchivos)
	if err != nil {
		logger.Fatal("error al instanciar el manejador de archivos", err)
	}

	emailCfg := configEmail(env)
	emailService := email.NewEmailService(emailCfg)

	var constructorSesion *sesiones.ConstructorSesion
	var constructorSesionJWT *sesiones.ConstructorSesionJWT
	regTipos := sesiones.NuevoRegistroTipos()
	fnObtenerDatos := func(u uuid.UUID, p sesiones.Perfil, m ...map[string]any) []any {
		return []any{}
	}
	constructorSesion = sesiones.NuevoConstructor(pqWrite, configSesiones(env), regTipos, fnObtenerDatos)
	constructorSesionJWT, _ = sesiones.NuevoConstructorJWT(pqWrite, configSesiones(env), configJWT(env), regTipos, fnObtenerDatos)

	a, err := app.NewApp(
		context.Background(),
		env,
		pqWrite,
		pqRead,
		conexionRedis,
		logger,
		instrumentacion,
		constructorSesion,
		constructorSesionJWT,
		manejadorArchivos,
		emailService)

	if err != nil {
		logger.Fatal("error al crear app", err)
	}

	config.InitRepositories(a)
	config.InitServices(a)
	config.WireResolvers(a)
	config.InitHandlers(a)


	r := chi.NewRouter()

	app_middleware.RegistrarMiddlewaresApp(r, a)

	a.Handlers.Login.RegisterRoutes(r)

	if env.Ambiente == entorno.AmbienteLocal {
		staticDir := env.DirArchivosLocal
		r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir(staticDir))))
	}

	r.Group(func(r chi.Router) {
		r.Use(app_middleware.MiddlewareRequiereLogin)
		r.Use(app_middleware.PolicyMiddleware(a))

		a.Handlers.Usuario.RegisterRoutes(r)
		a.Handlers.Perfil.RegisterRoutes(r)
		a.Handlers.Policy.RegisterRoutes(r)
		a.Handlers.Persona.RegisterRoutes(r)
		a.Handlers.Parametro.RegisterRoutes(r)
	})

	app_middleware.RegistrarHandler404(r)

	srv := &http.Server{
		Addr:    env.ServerListen,
		Handler: r,
	}

	go func() {
		logger.Info("levantando servidor", "puerto", env.ServerListen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("error al servir http", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutDown()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		logger.Fatal("Server forced to shutdown", err)
	}

	logger.Info("Server exiting")
}

func isMigrationCommand(args []string) bool {
	return len(args) == 1 && args[0] == "migrate"
}

func configInstrumentacion(env *entorno.Env) *logs.Config {
	return &logs.Config{
		Servicio:                     env.Servicio,
		IdProyecto:                   env.IdProyecto,
		Version:                      env.Version,
		Ambiente:                     logs.Ambiente(env.Ambiente),
		LogsConColores:               env.LogsConColores,
		AgregarSource:                env.AgregarSource,
		Nivel:                        env.Nivel,
		NombreSpanCorto:              env.NombreSpanCorto,
		QuitarArgsQuery:              env.QuitarArgsQuery,
		QuitarPrefijoQuery:           env.QuitarPrefijoQuery,
		QuitarStatement:              env.QuitarStatement,
		TasaSampling:                 env.TasaSampling,
		IntervaloMetricas:            env.IntervaloMetricas,
		HabilitarMetricasAutomaticas: env.HabilitarMetricasAutomaticas,
		DeshabilitarTracing:          env.DeshabilitarTracing,
		DeshabilitarMetricas:         env.DeshabilitarMetricas,
		UsarExportadorConsola:        env.UsarExportadorConsola,
	}
}

func configEmail(env *entorno.Env) email.Config {
	return email.Config{
		Host:     env.SMTPHost,
		Port:     env.SMTPPort,
		User:     env.SMTPUser,
		Password: env.SMTPPassword,
		From:     env.SMTPFrom,
		FromName: env.SMTPFromName,
		UseTLS:   env.SMTPUseTLS,
	}
}

func configSesiones(env *entorno.Env) *sesiones.ConfigSesion {
	return &sesiones.ConfigSesion{
		CookieNombre:               env.CookieNombre,
		CookiePath:                 env.CookiePath,
		CookieDominio:              env.CookieDominio,
		CookieEsSecure:             env.CookieEsSecure,
		CookieSoloHTTP:             env.CookieSoloHTTP,
		CookieMismoSitio:           env.CookieMismoSitio,
		CookieCodificar:            env.CookieCodificar,
		DuracionInactiva:           env.DuracionInactiva,
		DuracionMaxima:             env.DuracionMaxima,
		FlashTTL:                   env.FlashTTL,
		MaximosMensajesFlash:       env.MaximosMensajesFlash,
		HabilitarLimitePeticiones:  env.HabilitarLimitePeticiones,
		PeticionesMaximasPorMinuto: env.PeticionesMaximasPorMinuto,
		IntervaloRegeneracion:      env.IntervaloRegeneracion,
		HabilitarFingerprinting:    env.HabilitarFingerprinting,
		SesionDuraMasDeUnDia:       env.SesionDuraMasDeUnDia,
		HabilitarRedis:             env.HabilitarRedis,
	}
}

func configArchivos(env *entorno.Env) *archivos.Config {
	return &archivos.Config{
		Ambiente:               archivos.Ambiente(env.Ambiente),
		DirArchivosLocal:       env.DirArchivosLocal,
		RutaPubLocal:           env.RutaPubLocal,
		RutaPrivLocal:          env.RutaPrivLocal,
		IdProyectoGCP:          env.IdProyecto,
		CubetaPublicaGCP:       env.CubetaPublicaGCP,
		CubetaPrivadaGCP:       env.CubetaPrivadaGCP,
		Servicio:               env.Servicio,
		UrlCDN:                 env.UrlCDN,
		NombreURLMapCDN:        env.NombreURLMapCDN,
		ExpiracionURLsFirmados: env.ExpiracionURLsFirmados,
		CacheControl:           env.CacheControl,
	}
}

func configJWT(env *entorno.Env) *sesiones.ConfigJWT {
	return &sesiones.ConfigJWT{
		SecretKey:           env.JWTSecretKey,
		Algorithm:           env.JWTAlgorithm,
		AccessTokenExpiry:   env.JWTAccessTokenExpiry,
		RefreshTokenExpiry:  env.JWTRefreshTokenExpiry,
		EnableRefreshTokens: env.JWTEnableRefreshTokens,
		Issuer:              env.Servicio,
		Audience:            env.JWTAudience,
		AllowedClockSkew:    env.JWTAllowedClockSkew,
	}
}
