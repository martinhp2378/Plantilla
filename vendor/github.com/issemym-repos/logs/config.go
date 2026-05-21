package logs

import "time"

type Ambiente string

const (
	AmbienteLocal      Ambiente = "LOCAL"
	AmbienteDesarrollo Ambiente = "DESARROLLO"
	AmbienteProduccion Ambiente = "PRODUCCION"
)

type Config struct {
	// nombre del app; ej: "portal"
	Servicio string
	// Ej. "1.0.0"
	Version  string
	Ambiente Ambiente

	// id del proyecto en gcp
	IdProyecto string

	LogsConColores bool      // para usar colores en logs (solo local)
	AgregarSource  bool      // para incluir ruta archivo fuente
	Nivel          Severidad // nivel mínimo del log

	ArchivoCredenciales string

	// Instrumentación BD

	// Solo usar la primer palabra del statement (select, insert, etc..) como
	// nombre del span en lugar de todo el statement
	NombreSpanCorto bool

	// los parámetros del query se agregan al span en el atributo
	// `pgx.query.parameters`, este atributo deshabilita esa funcionalidad.
	// Este atributo es implícitamente `true` si `QuitarStatement` es `true`.
	QuitarArgsQuery bool
	// Por default el span tiene un prefijo: query o batch query, este parámetro
	// lo quita
	QuitarPrefijoQuery bool
	// Deshabilitar loguear el statement en los atributos del span.
	QuitarStatement bool

	// Instrumentación

	TasaSampling                 float64
	IntervaloMetricas            time.Duration
	HabilitarMetricasAutomaticas bool

	// Instrumentación en ambiente local

	// NOTA: Si se pretende usar middleware HTTP (MiddlewareOtel), tanto el
	// tracing como las métricas deben estar habilitados. El middleware HTTP
	// de OpenTelemetry requiere ambos componentes.
	//
	// Para una configuración compatible con HTTP middleware, se debe usar
	// NuevaInstrumentacionParaHTTP() en lugar de NuevaInstrumentacion().
	DeshabilitarTracing   bool
	DeshabilitarMetricas  bool
	UsarExportadorConsola bool
}

// para almacenar el logger en el Context
type llaveContexto string

const (
	llaveLogger          llaveContexto = "logger"
	llaveTracer          llaveContexto = "tracer"
	llaveInstrumentacion llaveContexto = "instrumentacion"
)
