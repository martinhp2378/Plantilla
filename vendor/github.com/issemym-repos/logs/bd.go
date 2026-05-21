package logs

import (
	"context"
	"strings"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AgregarTracingPgxpool se debe utilizar antes de crear un pool de pgx para
// hablitar la traza automática de procesos de la BD.
func AgregarTracingPgxpool(poolConfig *pgxpool.Config,
	config *Config, i *Instrumentacion) {

	opts := []otelpgx.Option{
		otelpgx.WithTracerProvider(i.tracerProvider),
		otelpgx.WithTracerAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("service.name", config.Servicio),
		),
	}

	if config.NombreSpanCorto {
		opts = append(opts, otelpgx.WithTrimSQLInSpanName())
	}

	if !config.QuitarArgsQuery && !config.QuitarStatement {
		opts = append(opts, otelpgx.WithIncludeQueryParameters())
	}

	if config.QuitarPrefijoQuery {
		opts = append(opts, otelpgx.WithDisableQuerySpanNamePrefix())
	}

	if config.QuitarStatement {
		opts = append(opts, otelpgx.WithDisableSQLStatementInAttributes())
	}

	// Ver tracerConGrabado y TraceQueryStart
	tracerBase := otelpgx.NewTracer(opts...)
	poolConfig.ConnConfig.Tracer = &tracerConGrabado{
		Tracer:          tracerBase,
		instrumentacion: i,
	}
}

func HabilitarEstadisticas(pool *pgxpool.Pool, config *Config) error {
	return otelpgx.RecordStats(pool, otelpgx.WithStatsAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("service.name", config.Servicio),
	))
}

func ExtraerOp(sql string) string {
	sql = strings.TrimSpace(strings.ToLower(sql))

	ops := []string{"select", "insert", "update", "delete", "create", "drop",
		"alter", "truncate", "explain", "analyze", "with"}
	for _, op := range ops {
		if strings.HasPrefix(sql, op) {
			return op
		}
	}

	return "other"
}

// un wrapper para la instrumentación para evitar que el usuario tenga
// que crear un span cada vez que quiera instrumentar la BD.
type tracerConGrabado struct {
	*otelpgx.Tracer
	instrumentacion *Instrumentacion
}

// ----------------------------------------------------------------------------
// Estos métodos son reimplementaciones de otelpgx, es necesario porque se
// implementa nuestra lógica de crear un recording span de forma automática.
// Pero se tiene que evitar que la librería cree sus propios spans.
//
// NOTA: otelpgx puede crear múltiples spans para una sola operación
// de base de datos. Esto es normal, los spans representan:
//
// 1. Preparación de query (si es necesario) - span de nivel alto
// 2. Ejecución de query - span de nivel bajo con detalles de ejecución
//
// Por ejemplo, un pool.QueryRow() puede generar:
// - span "pool.acquire" (adquisición de conexión)
// - span "SELECT" prepare (sin parámetros de query)
// - span "SELECT" execute (con parámetros y rows_affected)
//
// Esto proporciona observabilidad completa del ciclo de vida de la operación,

// Llave para marcar que ya creamos un span con recording en este contexto de
// operación
type autoSpanKey struct{}

// crear un auto-span solo una vez por contexto
func (tg *tracerConGrabado) asegurarSpanGrabado(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)

	// ya existe un auto-span posiblemente creado por el usuario
	if span.IsRecording() {
		return ctx
	}

	// ya creamos un auto-span nosotros
	if ctx.Value(autoSpanKey{}) != nil {
		return ctx
	}

	nuevoCtx, _ := tg.instrumentacion.IniciarSpan(ctx, "db.operation")
	return context.WithValue(nuevoCtx, autoSpanKey{}, true)
}

func (tg *tracerConGrabado) TraceQueryStart(ctx context.Context,
	conn *pgx.Conn, datos pgx.TraceQueryStartData) context.Context {

	ctx = tg.asegurarSpanGrabado(ctx)
	return tg.Tracer.TraceQueryStart(ctx, conn, datos)
}

func (tg *tracerConGrabado) TraceCopyFromStart(ctx context.Context,
	conn *pgx.Conn,
	datos pgx.TraceCopyFromStartData) context.Context {

	ctx = tg.asegurarSpanGrabado(ctx)
	return tg.Tracer.TraceCopyFromStart(ctx, conn, datos)
}

func (tg *tracerConGrabado) TraceBatchStart(ctx context.Context,
	conn *pgx.Conn,
	datos pgx.TraceBatchStartData) context.Context {

	ctx = tg.asegurarSpanGrabado(ctx)
	return tg.Tracer.TraceBatchStart(ctx, conn, datos)
}

func (tg *tracerConGrabado) TraceConnectStart(ctx context.Context,
	datos pgx.TraceConnectStartData) context.Context {

	// No crear auto-span para conexión - esto es solo infraestructura
	return tg.Tracer.TraceConnectStart(ctx, datos)
}

func (tg *tracerConGrabado) TracePrepareStart(ctx context.Context,
	conn *pgx.Conn,
	datos pgx.TracePrepareStartData) context.Context {

	// No crear auto-span para prepare - esto es parte de la operación de query
	return tg.Tracer.TracePrepareStart(ctx, conn, datos)
}

func (tg *tracerConGrabado) TraceAcquireStart(ctx context.Context,
	pool *pgxpool.Pool, datos pgxpool.TraceAcquireStartData) context.Context {

	// No crear auto-span para acquire - esto es solo infraestructura
	// El span se creará en la operación real (Query, Exec, etc.)
	return tg.Tracer.TraceAcquireStart(ctx, pool, datos)
}
