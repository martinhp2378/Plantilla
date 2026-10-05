package bd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/issemym-repos/logs"
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"PlantillaInstitucional/shared/entorno"
	"PlantillaInstitucional/shared/tenantcontext"
)

var tenantSchemaRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type AppPools struct {
	readPool  *pgxpool.Pool
	writePool *pgxpool.Pool
	Tenant    *TenantTxRunner
}

func NewAppPools(readPool, writePool *pgxpool.Pool) *AppPools {
	return &AppPools{
		readPool:  readPool,
		writePool: writePool,
		Tenant:    NewTenantTxRunner(readPool, writePool),
	}
}

func (a *AppPools) ReadPool() *pgxpool.Pool {
	return a.readPool
}

func (a *AppPools) WritePool() *pgxpool.Pool {
	return a.writePool
}


func BuildConnectionURL(env *entorno.Env, isWrite bool) string {
	user := env.PqWriteUser
	pass := env.PqWritePass
	host := env.PqWriteHost
	port := env.PqWritePort
	db := env.PqWriteDb
	ssl := env.PqWriteSSLMode

	if !isWrite {
		user = env.PqReadUser
		pass = env.PqReadPass
		host = env.PqReadHost
		port = env.PqReadPort
		db = env.PqReadDb
		ssl = env.PqReadSSLMode
	}

	if strings.HasPrefix(host, "/cloudsql/") {
		return fmt.Sprintf("postgres://%s:%s@localhost/%s?host=%s&sslmode=%s",
			url.QueryEscape(user), url.QueryEscape(pass), db, url.QueryEscape(host), ssl)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(user), url.QueryEscape(pass), host, port, db, ssl)
}

func RunMigrations(env *entorno.Env) error {
	connURL := BuildConnectionURL(env, true)

	m, err := migrate.New(
		"file://sql/migrations",
		connURL,
	)
	if err != nil {
		return fmt.Errorf("error al instanciar migración: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("error al ejecutar migraciones: %w", err)
	}

	return nil
}

func ConnectPostgres(ctx context.Context, env *entorno.Env, instrumentacion *logs.Instrumentacion) (*AppPools, error) {
	urlRead := BuildConnectionURL(env, false)
	readPoolConfig, err := pgxpool.ParseConfig(urlRead)
	if err != nil {
		return nil, fmt.Errorf("error al parsear url de read: %w", err)
	}
	readPoolConfig.MaxConns = env.PqReadMaxConns
	readPoolConfig.MinConns = env.PqReadMinConns
	readPoolConfig.MaxConnIdleTime = env.PqReadMaxConnIdle
	readPoolConfig.MaxConnLifetime = env.PqReadMaxConnLife
	readPoolConfig.HealthCheckPeriod = 1 * time.Minute

	readPoolConfig.ConnConfig.RuntimeParams["timezone"] = "America/Mexico_City"
	readPoolConfig.ConnConfig.RuntimeParams["lock_timeout"] = "10s"

	readPoolConfig.AfterConnect = func(ctx context.Context, pgxConn *pgx.Conn) error {
		pgxuuid.Register(pgxConn.TypeMap())
		return nil
	}

	logsConfigRead := &logs.Config{
		Servicio:           env.Servicio,
		NombreSpanCorto:    env.NombreSpanCorto,
		QuitarPrefijoQuery: env.QuitarPrefijoQuery,
		QuitarArgsQuery:    env.QuitarArgsQuery,
		QuitarStatement:    env.QuitarStatement,
	}

	logs.AgregarTracingPgxpool(readPoolConfig, logsConfigRead, instrumentacion)

	readPool, err := pgxpool.NewWithConfig(ctx, readPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("error al crear pool de read: %w", err)
	}

	err = readPool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al hacer ping al pool de read: %w", err)
	}

	urlWrite := BuildConnectionURL(env, true)
	writePoolConfig, err := pgxpool.ParseConfig(urlWrite)
	if err != nil {
		return nil, fmt.Errorf("error al parsear url de write: %w", err)
	}
	writePoolConfig.MaxConns = env.PqWriteMaxConns
	writePoolConfig.MinConns = env.PqWriteMinConns
	writePoolConfig.MaxConnIdleTime = env.PqWriteMaxConnIdle
	writePoolConfig.MaxConnLifetime = env.PqWriteMaxConnLife
	writePoolConfig.HealthCheckPeriod = 1 * time.Minute

	writePoolConfig.ConnConfig.RuntimeParams["timezone"] = "America/Mexico_City"
	writePoolConfig.ConnConfig.RuntimeParams["lock_timeout"] = "10s"

	// Agregar extension uuid
	writePoolConfig.AfterConnect = func(ctx context.Context, pgxConn *pgx.Conn) error {
		pgxuuid.Register(pgxConn.TypeMap())
		return nil
	}

	logsConfigWrite := &logs.Config{
		Servicio:           env.Servicio,
		NombreSpanCorto:    env.NombreSpanCorto,
		QuitarPrefijoQuery: env.QuitarPrefijoQuery,
		QuitarArgsQuery:    env.QuitarArgsQuery,
		QuitarStatement:    env.QuitarStatement,
	}

	logs.AgregarTracingPgxpool(writePoolConfig, logsConfigWrite, instrumentacion)

	writePool, err := pgxpool.NewWithConfig(ctx, writePoolConfig)
	if err != nil {
		return nil, fmt.Errorf("error al crear pool de write: %w", err)
	}

	err = writePool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al hacer ping al pool de write: %w", err)
	}

	return &AppPools{
		readPool:  readPool,
		writePool: writePool,
		Tenant:    NewTenantTxRunner(readPool, writePool),
	}, nil
}

func getTenantSchema(ctx context.Context) (string, error) {
	tenant, err := tenantcontext.TenantFromContext(ctx)
	if err != nil {
		return "", err
	}
	return tenant.Schema, nil
}

// buildTenantSearchPath centraliza la construcción segura de search_path tenant-aware.
func buildTenantSearchPath(schema string) (string, error) {
	s := strings.TrimSpace(schema)
	if s == "" || s == "public" || !tenantSchemaRegex.MatchString(s) {
		return "", fmt.Errorf("tenant schema inválido")
	}
	return fmt.Sprintf("%s, institucion, sesion, sistema, pg_catalog", pgx.Identifier{s}.Sanitize()), nil
}
