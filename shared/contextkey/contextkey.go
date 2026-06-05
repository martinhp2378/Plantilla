package contextkey

// contextKey es un tipo privado para evitar colisiones.
type contextKey string

// AuditUUIDKey es la clave para inyectar y leer el UUID del usuario para auditoría desde el contexto.
const AuditUUIDKey contextKey = "auditUUID"

// TenantSchemaKey es la clave para inyectar y leer el esquema del tenant desde el contexto.
const TenantSchemaKey contextKey = "tenantSchema"

// CSRFTokenKey es la clave para inyectar y leer el token CSRF desde el contexto.
const CSRFTokenKey contextKey = "csrfToken"
