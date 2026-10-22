# 🚀 Guía de Desarrollo

Esta sección detalla cómo utilizar la plantilla para iniciar tu propio proyecto y cómo trabajar en el día a día.

## 🌟 Cómo Empezar tu Nuevo Proyecto

Si vas a usar esta plantilla para un proyecto nuevo, sigue estos pasos iniciales:

1.  **Clona el repositorio** (o usa el botón "Use this template" si está en GitHub).
2.  **Renombra el módulo Go**: Usa `find` y `sed` (o buscar y reemplazar en tu IDE) para cambiar `PlantillaInstitucional` por la ruta de tu nuevo módulo en `go.mod` y en todos los `.go` y `.templ`.
3.  **Configura tu entorno**:
    ```bash
    cp .env.example .env
    ```
    *(Ajusta las variables de base de datos, credenciales SMTP y el `DOMAIN` base de tu proyecto).*
4.  **Registra un Tenant (Institución)**:
    El sistema exige multi-tenancy. Debes crear el esquema y un registro inicial en la BD (o utilizar un script seed) antes de poder acceder por el navegador a tu subdominio (ej. `midominio.localhost`).

## 🛠️ Herramientas Recomendadas

*   **Go**: Versión 1.25+.
*   **Templ CLI**: Para generar código Go a partir de archivos `.templ`.
*   **GolangCI-Lint**: Para análisis estático de código.
*   **Docker Desktop / OrbStack**: Para gestionar el entorno completo.
*   **VS Code / GoLand / Neovim**: Con plugins de Go y Templ activados.

## 🏃 Flujo de Trabajo Local

### 1. Entorno Docker (Base de Datos y Servicios)
Para levantar el sistema con todos sus servicios (incluyendo la aplicación y PostgreSQL):
```bash
docker compose -f docker-compose.local.yml up --build -d
```

### 2. Comandos `Makefile` de Uso Diario

| Comando | Acción |
| :--- | :--- |
| `make new-module NAME=x [SCHEMA=y] [TITLE="Z"]` | Genera un módulo CRUD completo con vistas Templ v2, reportes PDF/Excel, migraciones SQL y registro en menú. |
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
