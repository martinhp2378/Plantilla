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

Los repositorios operan bajo estricto aislamiento transaccional utilizando `TenantTxRunner`, el cual encapsula el acceso a los pools de `Read` y `Write`.

### Patrón de Repositorio:
1.  **Aislamiento Transaccional Obligatorio**: No se permite acceso directo a `*pgxpool.Pool`. Todas las operaciones deben ejecutarse dentro de los closures `r.db.Tenant.Read(ctx, func(tx pgx.Tx) error)` o `r.db.Tenant.Write(...)`.
2.  **Configuración Automática**: El runner valida el `TenantContext`, inyecta la variable de sesión `app.tenant_id` y configura el `search_path` (ej. `SET LOCAL search_path TO "schema_name", pg_catalog`) de manera segura, revirtiéndolo al hacer commit o rollback.
3.  **Auditoría y Sesión**: Las variables para triggers de auditoría (como `session_variables.uuid_persona`) se inyectan automáticamente en la misma transacción. No debe llamarse a configuradores externos o persistentes.

### Aislamiento Multi-tenant (RLS):
El motor PostgreSQL aplica aislamiento en defensa en profundidad:
*   **Roles y Permisos**: La aplicación se conecta con un rol de privilegios mínimos (`app_runtime_tenant`) que posee `NOBYPASSRLS` y no tiene permisos sobre el esquema `public`.
*   **Row Level Security (RLS)**: Cada tabla tenant-scoped (ej. `persona`, `usuario`, `sesion`) tiene habilitado `FORCE ROW LEVEL SECURITY`. Las políticas validan matemáticamente que `tenant_id::text = current_setting('app.tenant_id', true)` para operaciones `SELECT`, y exigen lo mismo con `WITH CHECK` para modificaciones.
*   **Fall-closed**: Si un query no pasa por el runner transaccional, carece de contexto local y RLS denegará silenciosa y universalmente el acceso.

## 🚀 Estrategia de Migraciones Multi-tenant (Expand & Contract)

Para garantizar zero-downtime y evitar reversas destructivas con pérdida de datos, cualquier cambio en las barreras de aislamiento debe ejecutarse en fases:

1. **Expand**: Se aplican cambios aditivos (roles, columnas, índices, RLS). Se requiere prueba canary, observabilidad y rollback de aplicación/configuración seguro que jamás restaure acceso a `public` ni desactive RLS.
2. **Contract**: Solo se ejecuta (vía un script SQL final destructivo) tras acumular evidencia técnica empírica de compatibilidad total y ausencia de consumidores antiguos (legacy pooling hooks).
3. **Rollback de DB (Forward-fix)**: La base de datos siempre avanza. Los scripts `.down.sql` de las fases expand retiran políticas y desactivan restricciones sin hacer `DROP COLUMN` que destruya datos.
