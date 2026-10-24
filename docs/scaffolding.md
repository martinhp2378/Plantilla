# ⚡ Guía Amigable: Cómo Crear y Personalizar un Nuevo Módulo CRUD

Esta guía te enseña cómo crear una pantalla completa del sistema (con tabla, buscador, ventana para crear y editar registros, botones para descargar reportes en **PDF** y **Excel**, permisos en base de datos y acceso en el menú lateral) **ejecutando un solo comando**.

Cada comando tiene un indicador claro para que sepas qué hacer:
- 🟢 **Copiar y pegar tal cual**: No tienes que cambiarle nada, solo cópialo, pégalo en tu terminal y presiona `Enter`.
- 🟡 **Modificar antes de ejecutar**: Te diremos exactamente qué palabra cambiar por el nombre de tu nuevo módulo.

---

## 🚀 Paso 1: Crear tu Nuevo Módulo

Abre la terminal dentro de la carpeta de tu proyecto y elige cualquiera de estas opciones (funcionan igual en **Mac, Linux y Windows**):

### Opción A: Interfaz Interactiva TUI (Navegable con teclado)

🟢 **Copiar y pegar tal cual**:
```bash
jrmp
```
*(Selecciona **"⚡ Crear nuevo módulo CRUD"** con las flechas, escribe el nombre, elige el esquema `sistema` o `institucion` con `←`/`→` y presiona `Enter`. También puedes usar `jrmp new-module` en modo asistente clásico).*

### Opción B: Creación Rápida en 1 línea (Usando los valores por defecto)

🟡 **Modificar antes de ejecutar**:
Cambia únicamente la palabra **`proveedor`** por el nombre de tu módulo en **singular** (si tiene varias palabras, puedes separarlas con guion medio `-` o guion bajo `_`, por ejemplo: `orden-compra`, `tipo-contrato`, `almacen`).

```bash
jrmp new-module proveedor
```

### Opción C: Creación Personalizada (Eligiendo esquema de BD y título con acentos)

🟡 **Modificar antes de ejecutar**:
- Cambia **`orden-compra`** por el nombre técnico de tu módulo en singular.
- Cambia **`institucion`** por el esquema de PostgreSQL donde quieres guardar la tabla (`sistema` o `institucion`).
- Cambia **`"Órdenes de Compra"`** por el título exacto (con mayúsculas, espacios y acentos) que quieres que vean los usuarios en la pantalla y en los reportes.

```bash
jrmp new-module orden-compra -schema institucion -title "Órdenes de Compra"
```

> [!TIP]
> **¿Qué pasa si escribes guiones o espacios en el nombre?**
> No te preocupes: `jrmp` convierte automáticamente `orden-compra` en `orden_compra` para las carpetas, paquetes y tablas de base de datos, y en `OrdenCompra` para el código interno de Go y Templ.

---

## 🔄 Paso 2: Ver tu Nuevo Módulo en el Navegador

Si ya tienes encendido el sistema con `jrmp up`, gracias a **Air (recarga en vivo)** el contenedor detectará el nuevo módulo, recompilará y ejecutará la migración automáticamente en unos segundos (si estaba apagado, enciéndelo con `jrmp up`).

1. Entra o recarga tu navegador en **[http://localhost:8080/sistema](http://localhost:8080/sistema)**.
2. Verás tu nuevo módulo en el **menú lateral izquierdo**.
3. ¡Ya puedes **crear registros**, **editarlos en su ventana modal**, **buscar/filtrar** y **descargar los reportes en PDF y Excel**!

---

## 🧹 ¿Hiciste un Módulo Solo para Probar y Quieres Borrarlo?

Si creaste un módulo de prueba (por ejemplo, llamado `pruebita` o `prueba_borrarlo`) y quieres eliminarlo por completo para no subir basura al repositorio de Git, ejecuta este único comando (que borra la carpeta del módulo, sus migraciones, su registro en `cmd/main.go` y, si tu base de datos local está encendida, revierte también la tabla, las políticas ABAC y `schema_migrations` en PostgreSQL):

🟡 **Modificar antes de ejecutar**:
Cambia **`prueba_borrarlo`** por el nombre de tu módulo de prueba:

```bash
jrmp remove-module prueba_borrarlo
```

---

## 🗂️ ¿Qué Archivos Creó el Generador por Ti?

Por defecto, tu módulo nuevo viene con **4 campos listos para usar**: `Nombre`, `Código`, `Descripción` y `Activo` (Estatus). Todo se encuentra dentro de una carpeta con el nombre de tu módulo:

| Archivo | Qué hace (en palabras sencillas) | ¿Cuándo necesitas abrirlo? |
| :--- | :--- | :--- |
| `sql/migrations/..._create_<modulo>_table.up.sql` | Crea la tabla en PostgreSQL y da permisos para entrar al módulo. | Cuando quieras agregar más columnas a la tabla de base de datos. |
| `<modulo>/types.go` | Define qué datos tiene un registro de tu módulo. | Cuando quieras agregar o cambiar campos (ej. `Telefono`, `FechaNacimiento`). |
| `<modulo>/repository/queries.go` | Contiene las consultas SQL (`SELECT`, `INSERT`, `UPDATE`). | Cuando agregues columnas nuevas y quieras guardarlas o leerlas en la BD. |
| `<modulo>/repository/repository.go` | Conecta las consultas SQL con los campos de Go. | Cuando agregues columnas nuevas al leer o guardar. |
| `<modulo>/service/reportes.go` | Configura las columnas que salen en el **PDF** y en el **Excel**. | Cuando quieras cambiar los títulos o campos que se descargan en los reportes. |
| `<modulo>/handler/types.go` | Lee los datos que el usuario escribió en el formulario web (`parseForm`). | Cuando agregues un campo nuevo al formulario. |
| `<modulo>/view/components/form_v2.templ` | Dibuja la ventana emergente (modal) de **Nuevo / Editar**. | Cuando quieras agregar o quitar cajas de texto en el formulario. |
| `<modulo>/view/components/table_v2.templ` | Dibuja las columnas y filas de la **tabla principal**. | Cuando quieras mostrar más o menos columnas en la tabla. |
| `<modulo>/handler/register.go` | Configura el texto y el ícono que aparecen en el **menú lateral**. | Cuando quieras cambiar el ícono o nombre del menú. |

---

## 🛠️ Cómo Agregar un Campo Nuevo a tu Módulo (Paso a Paso)

Supongamos que creaste el módulo `proveedor` y, además de `Nombre`, `Código` y `Descripción`, quieres agregar un campo llamado **`Telefono`**. Solo sigue este orden:

1. **En la migración SQL (`sql/migrations/..._create_proveedor_table.up.sql`)**:
   Agrega la columna `telefono VARCHAR(20) NOT NULL DEFAULT ''` debajo de `descripcion`.
2. **En `<modulo>/types.go`**:
   Agrega `Telefono string` dentro de `type Proveedor struct`.
3. **En `<modulo>/repository/queries.go` y `<modulo>/repository/repository.go`**:
   Agrega `telefono` en el `SELECT`, `INSERT` y `UPDATE`, y pásalo en `Scan(...)` y `Exec(...)`.
4. **En `<modulo>/handler/types.go` (función `parseForm`)**:
   Agrega `Telefono: r.FormValue("telefono"),`.
5. **En las vistas (`form_v2.templ` y `table_v2.templ`)**:
   Copia el bloque del input de `codigo` en `form_v2.templ` y cámbialo para `telefono`, y agrega la columna en `table_v2.templ`.
6. **En los reportes (`<modulo>/service/reportes.go`)**:
   Agrega `"Teléfono"` en `ReportHeaders` y `"Telefono"` en `ReportFields`.
7. **Regenera las vistas y valida tu módulo**:
   🟢 **Copiar y pegar tal cual**:
   ```bash
   jrmp check
   ```

---

## ✅ Lista de Verificación antes de Guardar en Git

Antes de subir tu nuevo módulo al repositorio del equipo, corre este comando para verificar que no haya ningún error de formato ni de compilación:

🟢 **Copiar y pegar tal cual**:
```bash
jrmp check
```

Si termina sin errores:
- [ ] El formulario crea registros nuevos sin error.
- [ ] El botón de editar abre la ventana modal y guarda los cambios.
- [ ] Los botones de **PDF** y **Excel** descargan los archivos correctamente y los acentos se ven bien.
