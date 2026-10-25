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
- **Auditoría Continua de Vulnerabilidades**: Escaneo automatizado de código con `govulncheck` (análisis de árbol de llamadas) y de dependencias/contenedores con `Trivy` (`jrmp security` o `make security`), integrado en CI/CD.

---

## 🏁 Inicio Rápido

> 👉 Si vas a crear un **proyecto nuevo a partir de esta plantilla** (clonar, desvincular Git y renombrar con los datos de tu sistema), sigue primero la **[Guía Paso a Paso: De Cero a tu Propio Proyecto](./docs/primeros-pasos.md)**.

### Prerrequisitos
- Go 1.25 o superior
- Docker Desktop / OrbStack (debe estar abierto y en ejecución)
- [Templ CLI](https://templ.guide/quick-start/installation) (`go install github.com/a-h/templ/cmd/templ@latest`)
- [Air](https://github.com/air-verse/air) (`go install github.com/air-verse/air@latest`)
- **JRMP CLI** (`go install github.com/issemym-repos/PlantillaInstitucional/cmd/jrmp@latest`)

### Crear un Proyecto o Gestionar Módulos con la Interfaz Interactiva TUI (`jrmp`)

🟢 *Abre el panel interactivo navegable con flechas (`↑`/`↓`, `Enter`, `Tab`, `Esc`)*:
```bash
jrmp
```
*(También puedes abrirlo con `jrmp tui` o ejecutar directamente `jrmp create` para clonar la plantilla, desvincular Git, renombrar el proyecto con tus datos, generar el archivo `.env` y opcionalmente levantar Docker).*

### O Levantar un Proyecto Ya Clonado

1.  **Inicializar nombre y `.env`** (🟡 *Cambia `MiNuevoProyecto` por el nombre de tu sistema*):
    ```bash
    jrmp init MiNuevoProyecto
    ```

2.  **Construir y encender con Docker** (🟢 *Copiar y pegar tal cual*):
    ```bash
    jrmp up
    ```

3.  **Abrir en el navegador e iniciar sesión**:
    - **URL**: [http://localhost:8080/login](http://localhost:8080/login)
    - **Correo de prueba**: `admin@issemym.gob.mx`
    - **Contraseña de prueba**: `admin123`

---

## ⚡ Generación de Módulos CRUD (`jrmp new-module`)

Para crear un módulo CRUD completo (tabla, buscador, modales de alta/edición, exportación a **PDF** y **Excel**, pruebas unitarias, migración SQL con permisos y registro automático en el menú lateral):

### 1. Ejecutar el generador (interactivo o directo)
🟢 *Modo interactivo (te preguntará nombre, esquema y título)*:
```bash
jrmp new-module
```

🟡 *Modo directo (cambia **`proveedor`** por el nombre de tu módulo en singular)*:
```bash
jrmp new-module proveedor
```

O especificando el esquema de base de datos y el título con acentos:
🟡 *Modificar antes de ejecutar: cambia **`orden-compra`**, **`institucion`** y **`"Órdenes de Compra"`** por los datos de tu módulo*:
```bash
jrmp new-module orden-compra -schema institucion -title "Órdenes de Compra"
```

### 2. Eliminar un módulo de prueba (incluyendo reversión en PostgreSQL)
🟡 *Cambia **`proveedor`** por el módulo a limpiar*:
```bash
jrmp remove-module proveedor
```

### Parámetros Disponibles

| Parámetro | Obligatorio | Por defecto | Qué debes poner |
| :--- | :---: | :--- | :--- |
| `<nombre>` / `-name` | Sí | — | Nombre del módulo en singular (puedes usar guiones como `orden-compra`; se normaliza solo a `snake_case` y `PascalCase`). |
| `-schema` | No | `sistema` | Esquema de PostgreSQL donde se creará la tabla (`sistema` o `institucion`). |
| `-title` | No | Plural de `NAME` | Título con espacios y acentos para la pantalla, menú lateral y reportes. |

📖 Consulta la **[Guía Amigable para Crear y Personalizar Módulos](./docs/scaffolding.md)** (incluye cómo agregar campos nuevos y cómo borrar módulos de prueba).
