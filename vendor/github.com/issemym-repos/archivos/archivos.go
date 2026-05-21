package archivos

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

type Ambiente string

const (
	AmbienteLocal      Ambiente = "LOCAL"
	AmbienteDesarrollo Ambiente = "DESARROLLO"
	AmbienteProduccion Ambiente = "PRODUCCION"
)

type Config struct {
	Ambiente Ambiente

	// Configuración para ambiente local
	// DirArchivosLocal es la ruta (en el filesystem local) donde se encuentran
	// los archivos estáticos. Ej: "./static"
	// Coloca los archivos dentro dos subdirectorios public/ y private/ dentro,
	// los cuales se intentan crear si no existen.
	DirArchivosLocal string
	// RutaPubLocal es el URL base para archivos públicos Ej: "/static/public".
	// No debe incluir el dominio o protocolo.
	RutaPubLocal string
	// RutaPrivLocal es el URL base para archivos privados Ej: "/static/private".
	// No debe incluir el dominio o protocolo.
	RutaPrivLocal string

	// Nombre de proyecto GCP, Ej: "aplicativos-issemym-dev"
	IdProyectoGCP string
	// Nombre de la cubeta pública
	CubetaPublicaGCP string
	// Nombre de la cubeta privada
	CubetaPrivadaGCP string
	// Nombre del servicio en GCP, Ej: "plantilla"
	Servicio string
	// URL del CDN/Load balancer para Assets públicos, sin incluir el Servicio
	// ni /static Ej: "https://cdn.example.com"
	UrlCDN string
	// Nombre del recurso de URL Map para invalidación de caché CDN (opcional)
	NombreURLMapCDN string

	// Expiración para urls privados firmados
	ExpiracionURLsFirmados time.Duration // Default: 15 minutos

	// el header de Cache-Control para el writer de archivos en GCS
	CacheControl string // Default: "public, max-age=3600"
}

type ManejadorArchivos struct {
	config  *Config
	publico repositorioArchivos
	privado repositorioArchivos
	// El cliente GCP, nil si es ambiente local.
	cliente *storage.Client
	// TODO: Para implementar invalidación de cachés en CDN
	// computeClient *compute.GlobalOperationsClient
}

// Archivo es un wrapper sobre un path que almacena si es un público o privado.
type Archivo struct {
	manejador *ManejadorArchivos
	path      string
	esPublico bool
}

func (f *Archivo) Guardar(ctx context.Context, reader io.Reader) error {
	if f.esPublico {
		return f.manejador.publico.guardarArchivo(ctx, f.path, reader)
	}
	return f.manejador.privado.guardarArchivo(ctx, f.path, reader)
}

func (f *Archivo) GetURL(ctx context.Context) (string, error) {
	if f.esPublico {
		return f.manejador.publico.getURL(ctx, f.path)
	}
	return f.manejador.privado.getURL(ctx, f.path)
}

func (f *Archivo) Get(ctx context.Context) (io.ReadCloser, error) {
	if f.esPublico {
		return f.manejador.publico.archivo(ctx, f.path)
	}
	return f.manejador.privado.archivo(ctx, f.path)
}

func (f *Archivo) Eliminar(ctx context.Context) error {
	if f.esPublico {
		return f.manejador.publico.eliminar(ctx, f.path)
	}
	return f.manejador.privado.eliminar(ctx, f.path)
}

func (f *Archivo) Existe(ctx context.Context) (bool, error) {
	if f.esPublico {
		return f.manejador.publico.existe(ctx, f.path)
	}
	return f.manejador.privado.existe(ctx, f.path)
}

func (f *Archivo) Path() string {
	return f.path
}

func (f *Archivo) EsPublico() bool {
	return f.esPublico
}

type repositorioArchivos interface {
	guardarArchivo(context.Context, string, io.Reader) error
	getURL(context.Context, string) (string, error)
	archivo(context.Context, string) (io.ReadCloser, error)
	eliminar(context.Context, string) error
	existe(context.Context, string) (bool, error)
}

func NuevoManejador(config *Config) (*ManejadorArchivos, error) {
	// defaults y validación de config

	if config.ExpiracionURLsFirmados == 0 {
		config.ExpiracionURLsFirmados = 15 * time.Minute
	}

	if config.Ambiente == "" {
		config.Ambiente = AmbienteLocal
	}

	if config.Ambiente == AmbienteLocal {
		if config.DirArchivosLocal == "" {
			config.DirArchivosLocal = "./static"
		}
		if config.RutaPubLocal == "" {
			config.RutaPubLocal = "/static/public"
		}
		if config.RutaPrivLocal == "" {
			config.RutaPrivLocal = "/static/private"
		}
	} else {
		if config.IdProyectoGCP == "" {
			return nil, fmt.Errorf("IdProyectoGCP es requerido para ambientes no locales")
		}
		if config.CubetaPublicaGCP == "" {
			return nil, fmt.Errorf("CubetaPublicaGCP es requerido para ambientes no locales")
		}
		if config.CubetaPrivadaGCP == "" {
			return nil, fmt.Errorf("CubetaPrivadaGCP es requerido para ambientes no locales")
		}
		if config.Servicio == "" {
			return nil, fmt.Errorf("Servicio es requerido para ambientes no locales")
		}

		if config.CacheControl == "" {
			config.CacheControl = "public, max-age=3600"
		}
	}

	m := &ManejadorArchivos{config: config}

	if config.Ambiente == AmbienteLocal {
		m.publico = &repositorioLocal{
			basePath:  filepath.Join(config.DirArchivosLocal, "public"),
			urlPrefix: config.RutaPubLocal,
			config:    config,
		}

		m.privado = &repositorioLocal{
			basePath:  filepath.Join(config.DirArchivosLocal, "private"),
			urlPrefix: config.RutaPrivLocal,
			config:    config,
		}
	} else {
		cliente, err := storage.NewClient(context.Background())
		if err != nil {
			return nil, fmt.Errorf("error al crear cliente de GCP Storage: %w", err)
		}
		m.cliente = cliente

		// TODO: Agregar cliente para operaciones de CDN cuando se implemente

		m.publico = &repositorioGCP{
			cliente:     cliente,
			cubeta:      config.CubetaPublicaGCP,
			prefijoPath: config.Servicio + "/static",
			esPublico:   true,
			urlCDN:      config.UrlCDN,
			config:      config,
		}

		m.privado = &repositorioGCP{
			cliente:     cliente,
			cubeta:      config.CubetaPrivadaGCP,
			prefijoPath: config.Servicio,
			esPublico:   false,
			config:      config,
		}
	}

	return m, nil
}

func (m *ManejadorArchivos) ArchivoPublico(path string) *Archivo {
	return &Archivo{
		manejador: m,
		path:      path,
		esPublico: true,
	}
}

func (m *ManejadorArchivos) ArchivoPrivado(path string) *Archivo {
	return &Archivo{
		manejador: m,
		path:      path,
		esPublico: false,
	}
}

func (m *ManejadorArchivos) UrlPublico(ctx context.Context, pathPublico string) string {
	url, err := m.publico.getURL(ctx, pathPublico)
	if err != nil {
		return pathPublico
	}
	return url
}

func (m *ManejadorArchivos) UrlPrivado(ctx context.Context, filePath string) string {
	url, err := m.privado.getURL(ctx, filePath)
	if err != nil {
		return ""
	}
	return url
}

func (m *ManejadorArchivos) PrefijoUrlPublico() string {
	if m.config.Ambiente == AmbienteLocal {
		return m.config.RutaPubLocal
	}

	if m.config.UrlCDN != "" {
		return fmt.Sprintf("%s/%s/static", strings.TrimSuffix(m.config.UrlCDN, "/"), m.config.Servicio)
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s/static", m.config.CubetaPublicaGCP, m.config.Servicio)
}

func (m *ManejadorArchivos) Close() error {
	var errs []error

	if m.cliente != nil {
		if err := m.cliente.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error al cerrar cliente de storage: %w", err))
		}
	}

	// TODO: cerrar cliente compute cuando se implemente

	if len(errs) > 0 {
		return fmt.Errorf("errores al cerrar clientes: %v", errs)
	}
	return nil
}

type repositorioLocal struct {
	basePath  string
	urlPrefix string
	config    *Config
}

func (l *repositorioLocal) guardarArchivo(ctx context.Context, filePath string, reader io.Reader) error {
	fullPath := filepath.Join(l.basePath, filePath)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error al crear directorio: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("error al crear archivo: %w", err)
	}
	defer file.Close()

	// asegurar que el io.Reader recibido sea bufio.Reader
	reader = asegurarBuffer(reader)

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("error al escribir archivo: %w", err)
	}

	return nil
}

func (l *repositorioLocal) getURL(ctx context.Context, filePath string) (string, error) {
	return path.Join(l.urlPrefix, filePath), nil
}

func (l *repositorioLocal) archivo(ctx context.Context, filePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir archivo: %w", err)
	}
	return &bufReadCloser{reader: bufio.NewReader(file), closer: file}, nil
}

func (l *repositorioLocal) eliminar(ctx context.Context, filePath string) error {
	fullPath := filepath.Join(l.basePath, filePath)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("error al eliminar archivo: %w", err)
	}
	return nil
}

func (l *repositorioLocal) existe(ctx context.Context, filePath string) (bool, error) {
	fullPath := filepath.Join(l.basePath, filePath)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

type repositorioGCP struct {
	cliente     *storage.Client
	cubeta      string
	prefijoPath string
	esPublico   bool
	urlCDN      string
	config      *Config
}

func (g *repositorioGCP) getRutaObjeto(filePath string) string {
	if g.prefijoPath != "" {
		return path.Join(g.prefijoPath, filePath)
	}
	return filePath
}

func (g *repositorioGCP) guardarArchivo(ctx context.Context,
	rutaArchivo string, reader io.Reader) error {

	rutaObjeto := g.getRutaObjeto(rutaArchivo)
	cubeta := g.cliente.Bucket(g.cubeta)
	obj := cubeta.Object(rutaObjeto)

	writer := obj.NewWriter(ctx)

	// caché para archivos públicos
	if g.esPublico {
		writer.CacheControl = g.config.CacheControl
	}

	// asegurar que el io.Reader recibido sea bufio.Reader
	reader = asegurarBuffer(reader)
	if _, err := io.Copy(writer, reader); err != nil {
		writer.Close()
		return fmt.Errorf("error al escribir a GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("error al cerrar el writer GCS: %w", err)
	}

	return nil
}

func (g *repositorioGCP) getURL(ctx context.Context, rutaArchivo string) (string, error) {
	objectPath := g.getRutaObjeto(rutaArchivo)

	// url del cdn para archivos públicos
	if g.esPublico && g.urlCDN != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(g.urlCDN, "/"), objectPath), nil
	}

	if !g.esPublico {
		// URLs firmados (signed) para archivos privados
		opts := &storage.SignedURLOptions{
			Scheme:  storage.SigningSchemeV4,
			Method:  "GET",
			Expires: time.Now().Add(g.config.ExpiracionURLsFirmados),
		}

		urlFirmado, err := g.cliente.Bucket(g.cubeta).SignedURL(objectPath, opts)
		if err != nil {
			return "", fmt.Errorf("error al general URL firmado: %w", err)
		}
		return urlFirmado, nil
	}

	// URL público de GCS por si no funcionó
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.cubeta, objectPath), nil
}

func (g *repositorioGCP) archivo(ctx context.Context, filePath string) (io.ReadCloser, error) {
	pathObjecto := g.getRutaObjeto(filePath)
	cubeta := g.cliente.Bucket(g.cubeta)
	obj := cubeta.Object(pathObjecto)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al leer de GCS: %w", err)
	}

	return &bufReadCloser{reader: bufio.NewReader(reader), closer: reader}, nil
}

func (g *repositorioGCP) eliminar(ctx context.Context, filePath string) error {
	pathObjeto := g.getRutaObjeto(filePath)
	cubeta := g.cliente.Bucket(g.cubeta)
	obj := cubeta.Object(pathObjeto)

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("error al eliminar de GCS: %w", err)
	}

	// Invalidar caché de CDN si es un archivo público
	if g.esPublico && g.urlCDN != "" && g.config.NombreURLMapCDN != "" {
		if err := g.invalidarCacheCDN(ctx, pathObjeto); err != nil {
			// Log error but don't fail the operation - file was already deleted
			// In a real application, you might want to use the logs library here
			fmt.Printf("Warning: failed to invalidate CDN cache for %s: %v\n", pathObjeto, err)
		}
	}

	return nil
}

func (g *repositorioGCP) existe(ctx context.Context, filePath string) (bool, error) {
	objectPath := g.getRutaObjeto(filePath)
	bucket := g.cliente.Bucket(g.cubeta)
	obj := bucket.Object(objectPath)

	_, err := obj.Attrs(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// invalidarCacheCDN invalida el caché de CDN para un archivo específico
func (g *repositorioGCP) invalidarCacheCDN(ctx context.Context, objectPath string) error {
	// TODO: Implementar invalidación de caché CDN usando la API correcta
	// La implementación completa requeriría:
	// 1. Cliente para URL Maps de Compute Engine
	// 2. Configurar la invalidación usando InvalidateCache
	// 3. Manejar operaciones asíncronas
	//
	return nil
}

// Middleware es compatible con net/http. El prefijo de URL público
// se agrega al contexto para poder utilizarlo en componentes templ.
func (m *ManejadorArchivos) Middleware(next http.Handler) http.Handler {
	prefijo := m.PrefijoUrlPublico()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), llaveContexto, prefijo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type llaveCtx string

const llaveContexto llaveCtx = "prefijoUrlPublico"

// bufReadCloser se utiliza para que el método archivo que regresa un io.ReadCloser
// pueda regresar un buffered reader
type bufReadCloser struct {
	reader *bufio.Reader
	closer io.Closer
}

func (b *bufReadCloser) Read(p []byte) (int, error) { return b.reader.Read(p) }
func (b *bufReadCloser) Close() error               { return b.closer.Close() }

// PrefijoUrlPublicoDeContexto recupera el prefijo de URL público del contexto.
// Por ejemplo, dentro de componentes templ.
func PrefijoUrlPublicoDeContexto(ctx context.Context) string {
	if prefijo, ok := ctx.Value(llaveContexto).(string); ok {
		return prefijo
	}
	return ""
}

// Utilería que utiliza el manejador de archivos del contexto.

// Asset recupera el prefijo de URL público del contexto y genera un URL para un archivo
// público.
func Asset(ctx context.Context, rutaAsset string) string {
	prefijo := PrefijoUrlPublicoDeContexto(ctx)
	if prefijo == "" {
		return rutaAsset
	}
	if strings.HasPrefix(prefijo, "http") {
		// path.Join no funciona si el CDN empieza con https://
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(prefijo, "/"), rutaAsset)
	}
	return path.Join(prefijo, rutaAsset)
}

// ConfigDesdeEnv crea una configuración usando variables de entorno.
func ConfigDesdeEnv() *Config {
	config := &Config{
		Ambiente:               Ambiente(getEnvODefault("ARCHIVOS_AMBIENTE", string(AmbienteLocal))),
		DirArchivosLocal:       getEnvODefault("ARCHIVOS_DIR_LOCAL", "./static"),
		RutaPubLocal:           getEnvODefault("ARCHIVOS_RUTA_PUB_LOCAL", "/static/public"),
		RutaPrivLocal:          getEnvODefault("ARCHIVOS_RUTA_PRIV_LOCAL", "/static/private"),
		IdProyectoGCP:          os.Getenv("GOOGLE_CLOUD_PROJECT"),
		CubetaPublicaGCP:       os.Getenv("ARCHIVOS_CUBETA_PUBLICA"),
		CubetaPrivadaGCP:       os.Getenv("ARCHIVOS_CUBETA_PRIVADA"),
		Servicio:               os.Getenv("ARCHIVOS_SERVICIO"),
		UrlCDN:                 os.Getenv("ARCHIVOS_URL_CDN"),
		NombreURLMapCDN:        os.Getenv("ARCHIVOS_NOMBRE_URLMAP_CDN"),
		ExpiracionURLsFirmados: parseTimeEnv("ARCHIVOS_EXPIRACION_URLs", 15*time.Minute),
	}

	return config
}

func getEnvODefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseTimeEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// asegurarBuffer envuelve un io.Reader en un bufio.Reader, esta operación es
// segura de repetirse (se puede envolver un bufio.Reader).
func asegurarBuffer(r io.Reader) io.Reader {
	// 32 KiB es el tamaño del buffer que usa internamente io.Copy; se iguala
	// por si las dudas.
	return bufio.NewReaderSize(r, 32*1024)
}
