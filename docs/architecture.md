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

## 🔗 Gestión de Rutas (`GlobalRutas`)

Para asegurar la consistencia entre el ruteador del servidor (Chi) y las URLs en las vistas (HTMX/Templ), el proyecto utiliza un sistema centralizado en `shared/pagina/rutas.go`.

*   **Prefix**: Ruta base sin slash final (ej. `/configuracion/usuarios`), ideal para `r.Route()`.
*   **Base**: Ruta base con slash final (ej. `/configuracion/usuarios/`), ideal para enlaces de navegación.
*   **Helpers**: Métodos como `GetEditar(id)` o `GetBorrar(id)` generan automáticamente URLs consistentes.

## 💉 Inyección de Dependencias (DI)

La DI se realiza de forma manual en la carpeta `cmd/config/`. Esto proporciona control total sobre el ciclo de vida de los objetos.
*   `initrepository.go`: Instancia los repositorios.
*   `initservices.go`: Instancia los servicios pasando los repositorios requeridos.
*   `inithandlers.go`: Instancia los handlers pasando los servicios requeridos.

## 🏢 Multi-tenancy y Aislamiento de Datos

La plantilla implementa un modelo de **Aislamiento en Defensa en Profundidad**:
1.  **Detección (Edge)**: El `TenantMiddleware` identifica y valida el tenant basado en el subdominio (o header autorizado si es un proxy de confianza) e inyecta el `TenantContext` inmutable en la petición.
2.  **Contexto Transaccional**: El acceso a base de datos se realiza exclusivamente a través de `TenantTxRunner`. Este inyecta de forma segura el esquema (vía `SET LOCAL search_path`) y la variable de sesión `app.tenant_id`. Todo se revierte al hacer commit o rollback.
3.  **Aislamiento RLS (Row Level Security)**: Cada tabla multi-tenant posee políticas `FORCE ROW LEVEL SECURITY`. Si un query bypassa el runner o no provee la identidad exacta, el motor de base de datos deniega el acceso automáticamente (fail-closed).
