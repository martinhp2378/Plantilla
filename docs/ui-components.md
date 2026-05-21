# 🎨 Frontend y Componentes

La interfaz de usuario se construye utilizando **Templ**, **HTMX** y **Tailwind CSS**. Este enfoque permite crear aplicaciones interactivas con la simplicidad de renderizado del lado del servidor.

## 🧱 Templ: Componentes con Seguridad de Tipos

Templ nos permite escribir HTML con lógica de Go. Los componentes se definen en archivos `.templ`.

### Ejemplo de Componente:
```templ
package components

templ Saludo(nombre string) {
    <div class="p-4 bg-gray-100 rounded-lg">
        <h1 class="text-xl font-bold">Hola, { nombre }!</h1>
    </div>
}
```

## ⚡ HTMX: Interactividad sin JavaScript Complejo

HTMX nos permite hacer peticiones AJAX directamente desde atributos HTML.

### Patrones Comunes:
*   **Carga Parcial**: Usar `hx-get="/ruta/parcial"` y `hx-target="#contenedor"` para actualizar solo una parte de la página.
*   **Filtros en Tiempo Real**: Usar `hx-trigger="keyup changed delay:500ms"` en inputs de búsqueda.
*   **Modales**: Cargar el contenido del modal bajo demanda mediante HTMX.

## 📑 Layouts y Vistas v2

La plantilla utiliza una estructura de layouts en `shared/views/layouts/`:
*   `BaseConNav`: Para páginas que requieren sidebar y header (dashboard).
*   `BaseSinNav`: Para páginas públicas (login, errores).

### Estructura de una Página:
1.  **Handler**: Define los datos necesarios.
2.  **Layout**: Envuelve la página con los assets (CSS/JS).
3.  **Componentes**: Piezas reutilizables (Tablas, Formularios, Modales).

## 🛠️ Tailwind CSS e Iconos

*   **Tailwind**: Se usa para todo el estilizado. Las clases se escriben directamente en los archivos `.templ`.
*   **Material Icons**: Se incluyen por defecto para la iconografía mediante la fuente oficial de Google.

```templ
<span class="material-icons">settings</span>
```
