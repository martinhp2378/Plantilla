# 🏛️ Plantilla Institucional

Una base sólida y moderna para el desarrollo de aplicaciones institucionales robustas, construida con el poder de **Go**, la interactividad de **HTMX** y la seguridad de tipos de **Templ**.

> [!TIP]
> **¿Es tu primera vez usando esta plantilla, Go, Git o Docker?**
> Comienza aquí con nuestras guías paso a paso diseñadas con comandos listos para copiar y pegar:
> 1. 🌱 **[Guía Paso a Paso: De Cero a tu Propio Proyecto (Clonar, Renombrar y Levantar)](./docs/primeros-pasos.md)**
> 2. ⚡ **[Guía Amigable: Cómo Crear y Personalizar un Nuevo Módulo CRUD](./docs/scaffolding.md)**
> 3. 📚 **[Índice Completo de Documentación](./docs/index.md)**

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

> 👉 Si vas a crear un **proyecto nuevo a partir de esta plantilla** (clonar, desvincular Git y renombrar con los datos de tu sistema), sigue primero la **[Guía Paso a Paso: De Cero a tu Propio Proyecto](./docs/primeros-pasos.md)**.

### Prerrequisitos
- Go 1.25 o superior
- Docker Desktop / OrbStack (debe estar abierto y en ejecución)
- [Templ CLI](https://templ.guide/quick-start/installation)

### Levantar el Proyecto en tu Computadora

1.  **Crear el archivo de configuración local** (🟢 *Copiar y pegar tal cual*):
    ```bash
    cp .env.example .env
    ```

2.  **Construir y encender con Docker** (🟢 *Copiar y pegar tal cual*):
    ```bash
    docker compose -f docker-compose.local.yml up --build -d
    ```

3.  **Abrir en el navegador e iniciar sesión**:
    - **URL**: [http://localhost:8080/login](http://localhost:8080/login)
    - **Correo de prueba**: `admin@issemym.gob.mx`
    - **Contraseña de prueba**: `admin123`

---

## ⚡ Generación de Módulos CRUD (Scaffolding)

Para crear un módulo CRUD completo (tabla, buscador, modales de alta/edición, exportación a **PDF** y **Excel**, migración SQL con permisos y registro automático en el menú lateral):

### 1. Ejecutar el generador
🟡 *Modificar antes de ejecutar: cambia **`proveedor`** por el nombre de tu módulo en singular*:
```bash
make new-module NAME=proveedor
```

O especificando el esquema de base de datos y el título con acentos:
🟡 *Modificar antes de ejecutar: cambia **`orden-compra`**, **`institucion`** y **`"Órdenes de Compra"`** por los datos de tu módulo*:
```bash
make new-module NAME=orden-compra SCHEMA=institucion TITLE="Órdenes de Compra"
```

### 2. Actualizar el contenedor para ver el nuevo módulo en el navegador
🟢 *Copiar y pegar tal cual*:
```bash
docker compose -f docker-compose.local.yml build --no-cache app && docker compose -f docker-compose.local.yml up -d app
```

### Parámetros Disponibles

| Parámetro | Obligatorio | Por defecto | Qué debes poner |
| :--- | :---: | :--- | :--- |
| `NAME` | Sí | — | Nombre del módulo en singular (puedes usar guiones como `orden-compra`; se normaliza solo a `snake_case` y `PascalCase`). |
| `SCHEMA` | No | `sistema` | Esquema de PostgreSQL donde se creará la tabla (`sistema` o `institucion`). |
| `TITLE` | No | Plural de `NAME` | Título con espacios y acentos para la pantalla, menú lateral y reportes. |

📖 Consulta la **[Guía Amigable para Crear y Personalizar Módulos](./docs/scaffolding.md)** (incluye cómo agregar campos nuevos y cómo borrar módulos de prueba).
