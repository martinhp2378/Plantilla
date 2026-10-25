# 🌱 Guía Paso a Paso: De Cero a tu Propio Proyecto

Esta guía está diseñada para acompañarte desde que descargas la **Plantilla Institucional** por primera vez hasta que tienes tu propio sistema funcionando en tu computadora, renombrado con los datos de tu proyecto. **No necesitas experiencia previa en Go, Git ni Docker para seguirla.**

Para que sea muy fácil de seguir, cada comando tiene una etiqueta:
- 🟢 **Copiar y pegar tal cual**: No tienes que cambiarle nada, solo cópialo, pégalo en tu terminal y presiona `Enter`.
- 🟡 **Modificar antes de ejecutar**: Te indicaremos exactamente qué palabra debes cambiar por los datos de tu proyecto.

---

## 🛠️ Fase 0: Programas que necesitas instalar (solo la primera vez)

Antes de empezar, asegúrate de tener instalados estos programas en tu computadora:

| Programa | Para qué sirve | Enlace o Comando de instalación |
| :--- | :--- | :--- |
| **1. Git** | Descargar el código y guardar tu historial de cambios. | [Descargar Git](https://git-scm.com/downloads) |
| **2. Docker Desktop** | Levantar la base de datos PostgreSQL y la aplicación sin instalar nada extra. | [Descargar Docker Desktop](https://www.docker.com/products/docker-desktop/) |
| **3. Go (1.25+)** | El lenguaje de programación del sistema (necesario para el CLI y el generador). | [Descargar Go](https://go.dev/dl/) |
| **4. Visual Studio Code** | El editor de código recomendado para ver y modificar tus archivos. | [Descargar VS Code](https://code.visualstudio.com/) |
| **5. Templ CLI** | Compila las vistas HTML (`.templ`) en código Go cuando creas o editas pantallas. | Ver comando abajo 👇 |
| **6. Air** | Recarga la aplicación en vivo automáticamente cada vez que guardas un cambio en tu código. | Ver comando abajo 👇 |
| **7. JRMP CLI (`jrmp`)** | Herramienta oficial para crear proyectos desde cero, inicializar, generar/limpiar módulos y validar. | Ver comando abajo 👇 |

### Cómo instalar Templ CLI, Air y JRMP CLI (una vez que ya instalaste Go)

Abre tu terminal y ejecuta estos comandos para instalar las tres herramientas:

🟢 **Copiar y pegar tal cual**:
```bash
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/air-verse/air@latest
go install github.com/issemym-repos/PlantillaInstitucional/cmd/jrmp@latest
```

Si usas **Mac (macOS)** o **Linux**, asegúrate de que tu terminal encuentre los comandos `templ`, `air` y `jrmp` ejecutando también esta línea (solo necesitas hacerlo una vez):

🟢 **Copiar y pegar tal cual**:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
grep -qxF 'export PATH="$PATH:$(go env GOPATH)/bin"' ~/.zshrc 2>/dev/null || echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
```

Para comprobar que quedaron bien instalados, escribe:
🟢 **Copiar y pegar tal cual**:
```bash
templ version
air -v
jrmp version
```

> [!IMPORTANT]
> **Antes de continuar**, abre la aplicación **Docker Desktop** en tu computadora y espera a que el ícono de la ballena indique que está en ejecución (*Engine running*).

---

## ⚡ Vía Rápida: Panel Interactivo TUI (`jrmp`) o Asistente `jrmp create`

Si ya instalaste `jrmp` en el paso anterior, **puedes hacer las Fases 1, 2, 3 y 4 automáticamente desde el panel visual interactivo** (navegable con flechas `↑`/`↓` y `Enter`) o con un solo comando directo:

1. Entra a la carpeta donde guardas tus proyectos (por ejemplo, el Escritorio):
   🟢 **Copiar y pegar tal cual**:
   ```bash
   cd ~/Desktop
   ```
2. Abre la interfaz interactiva TUI (o ejecuta `jrmp create` directo):
   🟢 **Copiar y pegar tal cual**:
   ```bash
   jrmp
   ```
   *(En el menú selecciona **"🌱 Crear proyecto desde cero"**, llena los campos del formulario interactivo y presiona `Enter`. También puedes ejecutar directamente `jrmp create` o `jrmp create MiNuevoProyecto`).*

> 💡 **¿Prefieres clonar el repositorio manualmente paso a paso con Git?** Sigue leyendo la **Fase 1** aquí abajo.

---

## 📥 Fase 1 (Alternativa Manual): Descargar (Clonar) la Plantilla

Vamos a descargar una copia de la plantilla dentro de una carpeta nueva que llevará el nombre de tu proyecto.

### Paso 1.1: Abre tu Terminal
- **En Mac**: Presiona `Cmd + Espacio`, escribe `Terminal` y presiona `Enter`.
- **En Windows**: Abre **Git Bash** o **PowerShell**.

### Paso 1.2: Entra a la carpeta donde guardas tus proyectos
🟢 **Copiar y pegar tal cual** (esto te llevará a tu Escritorio, o puedes usar tu carpeta de preferencia):
```bash
cd ~/Desktop
```

### Paso 1.3: Descarga la plantilla con el nombre de tu nuevo proyecto
🟡 **Modificar antes de ejecutar**:
Cambia la palabra **`MiNuevoProyecto`** (aparece al final del primer comando y en el segundo comando) por el nombre real de tu sistema, **sin espacios ni acentos** (por ejemplo: `SistemaInventario`, `ControlVehicular`, `CatalogoProveedores`).

```bash
git clone https://github.com/issemym-repos/PlantillaInstitucional.git MiNuevoProyecto
cd MiNuevoProyecto
```
*(Nota: Si aún no habías instalado `jrmp`, puedes instalarlo directamente desde aquí ejecutando `go install ./cmd/jrmp`).*

---

## 🔗 Fase 2: Desvincular de la Plantilla Original en Git

Como descargaste el código desde la plantilla institucional, Git todavía recuerda la dirección de esa plantilla. Si no la desvinculas, podrías subir tus cambios por error a la plantilla base en lugar de al repositorio de tu nuevo proyecto.

### Paso 2.1: Borrar el vínculo con la plantilla original
🟢 **Copiar y pegar tal cual**:
```bash
git remote remove origin
```

### Paso 2.2: Verificar que quedó desvinculado
🟢 **Copiar y pegar tal cual**:
```bash
git remote -v
```
*(Si no muestra ningún texto al presionar `Enter`, ¡excelente! Ya no hay riesgo de tocar la plantilla original).*

### Paso 2.3: Conectar con el repositorio nuevo de tu proyecto (Opcional si ya lo crearon en GitHub/GitLab)
🟡 **Modificar antes de ejecutar**:
Si ya tienes la URL del repositorio vacío para tu proyecto nuevo, reemplaza **`URL_DE_TU_NUEVO_REPOSITORIO`** por esa dirección:
```bash
git remote add origin URL_DE_TU_NUEVO_REPOSITORIO
git branch -M main
git push -u origin main
```

---

## 🏷️ Fase 3: Renombrar el Proyecto y Configurar `.env` con `jrmp init`

Internamente, todos los archivos del código dicen `PlantillaInstitucional`. Con el comando `jrmp init` vamos a cambiar ese nombre interno por el de tu proyecto y a crear tu archivo `.env` listo para usar.

### Paso 3.1: Inicializar el proyecto con tu nombre

🟡 **Modificar antes de ejecutar**:
Cambia únicamente la palabra **`MiNuevoProyecto`** por el nombre de tu proyecto (sin espacios ni acentos, igual que en la Fase 1). Este comando funciona igual en Mac, Linux y Windows (y si lo ejecutas solo como `jrmp init`, te preguntará el nombre interactivamente):

```bash
jrmp init MiNuevoProyecto
```
*(Este comando renombra `go.mod`, actualiza todos los archivos `.go` y `.templ`, desvincula `origin` si seguía apuntando a la plantilla original y genera automáticamente tu archivo `.env` a partir de `.env.example`).*

### Paso 3.2: (Opcional) Revisar los datos de tu proyecto en el archivo `.env`

Si deseas personalizar algún parámetro adicional, abre la carpeta de tu proyecto en **Visual Studio Code** (`code .`) y revisa el archivo **`.env`** en la raíz:

| Línea en `.env` | Qué significa | Valor configurado por `jrmp init` |
| :--- | :--- | :--- |
| `AMBIENTE=LOCAL` | Indica que estás trabajando en tu computadora. | `LOCAL` |
| `SERVICIO=...` | Nombre corto de tu sistema (para logs y métricas). | Nombre de tu proyecto en minúsculas |
| `PQ_WRITE_DB=...` | Nombre de la base de datos local. | Nombre de tu proyecto en minúsculas |
| `PQ_WRITE_USER=postgres` | Usuario de la base de datos local. | `postgres` |
| `PQ_WRITE_PASSWORD=postgres` | Contraseña de la base de datos local. | `postgres` |
| `PQ_READ_DB=...` | Coincide con `PQ_WRITE_DB`. | Nombre de tu proyecto en minúsculas |
| `PQ_READ_USER=postgres` | Coincide con `PQ_WRITE_USER`. | `postgres` |
| `PQ_READ_PASSWORD=postgres` | Coincide con `PQ_WRITE_PASSWORD`. | `postgres` |

---

## 🚀 Fase 4: Encender tu Proyecto por Primera Vez

Ahora que tu proyecto tiene tu propio nombre y configuración, vamos a levantarlo con Docker.

### Paso 4.1: Construir y encender el sistema
🟢 **Copiar y pegar tal cual**:
```bash
jrmp up
```
*(O su equivalente con Docker Compose: `docker compose -f docker-compose.local.yml up --build -d`. La primera vez tardará entre 30 y 60 segundos).*

### Paso 4.2: Entrar al sistema desde tu navegador

1. Abre tu navegador (Chrome, Edge, Safari, etc.) y entra a:
   👉 **[http://localhost:8080/login](http://localhost:8080/login)**
2. Inicia sesión con el **usuario administrador de prueba** que ya viene precargado:
   - **Correo electrónico**: `admin@issemym.gob.mx`
   - **Contraseña**: `admin123`

---

## 🧰 Fase 5: Comandos `jrmp` para tu Día a Día

Guarda esta tabla a la mano; funcionan igual en **Mac, Linux y Windows**:

| ¿Qué quieres hacer? | Comando con `jrmp` (🟢 Copiar y pegar) |
| :--- | :--- |
| **Ver la lista de todos los comandos disponibles** | `jrmp help` |
| **Encender o actualizar el sistema en Docker** | `jrmp up` |
| **Ver los mensajes o errores del servidor en vivo** | `jrmp logs` *(presiona `Ctrl + C` para salir)* |
| **Apagar el sistema cuando termines de trabajar** | `jrmp down` |
| **Crear un nuevo módulo CRUD (interactivo o directo)** | `jrmp new-module` *(o `jrmp new-module proveedor`)* |
| **Eliminar un módulo de prueba (y limpiar la BD)** | `jrmp remove-module proveedor` |
| **Verificar formato Templ, linter y pruebas unitarias** | `jrmp check` |
| **Auditar vulnerabilidades y seguridad (govulncheck + Trivy)** | `jrmp security` |
| **Borrar la base de datos local y empezar desde cero** | `docker compose -f docker-compose.local.yml down -v` |

---

## 💾 Fase 6: Cómo Guardar tus Avances en Git (Sin Miedo)

Cada vez que termines una tarea y todo funcione bien en tu navegador, guarda tus cambios siguiendo estos 4 pasos en la terminal:

1. **Ver qué archivos modificaste** (🟢 Copiar y pegar tal cual):
   ```bash
   git status
   ```
2. **Preparar todos tus archivos para guardarlos** (🟢 Copiar y pegar tal cual):
   ```bash
   git add .
   ```
3. **Guardar una fotografía de tus cambios con una breve descripción** (🟡 Cambia el texto entre comillas por lo que hiciste):
   ```bash
   git commit -m "feat: configurar proyecto inicial y agregar modulo de proveedores"
   ```
4. **Subir tus cambios al servidor de Git** (🟢 Copiar y pegar tal cual):
   ```bash
   git push
   ```

---

## ⏭️ Siguiente Paso: ¡Crea tu Primer Módulo!

Ahora que tu proyecto está corriendo, aprende cómo crear pantallas completas con tabla, buscador, ventana de registro/edición y reportes en PDF y Excel con un solo comando:

👉 **[Ir a la Guía Amigable para Crear Módulos CRUD (Scaffolding)](./scaffolding.md)**
