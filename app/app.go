package app

import (
	bd "PlantillaInstitucional/app/db"
	"PlantillaInstitucional/shared/email"
	"PlantillaInstitucional/shared/entorno"
	sharedEvents "PlantillaInstitucional/shared/events"
	"PlantillaInstitucional/shared/securitypolicy"
	"PlantillaInstitucional/shared/tenants"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/issemym-repos/archivos"
	"github.com/issemym-repos/logs"
	"github.com/issemym-repos/sesiones"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Env             *entorno.Env
	Logger          *logs.Logger
	Instrumentacion *logs.Instrumentacion
	ConexionRedis   *redis.Client
	PqWrite         *pgxpool.Pool
	PqRead          *pgxpool.Pool
	Pools           *bd.AppPools

	ConstructorSesion    *sesiones.ConstructorSesion
	ConstructorSesionJWT *sesiones.ConstructorSesionJWT
	ManejadorArchivos    *archivos.ManejadorArchivos
	ServicioEmail        *email.EmailService
	EventDispatcher      *sharedEvents.Dispatcher

	Services     Services
	Repositories Repositories
	TenantPolicy securitypolicy.TenantPolicy

	tenants   map[string]string
	tenantsMu sync.RWMutex
}

type PolicyResolver interface {
	IsRouteAllowed(ctx context.Context, profile string, path string) (bool, error)
	IsProfileActive(ctx context.Context, profile string) (bool, error)
}

type Services struct {
	Policy PolicyResolver
}

type Repositories struct {
	Tenants tenants.Repository
}

func NewApp(ctx context.Context, env *entorno.Env,
	pqWrite *pgxpool.Pool, pqRead *pgxpool.Pool, conexionRedis *redis.Client,
	logger *logs.Logger, instrumentacion *logs.Instrumentacion,
	constructorSesion *sesiones.ConstructorSesion, constructorSesionJWT *sesiones.ConstructorSesionJWT,
	manejadorArchivos *archivos.ManejadorArchivos, servicioEmail *email.EmailService,
	policyResolvers ...PolicyResolver) (*App, error) {
	tenantPolicy, err := tenantPolicyFromEnvironment(env, timeNow)
	if err != nil {
		return nil, err
	}
	return NewAppWithTenantPolicy(ctx, env, pqWrite, pqRead, conexionRedis, logger, instrumentacion,
		constructorSesion, constructorSesionJWT, manejadorArchivos, servicioEmail, tenantPolicy, policyResolvers...)
}

func NewAppWithTenantPolicy(ctx context.Context, env *entorno.Env,
	pqWrite *pgxpool.Pool, pqRead *pgxpool.Pool, conexionRedis *redis.Client,
	logger *logs.Logger, instrumentacion *logs.Instrumentacion,
	constructorSesion *sesiones.ConstructorSesion, constructorSesionJWT *sesiones.ConstructorSesionJWT,
	manejadorArchivos *archivos.ManejadorArchivos, servicioEmail *email.EmailService, tenantPolicy securitypolicy.TenantPolicy,
	policyResolvers ...PolicyResolver) (*App, error) {
	if err := tenantPolicy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tenant policy: %w", err)
	}
	pools := bd.NewAppPools(pqRead, pqWrite)

	tenantsRepo := tenants.NewStaticRepository()
	dispatcher := sharedEvents.NewDispatcher()

	var policyResolver PolicyResolver
	if len(policyResolvers) > 0 {
		policyResolver = policyResolvers[0]
	}

	return &App{
		Env:                  env,
		Logger:               logger,
		Instrumentacion:      instrumentacion,
		ConexionRedis:        conexionRedis,
		PqWrite:              pqWrite,
		PqRead:               pqRead,
		Pools:                pools,
		ConstructorSesion:    constructorSesion,
		ConstructorSesionJWT: constructorSesionJWT,
		ManejadorArchivos:    manejadorArchivos,
		ServicioEmail:        servicioEmail,
		EventDispatcher:      dispatcher,
		Services: Services{
			Policy: policyResolver,
		},
		Repositories: Repositories{
			Tenants: tenantsRepo,
		},
		TenantPolicy: tenantPolicy,
		tenants:      make(map[string]string),
	}, nil
}

func (a *App) Close() {
	if a.ManejadorArchivos != nil {
		_ = a.ManejadorArchivos.Close()
	}
	if a.PqWrite != nil {
		a.PqWrite.Close()
	}
	if a.PqRead != nil {
		a.PqRead.Close()
	}
}

// GetTenantSchema busca el esquema asociado a un subdominio de forma segura.
func (a *App) GetTenantSchema(subdomain string) string {
	a.tenantsMu.RLock()
	defer a.tenantsMu.RUnlock()

	if schema, ok := a.tenants[subdomain]; ok {
		return schema
	}

	if a.Env != nil && a.Env.Ambiente == entorno.AmbienteLocal {
		return "public"
	}

	// Fallback por variable de entorno si el equipo lo necesita
	if strings.EqualFold(strings.TrimSpace(os.Getenv("ALLOW_PUBLIC_TENANT_FALLBACK")), "true") {
		return "public"
	}

	return ""
}

// RefrescarTenants recarga el mapa de tenants desde la base de datos.
func (a *App) RefrescarTenants(ctx context.Context) error {
	tenantMap, err := a.Repositories.Tenants.ListarTenantsActivos(ctx)
	if err != nil {
		return err
	}

	a.tenantsMu.Lock()
	a.tenants = tenantMap
	a.tenantsMu.Unlock()

	a.Logger.InfoContext(ctx, "Mapa de tenants actualizado", "total", len(tenantMap))
	return nil
}
