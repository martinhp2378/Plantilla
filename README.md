# 🏛️ Plantilla Institucional

Una base sólida y moderna para el desarrollo de aplicaciones institucionales robustas, construida con el poder de **Go**, la interactividad de **HTMX** y la seguridad de tipos de **Templ**.

---

## 🚀 Filosofía del Proyecto

Esta plantilla ha sido diseñada para reducir el tiempo de arranque de proyectos institucionales, forzando buenas prácticas arquitectónicas y proporcionando componentes listos para producción.

### Pilares Clave:
- **Productividad**: Desarrollo rápido con recarga en vivo (Air).
- **Simplicidad**: Menos JavaScript, más HTML interactivo (HTMX).
- **Seguridad**: Tipado fuerte desde la base de datos hasta la capa de vista (Templ).
- **Escalabilidad**: Arquitectura hexagonal lista para crecer.

---

## 🛠️ Stack Tecnológico

| Tecnología | Propósito |
| :--- | :--- |
| **Go (1.22+)** | Lenguaje core de alto rendimiento. |
| **Chi** | Router HTTP ligero e idiomático. |
| **Templ** | Generación de componentes HTML con seguridad de tipos. |
| **HTMX** | Interactividad AJAX moderna sin frameworks pesados. |
| **Tailwind CSS** | Estilizado rápido y consistente. |
| **PostgreSQL** | Motor de base de datos relacional robusto. |
| **Pgx** | Driver de PostgreSQL con soporte nativo para pools. |
| **Alpine.js** | Interactividad mínima en el cliente (modales, tabs). |

---

## 📐 Arquitectura

El proyecto sigue una **Arquitectura Hexagonal (Puertos y Adaptadores)** simplificada:

1.  **Handler**: Capa de transporte (HTTP/HTMX). Se encarga de recibir peticiones y renderizar respuestas.
2.  **Service**: Lógica de negocio pura. Orquestación de casos de uso.
3.  **Repository**: Adaptadores de persistencia. Consultas SQL específicas.
4.  **View**: Componentes Templ y Layouts base.

### Multi-tenancy
Soporte nativo para aislamiento de datos por esquemas de base de datos, identificado mediante subdominios.

---

## 📁 Estructura del Proyecto

```bash
├── app/                  # Lógica central y Middlewares globales
├── cmd/                  # Punto de entrada y configuración de DI
├── configuracion/        # Módulos transversales (Usuarios, Perfiles, Políticas)
├── institucion/          # Módulos de dominio institucional
├── shared/               # Utilidades compartidas y componentes UI globales
├── sql/                  # Migraciones de base de datos
├── static/               # Assets estáticos (CSS, JS, Imágenes)
└── vendor/               # Dependencias locales
```

---

## 🏁 Inicio Rápido

### Prerrequisitos
- Go 1.22 o superior
- Docker y Docker Compose
- [Templ CLI](https://templ.guide/quick-start/installation)

### Configuración Local

1.  **Preparar el Entorno**:
    ```bash
    cp .env.example .env
    ```

2.  **Levantar Base de Datos**:
    ```bash
    docker-compose -f docker-compose.local.yml up -d db
    ```

3.  **Generar Código Templ**:
    ```bash
    templ generate
    ```

4.  **Ejecutar la Aplicación**:
    ```bash
    go run cmd/main.go
    ```

---

## 📦 Despliegue y Migraciones

### Migraciones
Utilizamos `golang-migrate`. Las migraciones se ejecutan automáticamente al iniciar la aplicación.

Para crear una nueva migración:
```bash
migrate create -ext sql -dir sql/migrations -seq nombre_de_la_migración
```

---

## 📜 Licencia

MIT - Ver el archivo [LICENSE](LICENSE) para más detalles.
