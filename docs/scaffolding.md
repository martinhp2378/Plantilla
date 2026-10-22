# ⚡ Guía Amigable: Cómo Crear y Personalizar un Nuevo Módulo CRUD

Esta guía te enseña cómo crear una pantalla completa del sistema (con tabla, buscador, ventana para crear y editar registros, botones para descargar reportes en **PDF** y **Excel**, permisos en base de datos y acceso en el menú lateral) **ejecutando un solo comando**.

Cada comando tiene un indicador claro para que sepas qué hacer:
- 🟢 **Copiar y pegar tal cual**: No tienes que cambiarle nada, solo cópialo, pégalo en tu terminal y presiona `Enter`.
- 🟡 **Modificar antes de ejecutar**: Te diremos exactamente qué palabra cambiar por el nombre de tu nuevo módulo.

---

## 🚀 Paso 1: Crear tu Nuevo Módulo

Abre la terminal dentro de la carpeta de tu proyecto y elige una de las dos opciones siguientes:

### Opción A: Creación Rápida (Usando los valores por defecto)

🟡 **Modificar antes de ejecutar**:
Cambia únicamente la palabra **`proveedor`** por el nombre de tu módulo en **singular** (si tiene varias palabras, puedes separarlas con guion medio `-` o guion bajo `_`, por ejemplo: `orden-compra`, `tipo-contrato`, `almacen`).

```bash
make new-module NAME=proveedor
```

### Opción B: Creación Personalizada (Eligiendo esquema de BD y título con acentos)

🟡 **Modificar antes de ejecutar**:
- Cambia **`orden-compra`** por el nombre técnico de tu módulo en singular.
- Cambia **`institucion`** por el esquema de PostgreSQL donde quieres guardar la tabla (`sistema` o `institucion`).
- Cambia **`"Órdenes de Compra"`** por el título exacto (con mayúsculas, espacios y acentos) que quieres que vean los usuarios en la pantalla y en los reportes.

```bash
make new-module NAME=orden-compra SCHEMA=institucion TITLE="Órdenes de Compra"
```

> [!TIP]
> **¿Qué pasa si escribes guiones o espacios en `NAME`?**
> No te preocupes: el sistema convierte automáticamente `orden-compra` en `orden_compra` para las carpetas, paquetes y tablas de base de datos, y en `OrdenCompra` para el código interno de Go y Templ.

---

## 🔄 Paso 2: Encender o Actualizar el Sistema para Ver tu Módulo

Una vez que el comando del Paso 1 termine y muestre el mensaje `🎉 Modulo creado exitosamente`, actualiza el contenedor de la aplicación para que cargue tu nuevo módulo y cree su tabla en la base de datos:

🟢 **Copiar y pegar tal cual**:
```bash
docker compose -f docker-compose.local.yml build --no-cache app && docker compose -f docker-compose.local.yml up -d app
```

Cuando termine de cargar (tarda unos 20 a 30 segundos):
1. Entra o recarga tu navegador en **[http://localhost:8080/sistema](http://localhost:8080/sistema)**.
2. Verás tu nuevo módulo en el **menú lateral izquierdo**.
3. ¡Ya puedes **crear registros**, **editarlos en su ventana modal**, **buscar/filtrar** y **descargar los reportes en PDF y Excel**!

---

## 🧹 ¿Hiciste un Módulo Solo para Probar y Quieres Borrarlo?

Si creaste un módulo de prueba (por ejemplo, llamado `pruebita` o `prueba_borrarlo`) y quieres eliminarlo por completo para no subir basura al repositorio de Git, sigue estos **2 pasos**:

### 1. Limpiar la tabla de prueba en la base de datos local

🟡 **Modificar antes de ejecutar**:
En el siguiente bloque, cambia las **3 veces** que aparece la palabra **`prueba_borrarlo`** por el nombre en `snake_case` (con guion bajo) del módulo que creaste, y verifica si el número `11` corresponde a la última migración real antes de tu prueba:

```bash
docker exec plantillainstitucional-db-1 psql -U postgres -d plantilla -c "
DELETE FROM sesion.politica WHERE recurso = 'prueba_borrarlo';
DELETE FROM sesion.recurso WHERE nombre = 'prueba_borrarlo';
DROP TABLE IF EXISTS sistema.prueba_borrarlo CASCADE;
UPDATE schema_migrations SET version = 11, dirty = false;
"
```
*(Nota: Si en tu archivo `.env` cambiaste el usuario de base de datos `PQ_WRITE_USER` por otro nombre distinto de `postgres`, reemplaza `-U postgres` por `-U TU_USUARIO`).*

### 2. Borrar la carpeta del módulo y restaurar el archivo principal

🟡 **Modificar antes de ejecutar**:
Cambia las **2 veces** que aparece **`prueba_borrarlo`** por el nombre de la carpeta de tu módulo de prueba:

```bash
rm -rf prueba_borrarlo sql/migrations/*_create_prueba_borrarlo_table.*
git checkout cmd/main.go
```

Por último, vuelve a actualizar el contenedor para que quede limpio:
🟢 **Copiar y pegar tal cual**:
```bash
docker compose -f docker-compose.local.yml up --build -d app
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
7. **Regenera las vistas y actualiza tu contenedor**:
   🟢 **Copiar y pegar tal cual**:
   ```bash
   make check
   docker compose -f docker-compose.local.yml build --no-cache app && docker compose -f docker-compose.local.yml up -d app
   ```

---

## ✅ Lista de Verificación antes de Guardar en Git

Antes de subir tu nuevo módulo al repositorio del equipo, corre este comando para verificar que no haya ningún error de formato ni de compilación:

🟢 **Copiar y pegar tal cual**:
```bash
make check
```

Si termina sin errores:
- [ ] El formulario crea registros nuevos sin error.
- [ ] El botón de editar abre la ventana modal y guarda los cambios.
- [ ] Los botones de **PDF** y **Excel** descargan los archivos correctamente y los acentos se ven bien.
