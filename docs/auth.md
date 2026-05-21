# 🔐 Autenticación y Autorización

La seguridad es una prioridad en la Plantilla Institucional, implementando un sistema de doble capa (Autenticación + Políticas).

## 🎟️ Autenticación (Sesiones y JWT)

### Sesiones Tradicionales
Para la interfaz web, usamos sesiones seguras gestionadas por la librería `sesiones`.
*   **Almacenamiento**: PostgreSQL (por defecto) o Redis.
*   **Middleware**: `a.ConstructorSesion.Middleware` inyecta la sesión en cada petición.
*   **Seguridad**: Cookies con flags `HttpOnly`, `Secure` y `SameSite`.

### JWT (JSON Web Tokens)
Soportado para APIs o comunicaciones stateless.
*   Configurado mediante `ConstructorSesionJWT`.
*   Permite autenticación mediante el header `Authorization: Bearer <token>`.

## 📜 Autorización (Middleware de Políticas)

No basta con estar autenticado; el usuario debe tener permiso para acceder al recurso específico.

### `PolicyMiddleware`
Ubicado en `app/middleware/policy.go`.
*   **Recurso**: Identificado por el path o prefijo.
*   **Acción**: Determina qué puede hacer el usuario (consultar, crear, modificar, eliminar).
*   **Reglas**: Las políticas se guardan en la base de datos (tabla `politica`) y vinculan Perfiles con Recursos y Acciones.

## 🕵️ Auditoría y Contexto

Para que los repositorios puedan registrar quién realiza una acción sin acoplarse al sistema de sesiones:

1.  **Middleware**: `AuditContextMiddleware` extrae el UUID del usuario de la sesión.
2.  **Contexto**: El UUID se inyecta en el `context.Context` bajo la llave `db.UserUUIDKey`.
3.  **Repositorio**: Recupera el UUID para llenar campos como `usuario_creacion` o `razon_i`.

---

## 👤 Perfiles Predeterminados

La plantilla incluye perfiles base listos para usar:
*   `superadmin`: Acceso total a configuración y gestión.
*   `operador`: Acceso a módulos de negocio (Personas).
*   `anonimo`: Para usuarios no autenticados con acceso limitado.
