# 🚀 Guía de Desarrollo

Esta sección detalla cómo utilizar la plantilla para iniciar tu propio proyecto y cómo trabajar en el día a día.

## 🌟 Cómo Empezar tu Nuevo Proyecto

> 👉 Si es tu primera vez con Go, Git o Docker, sigue la **[Guía Paso a Paso: De Cero a tu Propio Proyecto](./primeros-pasos.md)**, donde cada paso viene explicado desde cero.

Resumen rápido de comandos para inicializar un nuevo sistema a partir de esta plantilla:

*   **Opción A — Automática desde cero con `jrmp create`** (🟢 *Copiar y pegar tal cual*):
    ```bash
    jrmp create
    ```
    *(Pregunta el nombre del proyecto, clona la plantilla, desvincula `origin`, renombra el módulo Go, genera el archivo `.env` y opcionalmente levanta Docker).*

*   **Opción B — Clonado manual + `jrmp init`** (🟡 *Cambia `MiNuevoProyecto` por el nombre de tu sistema*):
    ```bash
    git clone https://github.com/issemym-repos/PlantillaInstitucional.git MiNuevoProyecto
    cd MiNuevoProyecto
    jrmp init MiNuevoProyecto
    jrmp up
    ```
    Accede en [http://localhost:8080/login](http://localhost:8080/login) con `admin@issemym.gob.mx` / `admin123`.

## 🛠️ Herramientas Recomendadas

*   **Go**: Versión 1.25+.
*   **JRMP CLI (`jrmp`)**: CLI oficial del ecosistema (`go install github.com/issemym-repos/PlantillaInstitucional/cmd/jrmp@latest`).
*   **Templ CLI**: Para generar código Go a partir de archivos `.templ` (`go install github.com/a-h/templ/cmd/templ@latest`).
*   **Air**: Para recarga en vivo del servidor en desarrollo (`go install github.com/air-verse/air@latest`).
*   **GolangCI-Lint**: Para análisis estático de código.
*   **Docker Desktop / OrbStack**: Para gestionar el entorno completo.
*   **VS Code / GoLand / Neovim**: Con plugins de Go y Templ activados.

## 🏃 Flujo de Trabajo Local

### 1. Entorno Docker (Base de Datos y Servicios)
Para levantar el sistema con todos sus servicios (incluyendo la aplicación con recarga en vivo Air y PostgreSQL):
```bash
jrmp up
```

### 2. Interfaz TUI y Comandos Directos `jrmp` (Mac, Linux y Windows)

| Comando | Acción |
| :--- | :--- |
| `jrmp` / `jrmp tui` | Abre la interfaz interactiva en terminal (TUI estilo `gentle-ai`) con menús y formularios paso a paso. |
| `jrmp create [NombreProyecto]` | Clona la plantilla desde cero, desvincula Git, renombra el proyecto y configura `.env`. |
| `jrmp init [NombreProyecto]` | Renombra el módulo Go en un repositorio ya clonado, desvincula `origin` y genera `.env`. |
| `jrmp new-module [nombre] [-schema y] [-title "Z"]` | Genera un módulo CRUD completo con vistas Templ v2, reportes PDF/Excel, pruebas unitarias, migraciones SQL y registro en menú. |
| `jrmp remove-module <nombre>` | Elimina un módulo generado, sus migraciones, su registro en `cmd/main.go` y revierte su tabla/políticas en PostgreSQL local. |
| `jrmp up` / `jrmp down` / `jrmp logs` | Enciende, apaga o muestra logs del entorno local en Docker. |
| `jrmp dev` | Ejecuta Air fuera de Docker resolviendo automáticamente `.exe` en Windows. |
| `jrmp templ-fmt` | Formatea todos los archivos `.templ`. |
| `jrmp templ-check` | Verifica formato de `.templ` y ejecuta `templ generate`. |
| `jrmp lint` | Ejecuta `golangci-lint run ./...`. |
| `jrmp test` | Ejecuta todas las pruebas unitarias (`go test ./...`). |
| `jrmp check` | Ejecuta la validación completa (`templ-check` + `lint` + `test`). |
| `jrmp security` | Ejecuta la auditoría completa de seguridad (`govulncheck` + `trivy`). |
| `jrmp vulncheck` | Ejecuta `govulncheck` en el proyecto raíz y en `cmd/jrmp`. |
| `jrmp trivy` | Ejecuta el escáner de seguridad Trivy en el repositorio (omitiendo `vendor`). |

### 3. Live Reload y Generación
Ejecuta la app usando [Air](https://github.com/air-verse/air) para recompilación automática:
```bash
air
```
Si modificas archivos `.templ`, asegúrate de correr `jrmp templ-check` (o `templ generate`).

## 📊 Reportes PDF, Excel y Assets

El sistema de reportes (`shared/reportes`) genera y almacena archivos privados con control de retención:
*   **PDF (`generarReportePDF.go`)**: Utiliza `gofpdf` con la tipografía institucional **Gotham** en formato TrueType (`.ttf`) con soporte nativo UTF-8 (`SplitText`), preservando acentos, diéresis y eñes sin distorsión. Cuenta con fallback automático transparente a `Arial` (CP-1252) si faltan los archivos y protección `asegurarFuenteValida` contra estados no inicializados.
*   **Excel (`generarReporteExcel.go`)**: Utiliza `excelize/v2` con formato institucional, autoajuste de columnas y encabezados con estilo.
*   **Imágenes**: Logos institucionales en `static/public/img/`.

## 🛡️ Auditoría de Seguridad (govulncheck y Trivy)

El proyecto incluye escaneo automatizado continuo tanto en desarrollo local como en el pipeline de CI/CD:

1.  **`govulncheck`**:
    *   Herramienta oficial de Go que analiza el grafo real de llamadas (`call-graph analysis`) para detectar si el código efectivamente invoca símbolos con vulnerabilidades reportadas en el *Go Vulnerability Database*.
    *   Ejecución local: `jrmp vulncheck` (o `make vulncheck`).
2.  **`Trivy`**:
    *   Escáner de seguridad de Aqua Security para detectar vulnerabilidades en dependencias, configuraciones de contenedores (`Dockerfile`) y secretos expuestos.
    *   Excluye automáticamente el directorio `vendor/` para evitar falsos positivos de pruebas upstream.
    *   Ejecución local: `jrmp trivy` (o `make trivy`; si no tienes `trivy` instalado, utiliza automáticamente el contenedor `aquasec/trivy:latest` vía Docker).
3.  **Auditoría Integral**:
    *   `jrmp security` (o `make security`) ejecuta ambos análisis secuencialmente.
    *   En GitHub Actions (`.github/workflows/ci.yml`), el job `security` corre en paralelo en cada push y pull request.

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
