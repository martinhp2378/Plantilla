# 🔐 Autenticación, Sesiones y Permisos (ABAC)

Esta guía explica cómo funciona el inicio de sesión en el sistema, cómo se protegen los formularios y los reportes privados, y cómo dar permisos a distintos perfiles de usuario sobre tus módulos.

---

## 🔑 1. Usuario de Prueba para Desarrollo Local

Cuando levantas el proyecto con Docker por primera vez, la base de datos ya incluye un usuario con acceso total (`superadmin`) para que puedas probar todas las pantallas de inmediato:

| Dato | Valor |
| :--- | :--- |
| **Pantalla de acceso** | [http://localhost:8080/login](http://localhost:8080/login) |
| **Correo electrónico** | `admin@issemym.gob.mx` |
| **Contraseña** | `admin123` |
| **Perfil asignado** | `superadmin` |

---

## 🎟️ 2. Cómo Funcionan las Sesiones

1. **Cookies Seguras**: Al iniciar sesión en la web, el sistema crea una sesión en PostgreSQL (o Redis si se activa en producción) y entrega al navegador una cookie protegida (`HttpOnly`, `SameSite=Strict`).
2. **Aislamiento de Sesiones**: El sistema utiliza `github.com/issemym-repos/sesiones` (`v0.4.5+`), que garantiza que ningún cliente sin cookie válida pueda recuperar la sesión de otro navegador.
3. **Rotación de Cookies en Despliegues de Seguridad**:
   - En el archivo `.env`, la variable `COOKIE_NOMBRE=plantilla_v2` define el nombre de la cookie de sesión.
   - Si en producción se requiere forzar que absolutamente todos los usuarios vuelvan a iniciar sesión tras una actualización crítica, basta con cambiar `COOKIE_NOMBRE` a un identificador nuevo (por ejemplo `mi_sistema_v2`) y limpiar la tabla `sesion.sesion`.

---

## 📜 3. Permisos por Perfil (ABAC)

No basta con que un usuario haya iniciado sesión; el sistema verifica en cada clic si su **Perfil** tiene permiso para entrar a ese **Recurso** (módulo) y realizar esa **Acción** (`consultar`, `crear`, `modificar`, `eliminar`).

### Perfiles que ya vienen en la Plantilla

| Perfil | Qué puede hacer por defecto |
| :--- | :--- |
| **`superadmin`** | Acceso total a todos los módulos del sistema y a los módulos nuevos creados con `make new-module`. |
| **`admin`** | Acceso de consulta, creación y modificación en los módulos generados con `make new-module`. |
| **`operador`** | Acceso operativo a módulos específicos que se le asignen en la tabla `sesion.politica`. |
| **`anonimo`** | Usuarios que aún no han iniciado sesión (solo pueden ver `/login` y recuperación de contraseña). |

### ¿Cómo dar permiso a otro perfil (ej. `operador`) sobre un módulo nuevo?

Cuando creas un módulo con `make new-module`, la migración SQL le da permiso automáticamente a `superadmin` y `admin`. Si también quieres darle permiso al perfil `operador` (o a un perfil nuevo), solo copia y pega estas líneas al final del archivo `.up.sql` de tu módulo:

🟡 **Modificar antes de guardar**:
Cambia **`proveedor`** por el nombre de tu módulo y **`operador`** por el perfil al que quieres darle acceso:

```sql
-- Dar permiso de consulta, creación y edición al perfil 'operador'
INSERT INTO sesion.politica (perfil, recurso, accion, razon_i) VALUES
    ('operador', 'proveedor', 'consultar', 'Permiso operativo de consulta'),
    ('operador', 'proveedor', 'crear', 'Permiso operativo de registro'),
    ('operador', 'proveedor', 'modificar', 'Permiso operativo de edicion')
ON CONFLICT (perfil, recurso, accion) DO NOTHING;
```

---

## 🛡️ 4. Protecciones Automáticas que ya trabajan por ti

No tienes que programar nada extra para contar con estas defensas:

1. **Protección CSRF (`nosurf`)**:
   - Todas las peticiones que guardan o borran datos (`POST`, `PUT`, `DELETE`) están protegidas contra ataques de falsificación de peticiones.
   - Si usas HTMX (`hx-post`, `hx-put`, `hx-delete`), el layout base ya envía el token CSRF automáticamente en cada clic.
2. **Descarga Protegida de Reportes (`/private-reports/*`)**:
   - Los reportes PDF y Excel generados nunca se guardan en carpetas públicas.
   - Solo un usuario con sesión activa y perteneciente a la institución (tenant) propietaria del reporte puede descargarlo.
3. **Bitácora de Auditoría Automática (`sistema.bitacora_cambios`)**:
   - Cada vez que un usuario crea, edita o elimina un registro, el sistema inyecta automáticamente su identificador (`uuid_persona`) en la transacción de PostgreSQL y un *trigger* guarda el historial del cambio sin que tengas que escribir código adicional.
