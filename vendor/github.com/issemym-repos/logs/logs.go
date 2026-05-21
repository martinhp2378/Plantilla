package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

type Severidad string

// Niveles de severidad que coinciden con los de GCP
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
const (
	SeveridadDefault   Severidad = "DEFAULT"
	SeveridadDebug     Severidad = "DEBUG"
	SeveridadInfo      Severidad = "INFO"
	SeveridadNotice    Severidad = "NOTICE"
	SeveridadWarning   Severidad = "WARNING"
	SeveridadError     Severidad = "ERROR"
	SeveridadCritical  Severidad = "CRITICAL"
	SeveridadAlert     Severidad = "ALERT"
	SeveridadEmergency Severidad = "EMERGENCY"
)

// Colores para logs, copiados de chi/middleware/terminal.go
var (
	// Colores normales
	nBlack   = []byte{'\033', '[', '3', '0', 'm'}
	nRed     = []byte{'\033', '[', '3', '1', 'm'}
	nGreen   = []byte{'\033', '[', '3', '2', 'm'}
	nYellow  = []byte{'\033', '[', '3', '3', 'm'}
	nBlue    = []byte{'\033', '[', '3', '4', 'm'}
	nMagenta = []byte{'\033', '[', '3', '5', 'm'}
	nCyan    = []byte{'\033', '[', '3', '6', 'm'}
	nWhite   = []byte{'\033', '[', '3', '7', 'm'}
	// Colores brillantes
	bBlack   = []byte{'\033', '[', '3', '0', ';', '1', 'm'}
	bRed     = []byte{'\033', '[', '3', '1', ';', '1', 'm'}
	bGreen   = []byte{'\033', '[', '3', '2', ';', '1', 'm'}
	bYellow  = []byte{'\033', '[', '3', '3', ';', '1', 'm'}
	bBlue    = []byte{'\033', '[', '3', '4', ';', '1', 'm'}
	bMagenta = []byte{'\033', '[', '3', '5', ';', '1', 'm'}
	bCyan    = []byte{'\033', '[', '3', '6', ';', '1', 'm'}
	bWhite   = []byte{'\033', '[', '3', '7', ';', '1', 'm'}

	reset = []byte{'\033', '[', '0', 'm'}
)

var isTTY bool

// Detectar si stdout es un TTY (copiado de chi)
func init() {
	fi, err := os.Stdout.Stat()
	if err == nil {
		m := os.ModeDevice | os.ModeCharDevice
		isTTY = fi.Mode()&m == m
	}
}

// colorWrite - escritura con colores basada en chi
func colorWrite(w io.Writer, usarColor bool, color []byte, s string, args ...interface{}) {
	if isTTY && usarColor {
		w.Write(color)
	}
	fmt.Fprintf(w, s, args...)
	if isTTY && usarColor {
		w.Write(reset)
	}
}

// printPrettyStack - impresión de stack trace con control de colores
func printPrettyStack(rvr interface{}, usarColores bool) {
	debugStack := debug.Stack()
	out, err := parsePrettyStack(debugStack, rvr, usarColores)
	if err == nil {
		os.Stderr.Write(out)
	} else {
		// Fallback a salida estándar
		os.Stderr.Write(debugStack)
	}
}

func parsePrettyStack(debugStack []byte, rvr interface{}, usarColores bool) ([]byte, error) {
	usarColor := isTTY && usarColores // hay que validar encima si es TTY
	buf := bytes.NewBuffer(nil)

	colorWrite(buf, false, bRed, "\n")
	colorWrite(buf, usarColor, bCyan, " panic: ")
	colorWrite(buf, usarColor, bBlue, "%v", rvr)
	colorWrite(buf, false, bWhite, "\n \n")

	// Procesar información del stack
	stack := strings.Split(string(debugStack), "\n")
	lines := []string{}

	// Localizar línea de panic
	for i := len(stack) - 1; i > 0; i-- {
		lines = append(lines, stack[i])
		if strings.HasPrefix(stack[i], "panic(") {
			lines = lines[0 : len(lines)-2] // remover boilerplate
			break
		}
	}

	// El stack se muestra en orden inverso
	for i := len(lines)/2 - 1; i >= 0; i-- {
		opp := len(lines) - 1 - i
		lines[i], lines[opp] = lines[opp], lines[i]
	}

	// Decorar stack trace con colores
	for i, line := range lines {
		if i%2 == 0 {
			// función
			if strings.Contains(line, ".") {
				parts := strings.Split(line, ".")
				if len(parts) > 1 {
					colorWrite(buf, usarColor, nYellow, " %s.", parts[0])
					colorWrite(buf, usarColor, bGreen, "%s", strings.Join(parts[1:], "."))
				}
			} else {
				colorWrite(buf, usarColor, bGreen, " %s", line)
			}
		} else {
			// archivo:número
			if strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					colorWrite(buf, usarColor, nBlue, "     %s:", parts[0])
					colorWrite(buf, usarColor, bGreen, "%s", strings.Join(parts[1:], ":"))
				}
			} else {
				colorWrite(buf, usarColor, nBlue, "     %s", line)
			}
		}
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

type Logger struct {
	*slog.Logger
	config *Config
}

func NuevoLogger(config *Config) *Logger {
	// defaults
	if config.Ambiente == "" {
		config.Ambiente = AmbienteLocal
	}
	if config.Servicio == "" {
		config.Servicio = "desconocido"
	}
	if config.Nivel == "" {
		if config.Ambiente == AmbienteLocal {
			config.Nivel = SeveridadDebug
		} else {
			config.Nivel = SeveridadInfo
		}
	}

	if config.Ambiente == AmbienteLocal {
		config.AgregarSource = true
	}

	var handler slog.Handler
	var nivel slog.Level

	// convertir a niveles de slog
	switch config.Nivel {
	case SeveridadDebug:
		nivel = slog.LevelDebug
	case SeveridadEmergency, SeveridadError, SeveridadCritical, SeveridadAlert:
		nivel = slog.LevelError
	case SeveridadNotice, SeveridadWarning:
		nivel = slog.LevelWarn
	default: // Info y Default
		nivel = slog.LevelInfo
	}

	attrs := []slog.Attr{
		slog.String("proyecto", config.IdProyecto),
		slog.String("servicio", config.Servicio),
		slog.String("version", config.Version),
		slog.String("ambiente", string(config.Ambiente)),
	}

	if config.Ambiente == AmbienteLocal {
		opts := &slog.HandlerOptions{
			Level:     nivel,
			AddSource: config.AgregarSource,
			ReplaceAttr: func(grupos []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					if t, ok := a.Value.Any().(time.Time); ok {
						a.Value = slog.StringValue(t.Format("15:04:05.000"))
					}
				}
				return a
			},
		}
		handler = slog.NewTextHandler(os.Stdout, opts).WithAttrs(attrs)
	} else {
		handler = &gcpHandler{
			// NO escribir a stdout directamente - el gcpHandler lo hará
			handler: slog.NewJSONHandler(io.Discard,
				&slog.HandlerOptions{
					AddSource: config.AgregarSource,
					Level:     nivel,
				}).WithAttrs(attrs),

			config: config,
			attrs:  attrs,
		}
	}

	return &Logger{
		Logger: slog.New(handler),
		config: config,
	}
}

type gcpHandler struct {
	handler slog.Handler
	config  *Config
	attrs   []slog.Attr
	groups  []string // grupos anidados de WithGroup
}

// WithAttrs implementa slog.Handler.
func (h *gcpHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Combinar attrs existentes con los nuevos
	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)

	return &gcpHandler{
		handler: h.handler.WithAttrs(attrs),
		config:  h.config,
		attrs:   newAttrs,
		groups:  h.groups,
	}
}

// WithGroup implementa slog.Handler.
func (h *gcpHandler) WithGroup(name string) slog.Handler {
	// Ver grupos anidados
	newGroups := make([]string, 0, len(h.groups)+1)
	newGroups = append(newGroups, h.groups...)
	newGroups = append(newGroups, name)

	return &gcpHandler{
		handler: h.handler.WithGroup(name),
		config:  h.config,
		attrs:   h.attrs,
		groups:  newGroups,
	}
}

// Enabled implementa slog.Handler.
func (h *gcpHandler) Enabled(ctx context.Context, nivel slog.Level) bool {
	return h.handler.Enabled(ctx, nivel)
}

// resolverValor convierte un slog.Value a su representación adecuada para JSON.
func resolverValor(v slog.Value) any {
	switch v.Kind() {
	case slog.KindGroup:
		// Para grupos, crear un mapa anidado
		attrs := v.Group()
		m := make(map[string]any, len(attrs))
		for _, attr := range attrs {
			m[attr.Key] = resolverValor(attr.Value)
		}
		return m
	case slog.KindAny:
		return v.Any()
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindBool:
		return v.Bool()
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time()
	default:
		return v.Any()
	}
}

// agregarAtributosAMapa agrega atributos a un mapa, manejando anidación de grupos
func agregarAtributosAMapa(attrs []slog.Attr, groups []string, destino map[string]any) {
	for _, attr := range attrs {
		valor := resolverValor(attr.Value)

		// Si hay grupos activos, anidar el atributo
		if len(groups) > 0 {
			// Navegar/crear la estructura de grupos
			actual := destino
			for _, nombreGrupo := range groups {
				if _, existe := actual[nombreGrupo]; !existe {
					actual[nombreGrupo] = make(map[string]any)
				}
				actual = actual[nombreGrupo].(map[string]any)
			}
			actual[attr.Key] = valor
		} else {
			destino[attr.Key] = valor
		}
	}
}

// Handle implementa slog.Handler.
func (h *gcpHandler) Handle(ctx context.Context, r slog.Record) error {
	severidad := h.nivelASeveridad(r.Level)

	// entrada estructurada de acuerdo a GCP (con json)
	// https://cloud.google.com/logging/docs/structured-logging
	entrada := map[string]any{
		"severity": severidad,
		"message":  r.Message,
		"timestamp": map[string]any{
			"seconds": r.Time.Unix(),
			"nanos":   r.Time.Nanosecond(),
		},
		"logging.googleapis.com/labels": map[string]string{
			"proyecto":    h.config.IdProyecto,
			"servicio":    h.config.Servicio,
			"version":     h.config.Version,
			"environment": string(h.config.Ambiente),
			"ambiente":    string(h.config.Ambiente),
		},
	}

	// instrumentación
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()
		idTraza := spanCtx.TraceID().String()
		idSpan := spanCtx.SpanID().String()

		entrada["logging.googleapis.com/trace"] =
			fmt.Sprintf("projects/%s/traces/%s", h.config.IdProyecto, idTraza)
		entrada["logging.googleapis.com/spanId"] = idSpan
		entrada["logging.googleapis.com/trace_sampled"] = spanCtx.IsSampled()
		entrada["trace_flags"] = spanCtx.TraceFlags().String()
	}

	// agregar ruta de archivo fuente si está disponible
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			entrada["logging.googleapis.com/sourceLocation"] = map[string]any{
				"file":     f.File,
				"line":     f.Line,
				"function": f.Function,
			}
		}
	}

	// Primero agregar atributos de WithAttrs (sin grupos)
	agregarAtributosAMapa(h.attrs, nil, entrada)

	// Luego procesar atributos del log actual
	var stackTrace string
	var errorValue any

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "stack" {
			stackTrace = a.Value.String()
		} else if a.Key == "error" {
			errorValue = a.Value.Any()
		} else if a.Key == "httpRequest" {
			entrada["httpRequest"] = a.Value.Any() // es un mapa anidado
		} else {
			// Aplicar grupos activos si los hay
			agregarAtributosAMapa([]slog.Attr{a}, h.groups, entrada)
		}
		return true
	})

	// incluir en el campo "message", el mensaje original, el error y el stack trace
	if stackTrace != "" {
		var messageBuilder strings.Builder
		messageBuilder.WriteString(r.Message)
		if errorValue != nil {
			messageBuilder.WriteString(": ")
			messageBuilder.WriteString(fmt.Sprintf("%v", errorValue))
		}
		messageBuilder.WriteString("\n")
		messageBuilder.WriteString(stackTrace)
		entrada["message"] = messageBuilder.String()
	}

	// Escribir la entrada final a stdout (una sola serialización)
	enc := json.NewEncoder(os.Stdout)
	return enc.Encode(entrada)
}

func (h *gcpHandler) nivelASeveridad(nivel slog.Level) any {
	switch {
	case nivel < slog.LevelInfo:
		return SeveridadDebug
	case nivel < slog.LevelWarn:
		return SeveridadInfo
	case nivel < slog.LevelError:
		return SeveridadWarning
	default:
		return SeveridadCritical
	}
}

// With agrega atributos extra a un logger
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
		config: l.config,
	}
}

// WithContext extrae valores del span del contexto y se las agrega al logger.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()

		idTraza := spanCtx.TraceID().String()
		idSpan := spanCtx.SpanID().String()

		logger = logger.With(
			"logging.googleapis.com/trace",
			fmt.Sprintf("projects/%s/traces/%s", l.config.IdProyecto, idTraza),

			"logging.googleapis.com/spanId", idSpan,
			"logging.googleapis.com/trace_sampled", spanCtx.IsSampled(),
		)
	}

	return logger
}

// Error loguea un error incluyendo su stack trace
func (l *Logger) Error(msg string, err error, args ...any) {
	// Si es ambiente local, imprimir stack trace
	if l.config.Ambiente == AmbienteLocal {
		printPrettyStack(err, l.config.LogsConColores)
		// no incluir el stack ya que se está imprimiendo arriba
		l.Logger.Error(msg, args...)
		return
	}
	// para ambientes GCP incluir el error y el stack como atributos
	stack := string(debug.Stack())
	argsConStack := append([]any{
		"error", err,
		"stack", stack,
	}, args...)
	l.Logger.Error(msg, argsConStack...)
}

// ErrorContext loguea un error con contexto y su stack trace
func (l *Logger) ErrorContext(ctx context.Context, msg string, err error, args ...any) {
	l.WithContext(ctx).Error(msg, err, args...)
}

// Fatal es como log.Fatal en donde se loguea el error y se sale
func (l *Logger) Fatal(msg string, err error, args ...any) {
	l.Error(msg, err, args...)
	os.Exit(1)
}

// Panic pretende reemplazar panic() incluyendo un logueo antes de hacer el
// panic
func (l *Logger) Panic(msg string, err error, args ...any) {
	l.Error(msg, err, args...)
	panic(err)
}

// LoggerDeContexto recupera el logger del contexto
func LoggerDeContexto(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(llaveLogger).(*Logger); ok {
		return logger.WithContext(ctx)
	}

	// Regresar un logger default si falla lo anterior
	return NuevoLogger(&Config{})
}

// WithLogger agrega el logger al contexto
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, llaveLogger, logger)
}

// HTTPRequest estructura una petición http
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
type HTTPRequest struct {
	RequestMethod                  string `json:"requestMethod"`
	RequestURL                     string `json:"requestUrl"`
	RequestSize                    int64  `json:"requestSize,omitempty"`
	Status                         int    `json:"status"`
	ResponseSize                   int64  `json:"responseSize,omitempty"`
	UserAgent                      string `json:"userAgent,omitempty"`
	RemoteIP                       string `json:"remoteIp,omitempty"`
	ServerIP                       string `json:"serverIp,omitempty"`
	Referer                        string `json:"referer,omitempty"`
	Latency                        string `json:"latency,omitempty"`
	CacheLookup                    bool   `json:"cacheLookup,omitempty"`
	CacheHit                       bool   `json:"cacheHit,omitempty"`
	CacheValidatedWithOriginServer bool   `json:"cacheValidatedWithOriginServer,omitempty"`
	CacheFillBytes                 int64  `json:"cacheFillBytes,omitempty"`
	Protocol                       string `json:"protocol,omitempty"`
}

// LogHTTPRequest loguea una petición en el formato estructurado de gcp
func (l *Logger) LogHTTPRequest(ctx context.Context, req HTTPRequest, msg string, args ...any) {
	argsConRequest := append([]any{"httpRequest", req}, args...)
	// poner una severidad basándose en el código http
	switch {
	case req.Status >= 500:
		l.WithContext(ctx).Error(msg, fmt.Errorf("HTTP error: %s", msg), argsConRequest...)
		return
	case req.Status >= 400:
		l.WithContext(ctx).Warn(msg, argsConRequest...)
		return
	default:
		l.WithContext(ctx).Info(msg, argsConRequest...)
		return
	}
}

func (l *Logger) MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = WithLogger(ctx, l)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// MiddlewareHTTPLogger loguea automáticamente todas las peticiones HTTP
// usando el formato estructurado de GCP.
// Debe ser usado después de MiddlewareLogger.
func (l *Logger) MiddlewareHTTPLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()

		requestSize := r.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		duracion := time.Since(inicio)

		status := ww.Status()
		if status == 0 {
			status = 200 // Default status code
		}
		tamanioRespuesta := int64(ww.BytesWritten())

		// Extraer IP del cliente (requiere del middleware.RealIP de chi)
		remoteIP := r.RemoteAddr

		// Estructura para GCP
		httpReq := HTTPRequest{
			RequestMethod: r.Method,
			RequestURL:    r.URL.String(),
			RequestSize:   requestSize,
			Status:        status,
			ResponseSize:  tamanioRespuesta,
			UserAgent:     r.UserAgent(),
			RemoteIP:      remoteIP,
			Referer:       r.Referer(),
			Latency:       duracion.String(),
			Protocol:      r.Proto,
		}

		// Crear mensaje descriptivo
		msg := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

		// Agregar información adicional como argumentos
		args := []any{
			"method", r.Method,
			"url", r.URL.String(),
			"status", status,
			"duration_ms", duracion.Milliseconds(),
			"remote_ip", remoteIP,
		}

		// Agregar request ID si está disponible (requiere del middleware.ReqID de chi)
		if reqID := middleware.GetReqID(r.Context()); reqID != "" {
			args = append(args, "request_id", reqID)
		}

		// Usar formato apropiado según el ambiente
		if l.config.Ambiente == AmbienteLocal {
			// Formato con colores y pretty print
			l.logHTTPRequestLocal(r, status, duracion, tamanioRespuesta, remoteIP)
		} else {
			// Formato estructurado en JSON para GCP
			l.LogHTTPRequest(r.Context(), httpReq, msg, args...)
		}
	})
}

// logHTTPRequestLocal loguea peticiones HTTP con color y pretty print
// mucho de esto se copió de chi
func (l *Logger) logHTTPRequestLocal(r *http.Request, status int, elapsed time.Duration, bytesWritten int64, remoteIP string) {
	usarColor := l.config.LogsConColores && isTTY
	buf := bytes.NewBuffer(nil)

	// Obtener información del caller para el source
	_, file, line, _ := runtime.Caller(2) // Caller(2) porque estamos dentro de un middleware
	source := fmt.Sprintf("%s:%d", file, line)

	// Colocar un nivel de acuerdo al código
	nivel := "INFO"
	if status >= 500 {
		nivel = "ERROR"
	} else if status >= 400 {
		nivel = "WARN"
	}

	// Agregar timestamp, nivel y source como lo hace slog
	now := time.Now()
	buf.WriteString(fmt.Sprintf("time=%s level=%s source=%s ",
		now.Format("15:04:05.000"), nivel, source))
	buf.WriteString("msg=\"")

	// Request ID con colores si está disponible (requiere del middleware.ReqID de chi)
	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		colorWrite(buf, usarColor, nYellow, "[%s] ", reqID)
	}

	// Método HTTP
	colorWrite(buf, usarColor, bMagenta, "%s ", r.Method)

	// URL/Path
	colorWrite(buf, usarColor, nCyan, "%s ", r.URL.Path)

	// Status code con color según rango (igual a chi)
	switch {
	case status < 200:
		colorWrite(buf, usarColor, bBlue, "%03d", status)
	case status < 300:
		colorWrite(buf, usarColor, bGreen, "%03d", status)
	case status < 400:
		colorWrite(buf, usarColor, bCyan, "%03d", status)
	case status < 500:
		colorWrite(buf, usarColor, bYellow, "%03d", status)
	default:
		colorWrite(buf, usarColor, bRed, "%03d", status)
	}

	// Tamaño de respuesta
	colorWrite(buf, usarColor, bBlue, " %dB", bytesWritten)

	// Tiempo de respuesta con color según duración (similar a chi)
	buf.WriteString(" in ")
	if elapsed < 500*time.Millisecond {
		colorWrite(buf, usarColor, nGreen, "%s", elapsed.Round(time.Millisecond))
	} else if elapsed < 5*time.Second {
		colorWrite(buf, usarColor, nYellow, "%s", elapsed.Round(time.Millisecond))
	} else {
		colorWrite(buf, usarColor, nRed, "%s", elapsed.Round(time.Millisecond))
	}

	// IP remota
	buf.WriteString(" from ")
	colorWrite(buf, usarColor, nWhite, "%s", remoteIP)

	// Cerrar el mensaje y agregar atributos adicionales como slog
	buf.WriteString(fmt.Sprintf("\" proyecto=\"%s\" servicio=%s version=%s ambiente=%s\n",
		l.config.IdProyecto, l.config.Servicio, l.config.Version, l.config.Ambiente))

	// Escribir directamente a stdout (y no a través de slog) para que
	// funcionen los colores
	os.Stdout.Write(buf.Bytes())
}
