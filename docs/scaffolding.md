# ⚡ Generación de Módulos (Scaffolding)

El generador de módulos (`cmd/tools/scaffold`) crea un módulo CRUD vertical completo —desde la migración SQL con políticas ABAC y Row Level Security (RLS) hasta las vistas interactivas con HTMX/Templ v2 y exportación de reportes en PDF y Excel— y lo monta automáticamente en la aplicación.

## Ruta Rápida (Quick Path)

1. **Generar el módulo**:
   ```bash
   make new-module NAME=proveedor SCHEMA=institucion TITLE="Proveedores"
   ```
2. **Compilar componentes Templ y verificar integridad**:
   ```bash
   make check
   ```
3. **Reconstruir y levantar el contenedor**:
   ```bash
   docker compose -f docker-compose.local.yml up --build -d app
   ```

Al iniciar la aplicación, la migración SQL se ejecuta automáticamente, el módulo aparece en el menú lateral y las rutas CRUD y de reportes quedan operativas.

---

## Parámetros del Generador

| Parámetro | Obligatorio | Valor por defecto | Descripción |
| :--- | :---: | :--- | :--- |
| `NAME` | Sí | — | Nombre de la entidad en singular (ej. `proveedor`, `orden_compra` o `orden-compra`). |
| `SCHEMA` | No | `sistema` | Esquema de PostgreSQL donde se creará la tabla (`sistema`, `institucion`, etc.). |
| `TITLE` | No | Plural de `NAME` | Título visible en la interfaz, encabezados de reportes y menú lateral. |

### Normalización Automática de Identificadores

El generador normaliza cualquier entrada con guiones (`-`) o espacios para garantizar código Go, rutas y objetos SQL válidos:

| Entrada (`NAME`) | Paquete / Tabla SQL (`snake_case`) | Structs / Componentes Templ (`PascalCase`) | Ruta HTTP Base |
| :--- | :--- | :--- | :--- |
| `proveedor` | `proveedor` | `Proveedor` | `/sistema/proveedores` |
| `orden-compra` | `orden_compra` | `OrdenCompra` | `/sistema/orden_compras` |
| `tipo documento` | `tipo_documento` | `TipoDocumento` | `/sistema/tipo_documentos` |

---

## Anatomía del Módulo Generado

Cada módulo es autónomo y sigue **Arquitectura Hexagonal / Vertical Slicing**:

| Archivo / Directorio | Capa | Responsabilidad |
| :--- | :--- | :--- |
| `<modulo>/types.go` | Dominio | Entidad principal, filtros de búsqueda y estructura de resultados paginados. |
| `<modulo>/rutas.go` | Dominio / Rutas | Definición canónica de rutas mediante `pagina.NuevaRuta("/<schema>/<plural>")`. |
| `<modulo>/repository/` | Adaptador Salida | Consultas SQL parametrizadas sobre `TenantTxRunner` (lectura/escritura aislada por tenant). |
| `<modulo>/service/service.go` | Aplicación | Validación de negocio (`validator/v10`), normalización y orquestación transaccional. |
| `<modulo>/service/reportes.go` | Aplicación | Generación de reportes **PDF** y **Excel** conectados al almacenamiento privado. |
| `<modulo>/handler/crud.go` | Adaptador Entrada | Endpoints HTTP/HTMX para crear (`POST`), actualizar (`PUT`) y eliminar (`DELETE`). |
| `<modulo>/handler/views.go` | Adaptador Entrada | Renderizado de página completa (`Index`), tabla parcial (`SearchTable`) y modales. |
| `<modulo>/handler/files.go` | Adaptador Entrada | Descarga de reportes (`/reportes/pdf`, `/reportes/excel`) con redirección segura `303 See Other`. |
| `<modulo>/handler/register.go` | Ensamblaje | Función `Register(a *app.App, r chi.Router)` que inyecta dependencias y registra el ítem en el menú lateral. |
| `<modulo>/view/` | Presentación | Vistas Templ v2 (`layouts/`, `components/table_v2.templ`, `components/form_v2.templ`). |
| `sql/migrations/` | Base de Datos | Archivos `.up.sql` y `.down.sql` secuenciales con tabla, índices, triggers de auditoría, RLS y permisos ABAC. |
| `cmd/main.go` | Punto de Entrada | Inyección automática del `import` y la llamada `<modulo>_handler.Register(a, r)`. |

---

## Funcionalidades Incluidas de Fábrica

### 1. CRUD Interactivo con HTMX y Templ v2
- **Búsqueda y filtrado en vivo**: Filtrado por texto y estatus con debounce (`delay:300ms`) y actualización parcial del contenedor `#<modulo>-table-container`.
- **Paginación dinámica**: Navegación por páginas preservando los filtros activos (`hx-include="#filter-form"`).
- **Modales desacoplados**: Alta y edición cargadas dinámicamente en `#modal` (`ModalTarget: "#modal"`), con cierre automático al guardar y refresco de tabla más notificación toast OOB (`hx-swap-oob="true"`).

### 2. Exportación de Reportes (PDF y Excel)
- **Botones integrados en Toolbar**: La vista principal incluye acciones directas para exportar a PDF y Excel mediante `export-actions.js`.
- **Soporte completo de acentos y caracteres en español**: El motor PDF (`shared/reportes/generarReportePDF.go`) traduce automáticamente cadenas UTF-8 a CP1252 mediante `UnicodeTranslatorFromDescriptor`, garantizando que tildes (`á, é, í, ó, ú`), diéresis (`ü`) y eñes (`ñ, Ñ`) se representen y midan correctamente.
- **Almacenamiento privado y retención**: Los archivos se guardan mediante `StorePrivateReport` bajo `/private-reports/reportes/<schema>/...` y su ciclo de vida se registra en `sistema.reportes_generados` con expiración y limpieza automática.

### 3. Aislamiento Multi-Tenant y Desarrollo Local
- **Producción**: Tanto las consultas de repositorio como la retención de reportes utilizan el UUID del tenant activo e inyectan `app.tenant_id` en la transacción local para cumplir con `FORCE ROW LEVEL SECURITY`.
- **Entorno Local (`LOCAL`)**: Cuando el contexto de desarrollo local utiliza el identificador `"local"`, el repositorio del módulo y el gestor de retención de reportes resuelven automáticamente el UUID del tenant semilla en `institucion.cat_tenants`, permitiendo probar altas, ediciones y descargas de reportes sin errores de casteo UUID (`22P02`).

---

## Cómo Revertir un Módulo de Prueba

Si generaste un módulo únicamente para pruebas locales y deseas eliminarlo antes de hacer commit:

1. **Revertir la migración en PostgreSQL** (suponiendo que la migración generada fue la `000011`):
   ```bash
   docker exec plantillainstitucional-db-1 psql -U JRosendo -d plantilla -c "
   DELETE FROM sesion.politica WHERE recurso = '<nombre_modulo>';
   DELETE FROM sesion.recurso WHERE nombre = '<nombre_modulo>';
   DROP TABLE IF EXISTS <schema>.<nombre_modulo> CASCADE;
   UPDATE schema_migrations SET version = 10, dirty = false;
   "
   ```
2. **Borrar los archivos generados y restaurar `cmd/main.go`**:
   ```bash
   rm -rf <nombre_modulo> sql/migrations/000011_create_<nombre_modulo>_table.*
   git checkout cmd/main.go
   ```

---

## Lista de Verificación para Personalizar un Módulo Nuevo

- [ ] **Campos de Dominio**: Ajustar los campos en `<modulo>/types.go` y en la migración `.up.sql` según los requerimientos del negocio.
- [ ] **Consultas SQL**: Actualizar `SELECT`, `INSERT` y `UPDATE` en `<modulo>/repository/queries.go` y `<modulo>/repository/repository.go`.
- [ ] **Formulario y Tabla**: Añadir o modificar los inputs en `<modulo>/view/components/form_v2.templ` y las columnas en `table_v2.templ`.
- [ ] **Encabezados de Reporte**: Sincronizar `<modulo>ReportHeaders` y `<modulo>ReportFields` en `<modulo>/service/reportes.go`.
- [ ] **Validación Completa**: Ejecutar `make check` (`templ fmt`, `templ generate`, `golangci-lint` y `go test ./...`) antes de abrir un Pull Request.
