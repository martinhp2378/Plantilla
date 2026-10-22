# 🏛️ Plantilla Institucional

Una base sólida y moderna para el desarrollo de aplicaciones institucionales robustas, construida con el poder de **Go**, la interactividad de **HTMX** y la seguridad de tipos de **Templ**.

---

## 🚀 Filosofía del Proyecto

Esta plantilla ha sido diseñada para reducir el tiempo de arranque de proyectos institucionales, forzando buenas prácticas arquitectónicas y proporcionando componentes listos para producción.

### Pilares Clave:
- **Productividad**: Desarrollo rápido con recarga en vivo (Air).
- **Simplicidad**: Menos JavaScript, más HTML interactivo (HTMX).
- **Seguridad Proactiva**: Tipado fuerte y protección integrada contra CSRF y Clickjacking.
- **Escalabilidad**: Arquitectura hexagonal lista para crecer.

---

## 🛠️ Stack Tecnológico

| Tecnología | Propósito |
| :--- | :--- |
| **Go (1.25+)** | Lenguaje core de alto rendimiento. |
| **Chi v5** | Router HTTP ligero e idiomático. |
| **Templ** | Generación de componentes HTML con seguridad de tipos. |
| **HTMX v2** | Interactividad AJAX moderna sin frameworks pesados. |
| **Tailwind CSS v4** | Estilizado rápido y consistente. |
| **PostgreSQL 16** | Motor de base de datos relacional robusto. |
| **Pgx v5** | Driver de PostgreSQL con soporte nativo para pools. |
| **Alpine.js** | Interactividad mínima en el cliente (modales, tabs). |

---

## 🛡️ Seguridad

El sistema incluye una capa de hardenización por defecto:
- **CSRF**: Protección global mediante `nosurf`, integrada automáticamente con HTMX.
- **Headers**: Implementación de cabeceras OWASP (HSTS, X-Frame-Options, X-Content-Type-Options).
- **Validación ABAC**: Middleware de políticas basado en perfiles y rutas.

---

## 🏁 Inicio Rápido

### Prerrequisitos
- Go 1.25 o superior
- Docker Desktop / OrbStack
- [Templ CLI](https://templ.guide/quick-start/installation)

### Configuración Local

1.  **Preparar el Entorno**:
    ```bash
    cp .env.example .env
    ```

2.  **Levantar con Docker**:
    El proyecto está configurado para sincronizar las variables del `.env` directamente con los contenedores.
    ```bash
    docker compose -f docker-compose.local.yml up --build -d
    ```

3.  **Generar Código Templ (Manual)**:
    Si no usas Docker para el desarrollo, recuerda generar los componentes:
    ```bash
    templ generate
    ```

---

## 📦 Migraciones

Las migraciones se ejecutan automáticamente al iniciar la aplicación. Para crear una nueva:
```bash
migrate create -ext sql -dir sql/migrations -seq nombre_de_la_migración
```

---

## ⚡ Generación de Módulos (Scaffolding)

Para crear un nuevo módulo CRUD vertical completo y autónomo (entidad, repositorio multi-tenant, servicio, reportes PDF/Excel, handlers, vistas Templ v2, migraciones SQL con RLS/ABAC, registro en el menú lateral y auto-montaje en `cmd/main.go`):

```bash
make new-module NAME=proveedor
```

### Parámetros Disponibles

| Parámetro | Obligatorio | Por defecto | Descripción |
| :--- | :---: | :--- | :--- |
| `NAME` | Sí | — | Nombre del módulo en singular (normaliza automáticamente guiones o espacios a `snake_case` y `PascalCase`). |
| `SCHEMA` | No | `sistema` | Esquema de base de datos en PostgreSQL (`sistema`, `institucion`, etc.). |
| `TITLE` | No | Plural de `NAME` | Título legible para la interfaz, menú de navegación y encabezados de reportes. |

Ejemplo personalizado:
```bash
make new-module NAME=orden-compra SCHEMA=institucion TITLE="Órdenes de Compra"
```

### Qué Incluye el Módulo Generado
1. **`/<modulo>`**: Entidad de dominio y rutas canónicas descentralizadas (`types.go`, `rutas.go`).
2. **`/<modulo>/repository`**: Consultas tipadas sobre `TenantTxRunner` con paginación, filtros dinámicos, ordenamiento seguro y soporte transparente para desarrollo local.
3. **`/<modulo>/service`**: Reglas de negocio, validación con `validator/v10` y exportación nativa a **PDF** (con soporte de acentos/UTF-8) y **Excel** (`reportes.go`).
4. **`/<modulo>/handler`**: Handlers HTTP/HTMX, descarga privada de reportes y función `Register(a *app.App, r chi.Router)` que monta rutas y registra el módulo en el menú lateral.
5. **`/<modulo>/view`**: Componentes Templ v2 (tabla dinámica con búsqueda/filtros, botones de exportación, modales de alta/edición apuntando a `#modal` y layout institucional).
6. **`/sql/migrations`**: Migraciones `.up.sql` y `.down.sql` secuenciales con triggers de auditoría, `FORCE ROW LEVEL SECURITY` y políticas ABAC para `superadmin` y `admin`.
7. **Auto-montaje en `cmd/main.go`**: Inyección automática del import y registro en una sola línea.

📖 Consulta la [Guía Completa de Scaffolding](./docs/scaffolding.md) y el [Índice de Documentación](./docs/index.md) para más detalles.
