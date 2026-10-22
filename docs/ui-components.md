# 🎨 Guía Amigable de Pantallas y Componentes (Frontend v2)

En esta plantilla no necesitas aprender React, Angular ni escribir archivos complejos de JavaScript para hacer pantallas dinámicas. Todo el diseño visual ya está armado en **Componentes Institucionales v2** listos para copiar y pegar dentro de tus archivos `.templ`.

> [!IMPORTANT]
> **Regla de Oro de las Vistas (`.templ`)**:
> Cada vez que modifiques o agregues algo en un archivo que termine en **`.templ`**, ejecuta este comando en tu terminal para que Go actualice los cambios:
>
> 🟢 **Copiar y pegar tal cual**:
> ```bash
> templ generate
> ```
> *(O si trabajas directo con Docker, reconstruye el contenedor con `docker compose -f docker-compose.local.yml up --build -d app`).*

---

## 🧱 1. Cajas de Texto y Formularios (`shared/views/components/v2/forms.templ`)

Cuando abras el archivo `view/components/form_v2.templ` de tu módulo, puedes copiar y pegar cualquiera de estos bloques para agregar más campos a tu ventana de **Nuevo / Editar**:

### A) Caja de texto normal, número, correo o fecha (`InputFieldV2`)

🟡 **Modificar antes de guardar**:
Cambia `Label` (el título que ve el usuario), `Name` (el nombre técnico del campo), `Type` (`"text"`, `"number"`, `"email"`, `"date"`) y `Value` (el campo de tu struct en Go).

```templ
@v2.InputFieldV2(v2.InputPropsV2{
	Label:       "Teléfono de Contacto",
	Name:        "telefono",
	Type:        "text",
	Placeholder: "Ej. 722 123 4567",
	Value:       item.Telefono,
	Required:    true,
	Errors:      errors,
})
```

### B) Lista desplegable de opciones (`SelectFieldV2`)

🟡 **Modificar antes de guardar**:
Cambia `Label`, `Name`, `SelectedValue` y la lista de `Options`.

```templ
@v2.SelectFieldV2(v2.SelectProps{
	Label:         "Tipo de Proveedor",
	Name:          "tipo_proveedor",
	SelectedValue: item.TipoProveedor,
	Required:      true,
	Errors:        errors,
	Options: []components.SelectOption{
		{Value: "BIENES", Label: "Proveedor de Bienes"},
		{Value: "SERVICIOS", Label: "Proveedor de Servicios"},
	},
})
```

### C) Área de texto grande para observaciones o descripciones (`TextAreaV2`)

🟡 **Modificar antes de guardar**:
Cambia `Label`, `Name` y `Value`.

```templ
@v2.TextAreaV2(v2.TextAreaPropsV2{
	Label:       "Observaciones",
	Name:        "observaciones",
	Placeholder: "Escribe aquí los detalles adicionales...",
	Value:       item.Observaciones,
	Rows:        3,
	Errors:      errors,
})
```

### D) Casilla de verificación Activo / Inactivo (`CheckboxV2`)

🟡 **Modificar antes de guardar**:
Cambia `Label`, `Description`, `Name` y `Checked`.

```templ
@v2.CheckboxV2(v2.CheckboxPropsV2{
	Label:       "Registro Activo",
	Description: "Habilitar este registro para su uso en el sistema",
	Name:        "activo",
	Value:       "true",
	Checked:     item.IsActive,
})
```

---

## 📊 2. Tablas Dinámicas (`shared/views/components/v2/table.templ`)

Cuando creas un módulo con `make new-module`, la tabla principal ya viene configurada en `view/components/table_v2.templ` con:
- **Buscador automático**: Filtra mientras el usuario escribe sin recargar toda la pantalla.
- **Botones de exportación**: Descarga directa en **PDF** y **Excel**.
- **Paginación**: Botones de página anterior/siguiente y selector de cantidad de registros.
- **Acciones por fila (`DataRowV2`)**: Botón para abrir el modal de edición en `#modal` y botón para eliminar con confirmación.

### Cómo agregar una columna nueva a tu tabla

1. En `view/layouts/<tu_modulo>_v2.templ`, busca la lista `Columns:` y agrega el título de tu nueva columna:
   ```go
   Columns: []v2.TableColumn{
       {Label: "Nombre", Key: "nombre", Sortable: true},
       {Label: "Código", Key: "codigo", Sortable: true},
       {Label: "Teléfono", Key: "telefono"}, // <-- Nueva columna
       {Label: "Descripción", Key: "descripcion"},
       {Label: "Estatus", Key: "is_active", Sortable: true},
   },
   ```
2. En `view/components/table_v2.templ`, agrega una celda `<td class="px-8 py-5">` dentro del ciclo `for _, item := range res.Items`:
   ```templ
   <td class="px-8 py-5">
       <span class="text-xs font-medium text-gray-600">{ item.Telefono }</span>
   </td>
   ```

---

## 🪟 3. Ventanas Emergentes / Modales (`ModalV2`)

Todos los modales del sistema se cargan dentro del contenedor `<div id="modal"></div>` que ya existe en el layout principal.
- Para que un botón o fila abra un modal vía HTMX, siempre debe apuntar a `#modal` (`ModalTarget: "#modal"` o `hx-target="#modal"`).
- Al guardar con éxito un formulario, el servidor devuelve el componente `compartidos.CloseModal()` junto con la notificación emergente (*toast*) para cerrar la ventana automáticamente y actualizar la tabla de fondo.

---

## 🎨 4. Íconos Disponibles (Material Icons)

El sistema incluye la librería oficial **Google Material Icons**. Para poner cualquier ícono en un botón, menú o tarjeta, solo usa su nombre en texto:

🟢 **Copiar y pegar** (cambiando la palabra `folder` por el ícono que busques en [fonts.google.com/icons](https://fonts.google.com/icons)):
```templ
<span class="material-icons">folder</span>
```

Ejemplos de íconos comunes:
- `person` (Persona / Usuario)
- `business` (Institución / Proveedor)
- `inventory_2` (Inventario / Artículos)
- `description` (Documento / Reporte)
- `settings` (Configuración)
- `edit` (Editar)
- `delete` (Eliminar)
