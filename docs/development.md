# 🚀 Guía de Desarrollo

Esta sección detalla cómo utilizar la plantilla para iniciar tu propio proyecto y cómo trabajar en el día a día.

## 🌟 Cómo Empezar tu Nuevo Proyecto

> 👉 Si es tu primera vez con Go, Git o Docker, sigue la **[Guía Paso a Paso: De Cero a tu Propio Proyecto](./primeros-pasos.md)**, donde cada paso viene explicado desde cero.

Resumen rápido de comandos para inicializar un nuevo sistema a partir de esta plantilla:

1.  **Clonar el repositorio con el nombre de tu nuevo proyecto** (🟡 *Cambia `MiNuevoProyecto` por el nombre de tu sistema*):
    ```bash
    git clone https://github.com/issemym-repos/PlantillaInstitucional.git MiNuevoProyecto
    cd MiNuevoProyecto
    git remote remove origin
    ```
2.  **Renombrar el módulo Go en todo el código** (🟡 *Cambia `MiNuevoProyecto` por el nombre de tu sistema*):
    ```bash
    make init-project MODULE=MiNuevoProyecto
    ```
3.  **Configurar tu entorno local** (🟢 *Copiar y pegar tal cual*):
    ```bash
    cp .env.example .env
    ```
4.  **Levantar con Docker** (🟢 *Copiar y pegar tal cual*):
    ```bash
    docker compose -f docker-compose.local.yml up --build -d
    ```
    Accede en [http://localhost:8080/login](http://localhost:8080/login) con `admin@issemym.gob.mx` / `admin123`.

## 🛠️ Herramientas Recomendadas

*   **Go**: Versión 1.25+.
*   **Templ CLI**: Para generar código Go a partir de archivos `.templ` (`go install github.com/a-h/templ/cmd/templ@latest`).
*   **Air**: Para recarga en vivo del servidor en desarrollo (`go install github.com/air-verse/air@latest`).
*   **GolangCI-Lint**: Para análisis estático de código.
*   **Docker Desktop / OrbStack**: Para gestionar el entorno completo.
*   **VS Code / GoLand / Neovim**: Con plugins de Go y Templ activados.

## 🏃 Flujo de Trabajo Local

### 1. Entorno Docker (Base de Datos y Servicios)
Para levantar el sistema con todos sus servicios (incluyendo la aplicación con recarga en vivo Air y PostgreSQL):
```bash
docker compose -f docker-compose.local.yml up --build -d
```

### 2. Comandos `Makefile` de Uso Diario

| Comando | Acción |
| :--- | :--- |
| `make init-project MODULE=MiNuevoProyecto` | Renombra el módulo Go del proyecto en `go.mod` y en todos los archivos `.go` y `.templ`. |
| `make new-module NAME=x [SCHEMA=y] [TITLE="Z"]` | Genera un módulo CRUD completo con vistas Templ v2, reportes PDF/Excel, pruebas unitarias, migraciones SQL y registro en menú. |
| `make remove-module NAME=x` | Elimina un módulo generado, sus migraciones, su registro en `cmd/main.go` y revierte su tabla/políticas en PostgreSQL local. |
| `make templ-fmt` | Formatea todos los archivos `.templ`. |
| `make templ-check` | Verifica formato de `.templ` y ejecuta `templ generate`. |
| `make lint` | Ejecuta `golangci-lint run ./...`. |
| `make test` | Ejecuta todas las pruebas unitarias (`go test ./...`). |
| `make check` | Ejecuta la validación completa (`templ-check` + `lint` + `test`). |

### 3. Live Reload y Generación
Ejecuta la app usando [Air](https://github.com/air-verse/air) para recompilación automática:
```bash
air
```
Si modificas archivos `.templ`, asegúrate de correr `templ generate` (o `make templ-check`).

## 📊 Reportes PDF, Excel y Assets

El sistema de reportes (`shared/reportes`) genera y almacena archivos privados con control de retención:
*   **PDF (`generarReportePDF.go`)**: Utiliza `gofpdf` con traducción automática de caracteres UTF-8 a CP1252 (`UnicodeTranslatorFromDescriptor`), soportando acentos, diéresis y eñes en títulos, encabezados y celdas sin configuración adicional.
*   **Excel (`generarReporteExcel.go`)**: Utiliza `excelize/v2` con formato institucional, autoajuste de columnas y encabezados con estilo.
*   **Imágenes**: Logos institucionales en `static/public/img/`.

## 🧪 Pruebas (Testing)

Ejecuta todas las pruebas con validación exhaustiva:
```bash
go test -count=1 ./...
```

## 📦 Gestión de Dependencias

El proyecto utiliza el directorio `vendor/` para garantizar compilaciones reproducibles dentro y fuera de Docker:
```bash
go mod tidy
go mod vendor
```

---

## 📝 Convenciones Críticas

1.  **Transacciones Multi-tenant Obligatorias**: NUNCA accedas directamente a un pool de base de datos en repositorios de dominio. Todos los repositorios DEBEN ejecutar operaciones a través de `r.db.Tenant.Read(ctx, func(tx pgx.Tx) error)` o `r.db.Tenant.Write(...)`.
2.  **Cero prefijos de esquema en queries de negocio**: Evita acoplar queries al esquema salvo cuando sea requerido explícitamente; `TenantTxRunner` y `search_path` gestionan la resolución junto con `FORCE ROW LEVEL SECURITY`.
3.  **Rutas Descentralizadas**: NUNCA hardcodees rutas como strings manuales en vistas o handlers. Declara `var Rutas = pagina.NuevaRuta(...)` en el paquete raíz de cada módulo (consulta la [Guía de Scaffolding](./scaffolding.md)).
4.  **Naming**: Normaliza nombres de paquetes, tablas y rutas en `snake_case`, y structs o componentes exportados en `PascalCase`.
