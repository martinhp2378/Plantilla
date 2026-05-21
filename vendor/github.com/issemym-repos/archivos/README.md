# Archivos - Manejador de archivos estáticos para apps ISSEMYM

## Introducción

Esta librería permite la administración de archivos estáticos. Permite 3
ambientes de operación: Local, Desarrollo y Producción; estos dos últimos se
asume son hospedados en Cloud Run en GCP y utilizan buckets de Cloud Store y el
Cloud CDN. El diseño es el siguiente:

Se dividen los archivos en públicos y privados. En **ambiente local** esto
implica:

- Hay una ruta en el filesystem base `DirArchivosLocal` donde se almacenan los
archivos. La librería automáticamente crea las carpetas `public` y `private`
dentro de esta ruta.
- La librería genera URLs públicos y privados tomando como base las rutas
`RutaPubLocal` y `RutaPrivLocal`. Estos URLs resultan de concatenar la ruta
relativa dentro del directorio `public` o `private` con la ruta base.

> Nota: En ambiente GCP, la ruta de los archivos públicos es la siguiente:
> `https://<URL_CDN>/nombre_aplicacion/static/<ruta_archivo>`. Debido a que
> La URL del CDN y del balance loader pueden ser las mismas, hay que evitar
> utilizar la ruta /static dentro de los aplicativos en ambientes no locales.

En ambiente GCP, la librería espera que exista una cubeta para los archivos
públicos y otra para los privados. La cubeta de públicos debe tener habilitado
el servicio por CDN.

## Instalación

```bash
go get -u github.com/issemym-repos/archivos
```

## Configuración

Para el ambiente local, los archivos se almacenan en el filesystem local.

```go
package main

import (
    "github.com/issemym-repos/archivos"
)

func main() {
    config := &archivos.Config{
        Ambiente:         archivos.AmbienteLocal,
        DirArchivosLocal: "./static",        // Directorio base
        RutaPubLocal:     "/static/public",  // URL base para archivos públicos  
        RutaPrivLocal:    "/static/private", // URL base para archivos privados
    }
    
    manejador, err := archivos.NuevoManejador(config)
    if err != nil {
        panic(err)
    }
    defer manejador.Close()
}
```

### Configuración desde el ambiente

Las llaves de ambiente que se requieren son:
```bash
ARCHIVOS_AMBIENTE=LOCAL|DESARROLLO|PRODUCCION
ARCHIVOS_DIR_LOCAL=./static
ARCHIVOS_RUTA_PUB_LOCAL=/static/public
ARCHIVOS_RUTA_PRIV_LOCAL=/static/private

# Para ambientes GCP
GOOGLE_CLOUD_PROJECT=aplicativos-issemym-dev
ARCHIVOS_CUBETA_PUBLICA=bucket-publica
ARCHIVOS_CUBETA_PRIVADA=bucket-privada
ARCHIVOS_SERVICIO=SERVICIO  # el nombre del app/servicio
ARCHIVOS_URL_CDN=https://cdn.example.com
ARCHIVOS_EXPIRACION_URLs=15m  # expiración de URLs privados
```

Se proporciona una función de utilería para cargar/parsear estas llaves:

```go
func main() {
    config := archivos.ConfigDesdeEnv()
    manejador, err := archivos.NuevoManejador(config)
    if err != nil {
        panic(err)
    }
    defer manejador.Close()
}
```

## Uso básico

Se generan instancias de `Archivo` a partir de rutas que se definan:

```go
ctx := context.Background()

// Crear Archivos
archivoPublico := manejador.ArchivoPublico("css/style.css")           
archivoPrivado := manejador.ArchivoPrivado("docs/contrato.pdf")
```

Podemos entonces `Guardar` estos archivos, automáticamente se colocan en el
directorio o bucket correspondiente:

```go
err := archivoPublico.Guardar(ctx, cssReader)
if err != nil {
    // error de escritura del directorio o bucket
}
err := archivoPrivado.Guardar(ctx, pdfReader)
```

Podemos obtener una URL para acceder a estos archivos:

> Nota: En GCP para archivos privados los URLs son "signed" y duran
> 15 minutos por defecto.

```go
archivoPublico := manejador.ArchivoPublico("css/style.css")
url, err := archivoPublico.GetURL(ctx)
// Local: "/static/public/css/style.css"
// GCP: "https://cdn.example.com/SERVICIO/static/css/style.css"

archivoPrivado := manejador.ArchivoPrivado("documents/contract.pdf")
url, err := archivoPrivado.GetURL(ctx)
// Local: "/static/private/documents/contract.pdf"  
// GCP: "https://storage.googleapis.com/bucket/SERVICIO/documents/contract.pdf?X-Goog-..."
```

Siempre y cuando el archivo exista, podemos obtener un `io.ReadCloser` o
eliminar estos archivos:

```go
archivoPublico := manejador.ArchivoPublico("css/style.css")
reader, err := archivoPublico.Get(ctx)
if err == nil {
    defer reader.Close()
    contenido, _ := io.ReadAll(reader)
}

// El archivo existe?
existe, err := archivoPublico.Existe(ctx)

err = archivoPublico.Eliminar(ctx)
```

## Integración con templ y chi

Se provee de middleware y de una función llamada `Asset` para poder resolver
URLs públicos dentro de plantillas templ. Esta función Asset depende del
middleware para su funcionamiento correcto:

```go
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/issemym-repos/archivos"
)

func main() {
    config := archivos.ConfigDesdeEnv()
    manejadorArchivos, err := archivos.NuevoManejador(config)
    if err != nil {
        panic(err)
    }
    defer manejadorArchivos.Close()
    
    r := chi.NewRouter()
    
    // Usar el middleware del manejador para que `archivos.Asset(ctx, ruta)`
    // funcione correctamente
    r.Use(manejadorArchivos.Middleware)
    
    r.Get("/", handlerInicio)
    
    http.ListenAndServe(":8080", r)
}
```

## Uso en plantillas Templ

Ejemplo de uso de la función `Asset`, requiere de haber registrado el middleware
de manera previa:

```go
// inicio.templ
package templates

import "github.com/issemym-repos/archivos"

templ PaginaInicio() {
    <!DOCTYPE html>
    <html>
    <head>
        <title>Mi App</title>
        <link rel="stylesheet" href={ archivos.Asset(ctx, "css/style.css") }>
        <script src={ archivos.Asset(ctx, "js/app.js") }></script>
    </head>
    <body>
        <h1>Bienvenido</h1>
    </body>
    </html>
}
```

## Configuración para entorno GCP

```go
config := &archivos.Config{
    Ambiente:         archivos.AmbienteDesarrollo,
    IdProyectoGCP:   "aplicativos-issemym-dev",
    CubetaPublicaGCP: "mi-cubeta-publica-dev",
    CubetaPrivadaGCP: "mi-cubeta-privada-dev",
    // Se utiliza como subdirectorio en las cubetas
    Servicio:        "mi-aplicacion",              
    UrlCDN:          "https://cdn-dev.example.com",
    ExpiracionURLsFirmados: 30 * time.Minute,
    CacheControl: "public, max-age=3600"
}
```
