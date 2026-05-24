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
