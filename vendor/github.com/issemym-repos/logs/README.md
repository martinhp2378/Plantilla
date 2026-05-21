# Logs - Librería para logging e instrumentación de apps Issemym

Esta librería soporta los ambientes de Issemym: Local, Desarrollo y Producción. De los cuales Desa y Prod se pretenden
ejecutar en la nube Google GCP. En este caso aprovecha los servicios de Google Cloud Logging, Cloud Trace y Cloud
Monitoring.

## Características de Logging

- Logging con pretty print en ambiente local (similar al logger de chi).
- Logging estructurado con JSON para ambientes de GCP.
- Los stack traces se incluyen en caso de errores/panics no manejados.
- Incluye severidad y niveles de acuerdo a la librería interna log/logger.

## Instrumentación

La instrumentación se logra a partir del estándar de OpenTelemetry (OTel) y se divide en dos:

- Trazas (tracing) y spans: Se crea una traza por petición HTTP, los spans se pueden crear manual o automáticamente
por cada unidad de ejecución de la petición.
- Métricas (metrics) nos permiten obtener medidas como tiempo de ejecución, latencia, número de peticiones pendientes,
etc. y crear histogramas.
- Se incluyen middlewares compatibles con net/http (y con chi) para automatizar la instrumentación.

## Características del tracing

- En ambiente local puede mostrar la traza en consola. En ambiente GCP se utiliza un exportador para Cloud Trace y
Cloud Monitoring.
- La petición HTTP crea un span de forma automática.
- Todas las operaciones de BD crean automáticamente spans.
- Se pueden crear spans manualmente tanto para la lógica de negocios o para para trazar llamadas a servicios externos.

## Características de las métricas

- Automáticamente se generan estadísticas de las peticiones HTTP y de los queries de la BD.
- Se proporcionan métodos para generar métricas personalizadas.
- También se puede utlizar la librería de OTel para generar métricas más detalladas.

## Para instalar

```bash
go get -u github.com/issemym-repos/logs
```

## Uso sencillo para desarrollo local

```go
package main

import (
    "context"
	"fmt"
    "net/http"
	"time"
   
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/issemym-repos/logs"
)

func main() {
    config := &logs.Config{
        Servicio:                     "mi-aplicacion",
        Version:                      "1.0.0",
        Ambiente:                     logs.AmbienteLocal,
        LogsConColores:               true,  // Activar colores en logs (solo local)
        Nivel:                        logs.DEBUG,
        HabilitarMetricasAutomaticas: true,
		UsarExportadorConsola:        true,  // Para ver las trazas en consola 
    }

    logger := logs.NuevoLogger(config)

    // Iniciar instrumentación para HTTP
    instrumentacion, err := logs.NuevaInstrumentacionParaHTTP(config)
    if err != nil {
        logger.Fatal("Error inicializando instrumentación", err)
    }
    defer instrumentacion.Shutdown(context.Background())

    // Iniciar un router HTTP con nuestro stack de instrumentacion
    r := chi.NewRouter()
    // estos dos middleware de chi son usados por el logger
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
	// Opcional pero recomendado
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Recoverer)

    r.Use(logger.MiddlewareLogger)  // agregar el logger al context
    // loguear todas las peticiones HTTP
    r.Use(logger.MiddlewareHTTPLogger)
    r.Use(instrumentacion.MiddlewareInstrContexto)  // agregar la instrumentacion al contexto
    // Middleware para trazar peticiones HTTP
    r.Use(instrumentacion.MiddlewareOtel())  // hay que notar la llamada a función
    // unos cuantos atributos extras en las trazas (datos sesión, request id, ruta)
    r.Use(instrumentacion.MiddlewareAttrsAdicionales)
    r.Use(instrumentacion.MiddlewareMetricas)

    // Rutas ejemplo
    r.Get("/salud", saludHandler)
    r.Get("/usuarios/{id}", getUsuarioHandler)
   
    logger.Info("Iniciando servidor", "puerto", 8080)
    logger.Fatal("error en servidor", http.ListenAndServe(":8080", r))
}

func saludHandler(w http.ResponseWriter, r *http.Request) {
	logger := logs.LoggerDeContexto(r.Context())
    logger.Info("Se solicitó un check de salud")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func getUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	logger := logs.LoggerDeContexto(r.Context())
    idUsuario := chi.URLParam(r, "id")
   
    logger.Info("Obteniendo usuario",
        "id_usuario", idUsuario,
        "method", r.Method,
    )
   
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"id": "%s", "name": "Usuario %s"}`, idUsuario, idUsuario)
}
```

## Configuración

### Para ambiente Local

Utilizar exportadores de consola (opcional porque arrojan mucho texto) y prettyprint para los logs.

También colocar el nivel en DEBUG el cual es el más bajo.

```go
config := &logs.Config{
    Servicio:              "mi-servicio",
    Version:              "1.0.0",
    Ambiente:             logs.AmbienteLocal,
    LogsConColores:       true,  // Activar colores en logs (solo local)
    Nivel:                logs.DEBUG,
    UsarExportadorConsola: true,  // o false para evitar ver trazas en consola
}
```

### En Desarrollo (GCP)

```go
config := &logs.Config{
    Servicio:   "mi-servicio",
    Version:    "1.0.0",
    Ambiente:   logs.AmbienteDesarrollo,
    IdProyecto: "aplicativos-issemym-dev",
    Nivel:      logs.INFO,  // Nivel más alto
   
    // Opcional: Usar llaves de GCP si no se corre en Cloud Run
    ArchivoCredenciales: "/path/al/service-account.json",
}
```

### En Producción (GCP)

```go
config := &logs.Config{
    Servicio:   "mi-servicio",
    Version:    "1.0.0",
    Ambiente:   logs.AmbienteProduccion,
    IdProyecto: "aplicativos-issemym-prod",
    Nivel:      logs.WARNING,  // Nivel más alto
   
    // ArchivoCredenciales no se necesita para Cloud Run
   
    HabilitarMetricasAutomaticas: true,
    TasaSampling:                0.1, // Solo hacer sampling al 10% de las trazas
}
```

## Ejemplos de slog/logging

### Logging estructurado básico

```go
logger.Info("Usuario firmado exitosamente",
    "id_usuario", "12345",
    "email", "usuario@example.com",
    "login_method", "password",
)

logger.Warning("Intento de acceso denegado",
    "id_usuario", "12345",
    "recurso", "/admin/usuarios",
    "razon", "permisos_insuficientes",
)

// Se definen métodos propios para loguear errores, estos incluyen el stack
// trace en el log
logger.Error("Error conectando a base de datos", err,
    "database", "bd_usuarios",
)

// usar en lugar de panic()
logger.Panic("mensaje", errors.New("error fatal"))

// usar en lugar de log.fatal()
logger.Fatal("mensaje", errors.New("error fatal"))
```

### Logging de peticiones HTTP

```go
func (h *Handler) ServirHTTP(w http.ResponseWriter, r *http.Request) {
    inicio := time.Now()
   
    logger.Info("Petición iniciada",
        logger.Group("http",
            "method", r.Method,
            "path", r.URL.Path,
            "user_agent", r.UserAgent(),
            "remote_addr", r.RemoteAddr,
        ),
    )
   
    // Lógica de negocios...
   
    logger.Info("Request completado",
        logger.Group("http",
            "method", r.Method,
            "path", r.URL.Path,
            "status", 200,
            "duration_ms", time.Since(inicio).Milliseconds(),
        ),
        logger.Group("response",
            "content_type", "application/json",
            "size_bytes", responseSize,
        ),
    )
}
```

## Instrumentación de HTTP

### Orden de middlewares

Colocar los middlewares de esta manera, el orden es importante.

```go
func setupRouter(constructorSesion *sesiones.ConstructorSesion,
    logger *logs.Logger,
    instrumentacion *logs.Instrumentacion) chi.Router {

    r := chi.NewRouter()
   
    // Middleware de Chi (utilizados por el logger)
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    // Opcional pero recomendado
    r.Use(middleware.Timeout(60 * time.Second))
    r.Use(middleware.Recoverer)

    // Middleware de logging
    r.Use(logger.MiddlewareLogger)  // agregar el logger al context
    // loguear todas las peticiones HTTP
    r.Use(logger.MiddlewareHTTPLogger)
    
    // Middleware de sesiones
    r.Use(constructorSesion.Middleware)
   
    // Middleware de Instrumentación
    r.Use(instrumentacion.MiddlewareInstrContexto)  // agregar la instrumentacion al contexto
    // Middleware para trazar peticiones HTTP
    r.Use(instrumentacion.MiddlewareOtel())  // hay que notar la llamada a función
    // unos cuantos atributos extras en las trazas (datos sesión, request id, ruta)
    r.Use(instrumentacion.MiddlewareAttrsAdicionales)
   
    return r
}
```

## Creando spans de forma manual

```go
import (
	// ...
    "go.opentelemetry.io/otel/trace"
)

func hacerAlgoConRecurso(ctx context.Context, idRecurso string) (*Usuario, error) {
    // Obtener la instrumentación del contexto
    instrumentacion := logs.InstrumentacionDeContexto(ctx)
    if instrumentacion == nil {
        return nil, errors.New("no se encontró la instrumentación en el contexto")
    }
   
    // Nuestro span personalizado
    ctx, span := instrumentacion.IniciarSpan(ctx, "hacer_algo_con_recurso",
        trace.WithAttributes(
            attribute.String("recurso.id", idRecurso),
            attribute.String("recurso.operacion", "read"),
        ),
    )
    defer span.End()
   
    // Lógica de negocio...
	procesoExitoso := true
   
    span.SetAttributes(attribute.Bool("proceso.terminado", procesoExitoso))
    return usuario, nil
}
```

### Llamadas a servicios externos

```go

import (
	// ...
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func llamarAPIExterno(ctx context.Context, endpoint string) (*APIResponse, error) {
    ctx, span := logs.TrazarLlamadaAPI(ctx,
        "servicio-externo",
        http.MethodGet,
        endpoint)
    defer span.End()
   
    // Cliente http con traza
    client := &http.Client{
        Transport: otelhttp.NewTransport(http.DefaultTransport),
    }
   
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
   
    resp, err := client.Do(req)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    defer resp.Body.Close()
   
    // Agregar atributos a span
    span.SetAttributes(
        attribute.Int("http.status_code", resp.StatusCode),
        attribute.String("http.response.content_type", resp.Header.Get("Content-Type")),
    )
   
    if resp.StatusCode >= 400 {
        span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
    }

    return parsearRespuestaAPI(resp.Body)
}
```

## Métricas de forma manual

### Histogramas Personalizados

```go
import (
	// ...
    "go.opentelemetry.io/otel/attribute"	
)

func grabarMetricaPersonalizada(ctx context.Context, instrumentacion *logs.Instrumentacion) {
    instrumentacion.GrabarMetrica(ctx, "tiempo.procesamiento.personalizado",
        123.45, // valor en milisegundos
        attribute.String("componente", "servicio-usuario"),
        attribute.String("operacion", "procesamiento-datos"),
        attribute.Int("batch_size", 100),
    )
}
```

`GrabarMetricaNegocio` es similar a `GrabarMetrica` pero recibe un mapa de atributos.

```go
func grabarMetricasNegocio(ctx context.Context, operacion string, valor float64) {
    logs.GrabarMetricaNegocio(ctx, "negocio.operacion.duracion", valor, map[string]string{
        "operacion": operacion,
        "service":   "mi-servicio",
        "version":   "1.0.0",
    })
}

// Usar en nuestra lógica
func procesar(ctx context.Context) error {
    inicio := time.Now()
	
	nombreOperacion := "procesar_datos"
    
    defer func() {
        duracion := time.Since(inicio).Seconds() * 1000 // milliseconds
        grabarMetricasNegocio(ctx, nombreOperacion, duracion)
    }()

	// Lógica de negocio...
	
    return nil
}
```

## Instrumentación de bases de datos

Se agrega el tracer a la configuración de la BD y luego se llama `logs.HabilitarEstadisticas` en el pool.

```go
func ConnectPostgres(ctx context.Context, env entorno.Env) (*pgxpool.Pool, error) {
	urlWrite := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		env.PqWriteUser, env.PqWritePass, env.PqWriteHost, env.PqWritePort, env.PqWriteDb, env.PqWriteSSLMode)
	poolConfig, err := pgxpool.ParseConfig(urlWrite)
	if err != nil {
		panic(err)
	}
	poolConfig.MaxConns = env.PqWriteMaxConns
	poolConfig.MinConns = env.PqWriteMinConns
	poolConfig.MaxConnIdleTime = env.PqWriteMaxConnIdle
	poolConfig.MaxConnLifetime = env.PqWriteMaxConnLife
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	poolConfig.ConnConfig.RuntimeParams["timezone"] = "America/Mexico_City"
	poolConfig.ConnConfig.RuntimeParams["lock_timeout"] = "10s"

	// Agregar extension uuid
	poolConfig.AfterConnect = func(ctx context.Context, pgxConn *pgx.Conn) error {
		pgxuuid.Register(pgxConn.TypeMap())
		return nil
	}
	
    logsConfig := &logs.Config{
        Servicio:           "mi-servicio",
        NombreSpanCorto:    true,  // Usar nombres de span corto (SELECT vs SQL completo)
        QuitarPrefijoQuery: false, // Por default se agrega como prefijo "query" o "batch query"
        QuitarArgsQuery:    false, // Incluir los valores de los parámetros en los spans
        QuitarStatement:    false, // Incluir el statement en los atributos
    }
	
	// combinar las configuraciones
	logs.AgregarTracingPgxpool(poolConfig, logsConfig, instrumentacion)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("error al iniciar pool de postgres", err)
		return nil, err
	}
	
	// habilitar métricas
	if err = logs.HabilitarEstadisticas(pool, tracingConfig); err != nil {
        logger.Error("No se pudieron habilitar métricas en el pool", err)
		pool.Close()
		return nil, err
    }

	return pool, nil
}
```

Al usar este pool todos los queries crean un span y generan métricas.

## Despliegue en la Nube Google

Se utiliza un Google Service Account, de otra manera se puede utilizar un archivo de credenciales.
