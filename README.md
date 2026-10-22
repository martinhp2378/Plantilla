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
    Si no usas Docker para el desarrollo, recordá generar los componentes:
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

Para crear un nuevo módulo CRUD completo y desacoplado (entidad, repositorio, servicio, handlers, vistas Templ v2, migraciones SQL con ABAC y auto-montaje en `cmd/main.go`):

```bash
make new-module NAME=proveedor
```

Parámetros opcionales:
- `SCHEMA`: esquema de base de datos (por defecto `institucion`).
- `TITLE`: título legible para la vista (por defecto capitalizado en plural).

Ejemplo:
```bash
make new-module NAME=articulo SCHEMA=inventario TITLE="Artículos de Almacén"
```

El generador crea:
1. `/<nombre>`: Entidad y rutas canónicas (`types.go`, `rutas.go`).
2. `/<nombre>/repository`: Consultas tipadas con paginación, filtros dinámicos y ordenamiento seguro.
3. `/<nombre>/service`: Reglas de negocio y validación de datos.
4. `/<nombre>/handler`: Handlers HTTP REST/HTMX y función autónoma `Register(a *app.App, r chi.Router)`.
5. `/<nombre>/view`: Componentes Templ v2 (tabla con filtros/búsqueda, modales de alta/edición y layout base).
6. `/sql/migrations`: Migración `.up.sql` y `.down.sql` con auditoría, aislamiento multi-tenant y políticas ABAC.
7. Auto-montaje de una sola línea en `cmd/main.go`.
