# 🗄️ Guía Amigable de Base de Datos y Migraciones (PostgreSQL)

Esta guía explica cómo está organizada la base de datos en **PostgreSQL**, cómo conectarte a ella desde tu computadora para ver tus tablas y datos, cómo crear migraciones nuevas y cómo escribir consultas en Go de forma segura.

---

## 🗂️ 1. Cómo están Organizados los Esquemas

En lugar de meter todas las tablas revueltas en un solo lugar, PostgreSQL las organiza en **esquemas** (como si fueran carpetas dentro de la base de datos):

| Esquema | Qué tablas se guardan aquí | Ejemplos |
| :--- | :--- | :--- |
| **`sistema`** | Configuración general, catálogos del sistema, bitácora de auditoría y retención de reportes. | `sistema.cfg_parametros`, `sistema.bitacora_cambios`, `sistema.reportes_generados` |
| **`institucion`** | Tablas de negocio de la institución e identidad de personas. | `institucion.persona`, `institucion.cat_tenants`, `institucion.crud` |
| **`sesion`** | Usuarios, contraseñas, sesiones activas y permisos por perfil (ABAC). | `sesion.usuario`, `sesion.sesion`, `sesion.recurso`, `sesion.politica` |

---

## 🔍 2. Comandos Listos para Copiar y Pegar: Cómo Ver tu Base de Datos en Docker

Si quieres revisar qué tablas existen o hacer una consulta SQL rápida sin instalar programas externos:

### A) Entrar a la consola interactiva de PostgreSQL
🟢 **Copiar y pegar tal cual**:
```bash
docker exec -it plantillainstitucional-db-1 psql -U postgres -d plantilla
```
*(Nota: Si en tu archivo `.env` cambiaste `PQ_WRITE_USER` por otro usuario, cambia `postgres` por tu usuario. Para salir de la consola de PostgreSQL en cualquier momento, escribe `\q` y presiona `Enter`).*

### B) Comandos útiles dentro de la consola de PostgreSQL

| ¿Qué quieres ver? | Comando (🟢 Copiar y pegar dentro de `psql`) |
| :--- | :--- |
| Ver todas las tablas del esquema `sistema` | `\dt sistema.*` |
| Ver todas las tablas del esquema `institucion` | `\dt institucion.*` |
| Ver las columnas de una tabla específica | `\d sistema.cfg_parametros` |
| Ver qué número de migración tiene aplicada la BD | `SELECT * FROM schema_migrations;` |
| Salir de PostgreSQL y volver a tu terminal normal | `\q` |

---

## 📦 3. Qué son las Migraciones y Cómo Funcionan

Las **migraciones** son archivos `.sql` numerados que viven en la carpeta `sql/migrations/`. Sirven para que cuando un compañero descargue tus cambios y encienda su Docker, **sus tablas se creen o actualicen automáticamente** igual que en tu computadora.

- Cada migración tiene **2 archivos**:
  1. `000011_nombre.up.sql`: Instrucciones para **crear** la tabla o agregar columnas.
  2. `000011_nombre.down.sql`: Instrucciones para **deshacer** ese cambio si fuera necesario.

> [!TIP]
> **Cuando usas `jrmp new-module proveedor`**, el generador crea automáticamente ambos archivos `.up.sql` y `.down.sql` con el número siguiente que toca. ¡No tienes que crearlos a mano!

### ¿Cómo crear una migración manual si solo quieres agregar una columna a una tabla existente?

Si no estás creando un módulo nuevo, sino modificando una tabla que ya existe, puedes crear el par de archivos vacíos con este comando:

🟡 **Modificar antes de ejecutar**:
Cambia **`agregar_telefono_a_proveedor`** por una descripción corta separada por guiones bajos:
```bash
migrate create -ext sql -dir sql/migrations -seq agregar_telefono_a_proveedor
```

---

## 🏗️ 4. Cómo Hacer Consultas SQL en Go (`TenantTxRunner`)

Por seguridad institucional, todas las tablas cuentan con **Row Level Security (RLS)** para que los datos de una dependencia o tenant jamás se mezclen con los de otra.

Por eso, en los archivos dentro de `repository/`, las consultas siempre se envuelven dentro de **`r.db.Tenant.Read`** (para leer datos con `SELECT`) o **`r.db.Tenant.Write`** (para guardar o borrar con `INSERT`, `UPDATE`, `DELETE`).

### Ejemplo para Leer Datos (`Read`) — Listo para copiar y adaptar:
```go
err := r.db.Tenant.Read(ctx, func(tx pgx.Tx) error {
	return tx.QueryRow(ctx, `SELECT id, nombre, codigo FROM sistema.proveedor WHERE id = $1`, id).
		Scan(&item.ID, &item.Nombre, &item.Codigo)
})
```

### Ejemplo para Guardar o Actualizar Datos (`Write`) — Listo para copiar y adaptar:
```go
err := r.db.Tenant.Write(ctx, func(tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `UPDATE sistema.proveedor SET nombre = $1 WHERE id = $2`, nuevoNombre, id)
	return err
})
```
