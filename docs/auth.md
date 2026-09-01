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

## 🛡️ Capa de Seguridad (Hardenización)

Además de la autenticación, el sistema implementa medidas proactivas contra ataques comunes de la web.

### Activación y recuperación de contraseña

Las cuentas nuevas permanecen inactivas hasta que su titular consume un enlace de activación de un solo uso. Los tokens se generan con entropía criptográfica, se almacenan exclusivamente como hash, expiran en 24 horas y se eliminan al consumirse. La recuperación responde de forma no enumerante y limita solicitudes por dirección IP.

Los reportes de personas y parámetros se almacenan en el repositorio privado. En producción, esto entrega URLs firmadas de duración limitada. El despliegue local expone la ruta privada mediante el servidor estático de desarrollo; antes de exponer un entorno compartido, debe instalarse una capa de autorización que valide la sesión y la política antes de entregar dichos archivos.

### Protección CSRF (Cross-Site Request Forgery)
Gestionada mediante la librería `nosurf` en `app/middleware/security.go`.
*   **Funcionamiento**: Todas las peticiones mutables (POST, PUT, DELETE) requieren un token válido.
*   **Integración HTMX**: El token se inyecta automáticamente en los headers de cada petición AJAX gracias a la configuración en `base.templ`.
*   **Uso en Formularios**: En formularios tradicionales, se debe incluir un campo oculto:
    ```templ
    <input type="hidden" name="csrf_token" value={ utils.GetCSRFToken(ctx) } />
    ```

### Cabeceras de Seguridad (OWASP)
El `SecurityHeadersMiddleware` inyecta automáticamente cabeceras recomendadas:
*   `X-Frame-Options: DENY`: Previene ataques de Clickjacking.
*   `X-Content-Type-Options: nosniff`: Evita que el navegador adivine el tipo MIME.
*   `Strict-Transport-Security (HSTS)`: Fuerza el uso de HTTPS.

## 📜 Autorización (Middleware de Políticas)

No basta con estar autenticado; el usuario debe tener permiso para acceder al recurso específico.

### `PolicyMiddleware`
Ubicado en `app/middleware/policy.go`.
*   **Recurso**: Identificado por el path o prefijo.
*   **Acción**: Determina qué puede hacer el usuario (consultar, crear, modificar, eliminar).
*   **Reglas**: Las políticas se guardan en la base de datos (tabla `politica`) y vinculan Perfiles con Recursos y Acciones.

## 🕵️ Auditoría y Contexto

Para que la base de datos (mediante triggers) pueda registrar quién realiza una acción de forma segura en entornos multi-tenant:

1.  **Middleware**: `AuditContextMiddleware` extrae el UUID del usuario autenticado de la sesión.
2.  **Contexto**: El UUID se inyecta en el `context.Context` utilizando `contextkey.AuditUUIDKey`.
3.  **Inyección en DB**: Al iniciar una operación de base de datos, el `TenantTxRunner` lee este UUID del contexto y ejecuta `SET LOCAL session_variables.uuid_persona`. De esta forma, los triggers de auditoría (por ejemplo en `sistema.bitacora_cambios`) recuperan el autor exacto sin acoplar el código de los repositorios.

---

## 👤 Perfiles Predeterminados

La plantilla incluye perfiles base listos para usar:
*   `superadmin`: Acceso total a configuración y gestión.
*   `operador`: Acceso a módulos de negocio (Personas).
*   `anonimo`: Para usuarios no autenticados con acceso limitado.
