package app

import (
	perfil_handler "PlantillaInstitucional/configuracion/perfil/handler"
	perfil_repo "PlantillaInstitucional/configuracion/perfil/repository"
	perfil_service "PlantillaInstitucional/configuracion/perfil/service"
	"PlantillaInstitucional/configuracion/policy"
	policy_handler "PlantillaInstitucional/configuracion/policy/handler"
	policy_repo "PlantillaInstitucional/configuracion/policy/repository"
	usuario_handler "PlantillaInstitucional/configuracion/usuario/handler"
	usuario_repo "PlantillaInstitucional/configuracion/usuario/repository"
	usuario_service "PlantillaInstitucional/configuracion/usuario/service"
	persona_handler "PlantillaInstitucional/institucion/handler"
	persona_repo "PlantillaInstitucional/institucion/repository"
	persona_service "PlantillaInstitucional/institucion/service"
	login_repo "PlantillaInstitucional/login/db"
	login_handler "PlantillaInstitucional/login/handlers"
	"PlantillaInstitucional/pagination"
	parametros_handler "PlantillaInstitucional/parametros/handler"
	parametros_repo "PlantillaInstitucional/parametros/repository"
	parametros_service "PlantillaInstitucional/parametros/service"
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

	ConstructorSesion    *sesiones.ConstructorSesion
	ConstructorSesionJWT *sesiones.ConstructorSesionJWT
	ManejadorArchivos    *archivos.ManejadorArchivos
	ServicioEmail        *email.EmailService
	EventDispatcher      *sharedEvents.Dispatcher

	Handlers     Handlers
	Services     Services
	Repositories Repositories
	TenantPolicy securitypolicy.TenantPolicy

	tenants   map[string]string
	tenantsMu sync.RWMutex
}

type Handlers struct {
	Usuario   *usuario_handler.UsuarioHandler
	Perfil    *perfil_handler.PerfilHandler
	Policy    *policy_handler.PolicyHandler
	Persona   persona_handler.PersonaHandler
	Login     *login_handler.HandlerLogin
	Parametro *parametros_handler.Handler
}

type PolicyResolver interface {
	IsRouteAllowed(ctx context.Context, profile string, path string) (bool, error)
	IsProfileActive(ctx context.Context, profile string) (bool, error)
	ListarPoliticas(ctx context.Context, criterios pagination.CriteriosPagina[policy.Politica]) (pagination.CriteriosPagina[policy.Politica], error)
	ObtenerPolitica(ctx context.Context, nombre string) (policy.Politica, error)
	CrearPolitica(ctx context.Context, p policy.Politica) error
	ActualizarPolitica(ctx context.Context, p policy.Politica, nombreOriginal string) error
	EliminarPolitica(ctx context.Context, nombre string) error
	CatalogosPolitica(ctx context.Context) ([]string, []string, []string, error)
}

type Services struct {
	Usuario   usuario_service.UsuarioServiceProvider
	Perfil    perfil_service.PerfilServiceProvider
	Policy    PolicyResolver
	Persona   persona_service.PersonaService
	Parametro parametros_service.ParametrosServiceProvider
}

type Repositories struct {
	Tenants   tenants.Repository
	Usuario   *usuario_repo.UsuarioRepository
	Perfil    *perfil_repo.PerfilRepository
	Policy    *policy_repo.PolicyRepository
	Persona   *persona_repo.PersonaRepository
	Login     *login_repo.RepositorioPsql
	Parametro *parametros_repo.ParametrosRepository
}

func NewApp(ctx context.Context, env *entorno.Env,
	pqWrite *pgxpool.Pool, pqRead *pgxpool.Pool, conexionRedis *redis.Client,
	logger *logs.Logger, instrumentacion *logs.Instrumentacion,
	constructorSesion *sesiones.ConstructorSesion, constructorSesionJWT *sesiones.ConstructorSesionJWT,
	manejadorArchivos *archivos.ManejadorArchivos, servicioEmail *email.EmailService) (*App, error) {
	return NewAppWithTenantPolicy(ctx, env, pqWrite, pqRead, conexionRedis, logger, instrumentacion,
		constructorSesion, constructorSesionJWT, manejadorArchivos, servicioEmail, defaultTenantPolicy())
}

func NewAppWithTenantPolicy(ctx context.Context, env *entorno.Env,
	pqWrite *pgxpool.Pool, pqRead *pgxpool.Pool, conexionRedis *redis.Client,
	logger *logs.Logger, instrumentacion *logs.Instrumentacion,
	constructorSesion *sesiones.ConstructorSesion, constructorSesionJWT *sesiones.ConstructorSesionJWT,
	manejadorArchivos *archivos.ManejadorArchivos, servicioEmail *email.EmailService, tenantPolicy securitypolicy.TenantPolicy) (*App, error) {
	if err := tenantPolicy.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tenant policy: %w", err)
	}
	return &App{Env: env, Logger: logger, Instrumentacion: instrumentacion, ConexionRedis: conexionRedis,
		PqWrite: pqWrite, PqRead: pqRead, ConstructorSesion: constructorSesion,
		ConstructorSesionJWT: constructorSesionJWT, ManejadorArchivos: manejadorArchivos,
		ServicioEmail: servicioEmail, TenantPolicy: tenantPolicy, tenants: make(map[string]string)}, nil
}

func defaultTenantPolicy() securitypolicy.TenantPolicy {
	return securitypolicy.TenantPolicy{EnforcementEnabled: true,
		CanonicalHosts: []securitypolicy.CanonicalHostRule{{Host: "localhost", TenantID: "local", Schema: "local", Active: true}},
		AllowedSchemes: []string{"http", "https"}, GlobalAccessPolicyID: "global-readers-v1"}
}

func (a *App) Close() {
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
