# 🗄️ Base de Datos y Migraciones

La Plantilla Institucional utiliza **PostgreSQL** como motor relacional primario. La gestión se basa en esquemas para soportar multi-tenancy y migraciones versionadas.

## 🗂️ Estructura de Esquemas

El proyecto utiliza múltiples esquemas para organizar los datos:
*   `public`: Esquema por defecto y datos globales de tenants.
*   `sistema`: Configuración transversal del sistema (parámetros).
*   `sesion`: Almacenamiento de sesiones persistentes.
*   `institucion`: Datos de negocio (ej. catálogo de personas).

## 🚀 Gestión de Migraciones

Utilizamos `golang-migrate` para asegurar que todos los entornos tengan la misma estructura de base de datos.

### Ubicación
Las migraciones se encuentran en `sql/migrations/`.
*   `.up.sql`: Cambios para aplicar (crear tablas, agregar columnas).
*   `.down.sql`: Cambios para revertir.

### Ejecución Automática
Al iniciar la aplicación (`cmd/main.go`), el sistema intenta ejecutar todas las migraciones pendientes automáticamente usando la URL de conexión de escritura (`PQ_WRITE_HOST`).

*   **Entornos de Producción / Serverless**: En entornos como **Google Cloud Run** donde el servidor escala horizontalmente a múltiples instancias concurrentes, la ejecución automática simultánea al bootear puede generar bloqueos competitivos (`locks`) y latencias innecesarias (Cold Start).
*   **SKIP_MIGRATIONS**: Podés setear `SKIP_MIGRATIONS=true` en tu consola de Cloud Run para desactivar este comportamiento. De esta forma, las migraciones se deben correr de forma controlada y secuencial mediante un paso previo en tu pipeline de **Google Cloud Build** o utilizando un **Cloud Run Job** aislado.

### Crear Nueva Migración
```bash
migrate create -ext sql -dir sql/migrations -seq nombre_de_la_migracion
```

## 🏗️ Repositorios y Consultas

Los repositorios utilizan `Pgxpool` para una gestión eficiente de conexiones.

### Patrón de Repositorio:
1.  **Queries**: Se recomienda definir las constantes SQL en un archivo `queries.go` separado.
2.  **Transacciones**: Siempre usar el contexto (`ctx`) para permitir la propagación de timeouts y cancelaciones.
3.  **Audit**: Los repositorios extraen automáticamente el UUID del usuario del contexto si está disponible para logs o campos de auditoría.

### Multi-tenancy en BD:
Cada repositorio es "agnóstico" al esquema del tenant gracias al middleware que ejecuta `SET search_path` antes de entregar una conexión del pool. Esto significa que puedes escribir `SELECT * FROM mi_tabla` y PostgreSQL buscará automáticamente en el esquema del tenant actual.
