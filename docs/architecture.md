# 📐 Arquitectura

La Plantilla Institucional utiliza una **Arquitectura Hexagonal** (también conocida como Puertos y Adaptadores). El objetivo principal es desacoplar la lógica de negocio (el core) de las preocupaciones de infraestructura (base de datos, frameworks web, etc.).

## 🧩 Capas de la Aplicación

### 1. Dominio (`types.go`)
Define las estructuras de datos puras y, en algunos casos, lógica de validación básica. No tiene dependencias externas.

### 2. Handler (Capa de Adaptación de Entrada)
Ubicada en cada módulo (ej. `configuracion/usuario/handler/`).
*   **Función**: Interpreta las peticiones HTTP/HTMX.
*   **Responsabilidad**: Validar inputs, llamar al servicio y renderizar la vista (Templ) o responder JSON.
*   **Interfaz**: Implementa los "Puertos de Entrada".

### 3. Service (Lógica de Negocio)
Ubicada en `service/`.
*   **Función**: Orquestar los casos de uso.
*   **Responsabilidad**: Reglas de negocio, cálculos complejos y coordinación entre múltiples repositorios.
*   **Interfaz**: Define qué necesita del exterior para funcionar (Puertos de Salida).

### 4. Repository (Capa de Adaptación de Salida)
Ubicada en `repository/`.
*   **Función**: Gestionar la persistencia.
*   **Responsabilidad**: Ejecutar queries SQL mediante `Pgxpool`.
*   **Aislamiento**: Solo esta capa conoce los detalles de la base de datos. Se debe evitar el uso de prefijos de esquema explícitos (`sistema.`, `institucion.`) para permitir la flexibilidad del `search_path` en entornos multi-tenant.

## 🔗 Gestión Descentralizada de Rutas (`pagina.NuevaRuta`)

Para asegurar la consistencia entre el ruteador del servidor (Chi) y las URLs en las vistas (HTMX/Templ) sin acoplar todos los módulos en un único archivo global, cada módulo declara su propia instancia en `<modulo>/rutas.go` usando el constructor `pagina.NuevaRuta`:

```go
package proveedor

import "PlantillaInstitucional/shared/pagina"

var Rutas = pagina.NuevaRuta("/institucion/proveedores")
```

*   **`Prefix`**: Ruta base sin slash final (ej. `/institucion/proveedores`), ideal para `r.Route()`.
*   **`Base`**: Ruta base con slash final (ej. `/institucion/proveedores/`), ideal para enlaces y formularios `POST`.
*   **Helpers Tipados**: Métodos como `GetModalEditar(id)`, `GetEditar(id)`, `GetBorrar(id)`, `GetReportePDF()` o `GetReporteExcel()` generan automáticamente URLs consistentes para HTMX y acciones de exportación.

## 💉 Inyección de Dependencias y Registro Modular (`app.App`)

El ensamblaje de dependencias está descentralizado por módulo mediante el contenedor de infraestructura `app.App` (`app/app.go`). Cada módulo expone una función autónoma `Register(a *app.App, r chi.Router)` en `<modulo>/handler/register.go`:

1.  **Instanciación Local**: El módulo construye su propio repositorio (`a.Pools`), servicio (`a.Logger`, `a.ManejadorArchivos`) y handler.
2.  **Montaje de Rutas**: El handler registra sus subrutas directamente sobre el router `chi.Router`.
3.  **Registro en Menú Lateral**: El módulo se registra dinámicamente en el menú de navegación mediante `navigation.RegistrarItemMenu(...)`.
4.  **Punto Único en `cmd/main.go`**: Agregar un módulo requiere únicamente una línea `<modulo>_handler.Register(a, r)` (automatizada por `jrmp new-module`).

## 📊 Ciclo de Vida y Retención de Reportes Privados

Los reportes generados (PDF y Excel) nunca se exponen en directorios públicos:
1.  **Generación en Memoria**: `GenerarReportePDF` (con traducción automática UTF-8 a CP1252 para acentos) y `GenerarReporteExcel` construyen el archivo validando límites de filas (`10,000`) y tamaño (`50 MiB`).
2.  **Registro Transaccional (`shared/reportes/lifecycle.go`)**: Antes de escribir el archivo, `StorePrivateReport` registra el metadato en `sistema.reportes_generados` en estado `pending` y luego lo activa (`active`) con expiración a 24 horas. En entorno `LOCAL`, resuelve automáticamente el UUID del tenant semilla si el contexto usa `"local"`.
3.  **Descarga Autenticada**: El endpoint `/private-reports/*` verifica que el usuario tenga sesión activa y que el segmento de esquema coincida con el `TenantContext` antes de transmitir el archivo con `Cache-Control: no-store`.
4.  **Barrido Automático (Sweep)**: Un worker en segundo plano reclama y elimina periódicamente los reportes expirados tanto del almacenamiento como de PostgreSQL.

## 🏢 Multi-tenancy y Aislamiento de Datos

La plantilla implementa un modelo de **Aislamiento en Defensa en Profundidad**:
1.  **Detección (Edge)**: El `TenantMiddleware` identifica y valida el tenant basado en el subdominio (o header autorizado si es un proxy de confianza) e inyecta el `TenantContext` inmutable en la petición.
2.  **Contexto Transaccional**: El acceso a base de datos se realiza exclusivamente a través de `TenantTxRunner`. Este inyecta de forma segura el esquema (vía `SET LOCAL search_path`) y la variable de sesión `app.tenant_id`. Todo se revierte al hacer commit o rollback.
3.  **Aislamiento RLS (Row Level Security)**: Cada tabla multi-tenant posee políticas `FORCE ROW LEVEL SECURITY`. Si un query bypassa el runner o no provee la identidad exacta, el motor de base de datos deniega el acceso automáticamente (fail-closed).
