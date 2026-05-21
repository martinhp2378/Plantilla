package logs

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gcpmetric "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	gcpexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/option"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Instrumentacion struct {
	config         *Config
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *metricSdk.MeterProvider
	tracer         trace.Tracer
	meter          metric.Meter

	// métricas HTTP

	contadorPeticiones metric.Int64Counter
	duracionPeticion   metric.Float64Histogram
	peticionesActivas  metric.Int64UpDownCounter

	// NOTA: Las métricas de BD son proporcionadas automáticamente por otelpgx
	// a través de HabilitarEstadisticas() y no necesitan ser definidas aquí.
	// TODO: confirmar qué métricas da
	// las métricas que proporciona son:
	// - db.client.connections.usage - Tiempo de uso de cada conexión
	// - db.client.connections.idle - Número de conexiones inactivas
	// - db.client.connections.max - Conexiones máximas del pool
	// - db.client.connections.pending_requests - Peticiones pendientes
}

// NuevaInstrumentacionParaHTTP crea una nueva instancia de Instrumentacion
// y valida que la configuración sea compatible con el middleware HTTP.
//
// Esta es una función de conveniencia que combina NuevaInstrumentacion
// con ValidarConfiguracionParaHTTP.
func NuevaInstrumentacionParaHTTP(config *Config) (*Instrumentacion, error) {
	i, err := NuevaInstrumentacion(config)
	if err != nil {
		return nil, err
	}

	if err := i.ValidarConfiguracionParaHTTP(); err != nil {
		_ = i.Shutdown(context.Background())
		return nil, fmt.Errorf("configuración inválida para middleware HTTP: %w", err)
	}

	return i, nil
}

func NuevaInstrumentacion(config *Config) (*Instrumentacion, error) {
	// defaults

	if config.Ambiente == "" {
		config.Ambiente = AmbienteLocal
	}
	if config.Servicio == "" {
		config.Servicio = "desconocido"
	}

	if config.TasaSampling == 0 {
		config.TasaSampling = 1.0
	}

	if config.IntervaloMetricas == 0 {
		config.IntervaloMetricas = 60 * time.Second
	}

	i := &Instrumentacion{config: config}

	res, err := i.crearResource()
	if err != nil {
		return nil, fmt.Errorf("error al crear Resource: %w", err)
	}

	if !config.DeshabilitarTracing {
		if err := i.crearTracer(res); err != nil {
			return nil, fmt.Errorf("error al crear Tracer: %w", err)
		}
	}

	if !config.DeshabilitarMetricas {
		if err := i.crearMetrics(res); err != nil {
			return nil, fmt.Errorf("error al crear Métricas: %w", err)
		}
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return i, nil
}

// crear el Resource de Otel
func (i *Instrumentacion) crearResource() (*resource.Resource, error) {
	gcpRes := resource.Default()
	if i.config.Ambiente != AmbienteLocal {
		if detectado, err := gcp.NewDetector().Detect(context.Background()); err == nil {
			gcpRes, _ = resource.Merge(gcpRes, detectado)
		}
	}

	serviceRes, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(i.config.Servicio),
			semconv.ServiceVersion(i.config.Version),
			attribute.String("environment", string(i.config.Ambiente)),
		),
	)
	if err != nil {
		return nil, err
	}

	return resource.Merge(gcpRes, serviceRes)
}

func (i *Instrumentacion) crearTracer(res *resource.Resource) error {
	var exporter sdktrace.SpanExporter
	var err error

	// TODO: Permitir exportador GCP en ambiente local?
	if i.config.UsarExportadorConsola {
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	} else if i.config.Ambiente != AmbienteLocal {
		opts := []option.ClientOption{}
		if i.config.ArchivoCredenciales != "" {
			opts = append(opts, option.WithCredentialsFile(i.config.ArchivoCredenciales))
		}
		exporter, err = gcpexporter.New(
			gcpexporter.WithProjectID(i.config.IdProyecto),
			gcpexporter.WithTraceClientOptions(opts),
		)
	}

	if err != nil {
		return fmt.Errorf("error al crear el trace exporter: %w", err)
	}

	if exporter == nil { // no creamos un exporter arriba
		i.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.TraceIDRatioBased(i.config.TasaSampling)),
		)
	} else {
		// Usar un exportador síncrono en ambiente local para evitar problemas con buffering en tests
		processorOpt := sdktrace.WithBatcher(exporter)
		if i.config.UsarExportadorConsola && i.config.Ambiente == AmbienteLocal {
			processorOpt = sdktrace.WithSyncer(exporter)
		}
		i.tracerProvider = sdktrace.NewTracerProvider(
			processorOpt,
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.TraceIDRatioBased(i.config.TasaSampling)),
		)
	}

	// colocar como provider global
	otel.SetTracerProvider(i.tracerProvider)

	i.tracer = i.tracerProvider.Tracer(i.config.Servicio)

	return nil
}

func (i *Instrumentacion) crearMetrics(res *resource.Resource) error {
	var exporter metricSdk.Exporter
	var err error

	// TODO: Permitir exportador GCP en ambiente local?
	if i.config.UsarExportadorConsola {
		exporter, err = stdoutmetric.New()
	} else if i.config.Ambiente != AmbienteLocal {
		opts := []option.ClientOption{}
		if i.config.ArchivoCredenciales != "" {
			opts = append(opts, option.WithCredentialsFile(i.config.ArchivoCredenciales))
		}
		exporter, err = gcpmetric.New(
			gcpmetric.WithProjectID(i.config.IdProyecto),
			gcpmetric.WithMonitoringClientOptions(opts...),
		)
	}

	if err != nil {
		return fmt.Errorf("error al crear el metric exporter: %w", err)
	}

	if exporter == nil { // arriba no creamos un exporter
		i.meterProvider = metricSdk.NewMeterProvider(
			metricSdk.WithResource(res),
		)
	} else {
		i.meterProvider = metricSdk.NewMeterProvider(
			metricSdk.WithReader(metricSdk.NewPeriodicReader(exporter,
				metricSdk.WithInterval(i.config.IntervaloMetricas))),
			metricSdk.WithResource(res),
		)
	}

	// colocar como provider global
	otel.SetMeterProvider(i.meterProvider)

	i.meter = i.meterProvider.Meter(i.config.Servicio)

	// iniciar instrumentos para métricas HTTP
	if i.config.HabilitarMetricasAutomaticas {
		i.contadorPeticiones, err = i.meter.Int64Counter(
			"http.server.request.count",
			metric.WithDescription("Número total de peticiones HTTP"),
			metric.WithUnit("1"),
		)

		if err != nil {
			return fmt.Errorf("error al crear contador de peticiones: %w", err)
		}

		i.duracionPeticion, err = i.meter.Float64Histogram(
			"http.server.request.duration",
			metric.WithDescription("Duración de peticiones HTTP"),
			metric.WithUnit("ms"),
		)

		if err != nil {
			return fmt.Errorf("error al crear duración de peticiones: %w", err)
		}

		i.peticionesActivas, err = i.meter.Int64UpDownCounter(
			"http.server.active_requests",
			metric.WithDescription("Número de peticiones HTTP activas"),
			metric.WithUnit("1"),
		)

		if err != nil {
			return fmt.Errorf("error al crear peticiones activas: %w", err)
		}

	}

	return nil
}

// ValidarConfiguracionParaHTTP valida que la configuración sea compatible
// con el middleware HTTP de OpenTelemetry.
//
// El middleware otelhttp requiere que tanto el tracer provider como el meter
// provider estén disponibles. Por lo tanto, esta función verifica que:
// - El tracing no esté deshabilitado
// - Las métricas no estén deshabilitadas
func (i *Instrumentacion) ValidarConfiguracionParaHTTP() error {
	if i.config.DeshabilitarTracing {
		return fmt.Errorf("el middleware HTTP requiere que el tracing esté habilitado, pero DeshabilitarTracing=true")
	}

	if i.config.DeshabilitarMetricas {
		return fmt.Errorf("el middleware HTTP requiere que las métricas estén habilitadas, pero DeshabilitarMetricas=true")
	}

	if i.tracerProvider == nil {
		return fmt.Errorf("el tracer provider no está inicializado")
	}

	if i.meterProvider == nil {
		return fmt.Errorf("el meter provider no está inicializado")
	}

	return nil
}

func (i *Instrumentacion) Shutdown(ctx context.Context) error {
	var errs []error

	if i.tracerProvider != nil {
		if err := i.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("no se pudo hacer shutdown al tracer provider: %w", err))
		}
	}

	if i.meterProvider != nil {
		if err := i.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("no se pudo hacer shutdown al meter provider: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errores al hacer shutdown: %v", errs)
	}

	return nil
}

// MiddlewareInstrContexto agrega la instrumentación al contexto
func (i *Instrumentacion) MiddlewareInstrContexto(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), llaveInstrumentacion, i),
		))
	})
}

// MiddlewareOtel llama al middleware que proporciona la librería otelhttp
// con unas pocas opciones añadidas.
//
// NOTA: Este middleware requiere que tanto el tracing como las métricas
// estén habilitados en la configuración. Si cualquiera de los dos está
// deshabilitado, este método entrará en pánico.
//
//	instrumentacion, err := instrumentacion.NuevaInstrumentacion(config)
//	r := chi.NewRouter()
//	r.Use(instrumentacion.MiddlewareOtel())
//
// Use ValidarConfiguracionParaHTTP antes de llamar a este método para
// asegurar una configuración válida.
func (i *Instrumentacion) MiddlewareOtel() func(http.Handler) http.Handler {
	if err := i.ValidarConfiguracionParaHTTP(); err != nil {
		panic(fmt.Sprintf("configuración inválida para MiddlewareOtel: %v", err))
	}
	formatterNombreSpan := func(op string, r *http.Request) string {
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			if ruta := rctx.RoutePattern(); ruta != "" {
				return fmt.Sprintf("%s %s", r.Method, ruta)
			}
		}
		return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	}

	opts := []otelhttp.Option{
		otelhttp.WithTracerProvider(i.tracerProvider),
		otelhttp.WithMeterProvider(i.meterProvider),
		otelhttp.WithSpanNameFormatter(formatterNombreSpan),
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	}

	if i.config.Servicio != "" {
		opts = append(opts, otelhttp.WithSpanOptions(
			trace.WithAttributes(
				semconv.ServiceName(i.config.Servicio),
				attribute.String("environment", string(i.config.Ambiente)),
			),
		))
	}

	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "http.request", opts...)
	}
}

const LlaveUuidSesion = "sesion.uuid"
const LlaveUuidPersona = "persona.uuid"

// MiddlewareAttrsAdicionales agrega datos adicionales como datos de sesión y
// el request id y ruta.
//
//	instrumentacion, err := instrumentacion.NuevaInstrumentacion(config)
//	r := chi.NewRouter()
//	r.Use(instrumentacion.MiddlewareOtel())
//	r.Use(instrumentacion.MiddlewareAttrsAdicionales)
func (i *Instrumentacion) MiddlewareAttrsAdicionales(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())

		if span.IsRecording() {
			if idReq := middleware.GetReqID(r.Context()); idReq != "" {
				span.SetAttributes(attribute.String("http.request_id", idReq))
			}

			if idSesion := r.Context().Value(LlaveUuidSesion); idSesion != nil {
				if uuidSesion, ok := idSesion.(uuid.UUID); ok {
					span.SetAttributes(attribute.String("session.id", uuidSesion.String()))
				}
			}

			if idPersona := r.Context().Value(LlaveUuidPersona); idPersona != nil {
				if uuidPersona, ok := idPersona.(uuid.UUID); ok {
					span.SetAttributes(attribute.String("user.id", uuidPersona.String()))
				}
			}

			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				if ruta := rctx.RoutePattern(); ruta != "" {
					span.SetAttributes(attribute.String("http.route", ruta))
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// MiddlewareMetricas cuenta peticiones y su duración si la configuración
// `HabilitarMetricasAutomaticas` es true.
func (i *Instrumentacion) MiddlewareMetricas(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		if i.peticionesActivas != nil {
			i.peticionesActivas.Add(r.Context(), 1)
			defer i.peticionesActivas.Add(r.Context(), -1)
		}

		next.ServeHTTP(ww, r)

		// grabar métricas después de que termina el request
		if i.config.HabilitarMetricasAutomaticas {
			duracion := time.Since(inicio)
			estatus := ww.Status()
			if estatus == 0 {
				estatus = 200
			}

			ruta := "desconocida"
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				if patron := rctx.RoutePattern(); patron != "" {
					ruta = patron
				}
			}

			attrs := metric.WithAttributes(
				semconv.HTTPRequestMethodOriginal(r.Method),
				semconv.HTTPResponseStatusCode(estatus),
				semconv.HTTPRoute(ruta),
			)

			if i.contadorPeticiones != nil {
				i.contadorPeticiones.Add(r.Context(), 1, attrs)
			}

			if i.duracionPeticion != nil {
				i.duracionPeticion.Record(r.Context(), duracion.Seconds()*1000, attrs)
			}
		}
	})
}

func (i *Instrumentacion) IniciarSpan(ctx context.Context,
	nombre string,
	opts ...trace.SpanStartOption) (context.Context, trace.Span) {

	if i.tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	return i.tracer.Start(ctx, nombre, opts...)
}

func (i *Instrumentacion) GetTracer() trace.Tracer {
	return i.tracer
}

func (i *Instrumentacion) GetMeter() metric.Meter {
	return i.meter
}

// InstrumentacionDeContexto recupera la instancia *Instrumentacion del ctx.
func InstrumentacionDeContexto(ctx context.Context) *Instrumentacion {
	if i, ok := ctx.Value(llaveInstrumentacion).(*Instrumentacion); ok {
		return i
	}
	return nil
}

// GrabarHistograma graba un simple histograma personalizado
//
// NOTA: para obtener un indicador observable o asíncrono se tiene que usar
// `i.meter.RegisterCallback`.
//
// síncronos: Counter, Histogram, UpDownCounter
// asíncronos: ObservableGauge, ObservableCounter
func (i *Instrumentacion) GrabarMetrica(ctx context.Context, nombre string,
	valor float64, attrs ...attribute.KeyValue) {

	if i.meter == nil {
		return
	}

	if histograma, err := i.meter.Float64Histogram(nombre,
		metric.WithDescription("Histograma personalizado"),
		metric.WithUnit("1"),
	); err == nil {
		histograma.Record(ctx, valor, metric.WithAttributes(attrs...))
	}
}

func GrabarMetricaNegocio(ctx context.Context,
	nombre string, valor float64, etiquetas map[string]string) {

	i := InstrumentacionDeContexto(ctx)
	if i == nil {
		return
	}

	attrs := make([]attribute.KeyValue, 0, len(etiquetas))
	for k, v := range etiquetas {
		attrs = append(attrs, attribute.String(k, v))
	}

	i.GrabarMetrica(ctx, nombre, valor, attrs...)
}

func TrazarLlamadaAPI(ctx context.Context,
	servicio string,
	metodoHttp string,
	url string) (context.Context, trace.Span) {

	i := InstrumentacionDeContexto(ctx)
	if i == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	return i.IniciarSpan(ctx, fmt.Sprintf("http.%s", servicio),
		trace.WithAttributes(
			semconv.HTTPRequestMethodOriginal(metodoHttp),
			attribute.String("http.url", url),
			attribute.String("http.service", servicio),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
}
