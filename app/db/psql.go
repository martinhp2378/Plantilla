package bd

import (
	"context"
	"errors"
	"fmt"
	"log"
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

	contextkeys "PlantillaInstitucional/shared/contextkey"
	"PlantillaInstitucional/shared/entorno"
)

var tenantSchemaRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type AppPools struct {
	ReadPool  *pgxpool.Pool
	WritePool *pgxpool.Pool
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

	// Agregar extension uuid
	readPoolConfig.AfterConnect = func(ctx context.Context, pgxConn *pgx.Conn) error {
		pgxuuid.Register(pgxConn.TypeMap())
		return nil
	}

	// Multi-tenant: Cambiar el esquema antes de adquirir la conexión
	readPoolConfig.BeforeAcquire = func(ctx context.Context, pgxConn *pgx.Conn) bool {
		schema := getTenantSchema(ctx)
		searchPath, err := buildTenantSearchPath(schema)
		if err != nil {
			log.Printf("[WARN-DB-READ] tenant schema inválido; usando fallback seguro")
			return false
		}
		_, err = pgxConn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", searchPath))
		return err == nil
	}

	logsConfigRead := &logs.Config{
		Servicio:           env.Servicio,
		NombreSpanCorto:    env.NombreSpanCorto,    // Usar nombres de span corto (SELECT vs SQL completo)
		QuitarPrefijoQuery: env.QuitarPrefijoQuery, // Por default se agrega como prefijo "query" o "batch query"
		QuitarArgsQuery:    env.QuitarArgsQuery,    // Incluir los valores de los parámetros en los spans
		QuitarStatement:    env.QuitarStatement,    // Incluir el statement en los atributos
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

	// Multi-tenant: Cambiar el esquema antes de adquirir la conexión
	writePoolConfig.BeforeAcquire = func(ctx context.Context, pgxConn *pgx.Conn) bool {
		schema := getTenantSchema(ctx)
		searchPath, err := buildTenantSearchPath(schema)
		if err != nil {
			log.Printf("[WARN-DB-WRITE] tenant schema inválido; usando fallback seguro")
			return false
		}
		_, err = pgxConn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", searchPath))
		return err == nil
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
		ReadPool:  readPool,
		WritePool: writePool,
	}, nil
}

func getTenantSchema(ctx context.Context) string {
	if v := ctx.Value(contextkeys.TenantSchemaKey); v != nil {
		if schema, ok := v.(string); ok {
			return schema
		}
	}
	return "public"
}

// buildTenantSearchPath centraliza la construcción segura de search_path tenant-aware.
func buildTenantSearchPath(schema string) (string, error) {
	s := strings.TrimSpace(schema)
	if s == "" {
		s = "public"
	}
	if !tenantSchemaRegex.MatchString(s) {
		return "", fmt.Errorf("tenant schema inválido")
	}
	return fmt.Sprintf("%s, sesion, institucion, sistema, public", pgx.Identifier{s}.Sanitize()), nil
}
