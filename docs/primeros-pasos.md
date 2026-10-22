# 🌱 Guía Paso a Paso: De Cero a tu Propio Proyecto

Esta guía está diseñada para acompañarte desde que descargas la **Plantilla Institucional** por primera vez hasta que tienes tu propio sistema funcionando en tu computadora, renombrado con los datos de tu proyecto. **No necesitas experiencia previa en Go, Git ni Docker para seguirla.**

Para que sea muy fácil de seguir, cada comando tiene una etiqueta:
- 🟢 **Copiar y pegar tal cual**: No tienes que cambiarle nada, solo cópialo, pégalo en tu terminal y presiona `Enter`.
- 🟡 **Modificar antes de ejecutar**: Te indicaremos exactamente qué palabra debes cambiar por los datos de tu proyecto.

---

## 🛠️ Fase 0: Programas que necesitas instalar (solo la primera vez)

Antes de empezar, asegúrate de tener instalados estos 5 programas en tu computadora:

| Programa | Para qué sirve | Enlace o Comando de instalación |
| :--- | :--- | :--- |
| **1. Git** | Descargar el código y guardar tu historial de cambios. | [Descargar Git](https://git-scm.com/downloads) |
| **2. Docker Desktop** | Levantar la base de datos PostgreSQL y la aplicación sin instalar nada extra. | [Descargar Docker Desktop](https://www.docker.com/products/docker-desktop/) |
| **3. Go (1.25+)** | El lenguaje de programación del sistema (necesario para el generador de módulos). | [Descargar Go](https://go.dev/dl/) |
| **4. Visual Studio Code** | El editor de código recomendado para ver y modificar tus archivos. | [Descargar VS Code](https://code.visualstudio.com/) |
| **5. Templ CLI** | Compila las vistas HTML (`.templ`) en código Go cuando creas o editas pantallas. | Ver comando abajo 👇 |

### Cómo instalar Templ CLI (una vez que ya instalaste Go)

Abre tu terminal y ejecuta este comando:

🟢 **Copiar y pegar tal cual**:
```bash
go install github.com/a-h/templ/cmd/templ@latest
```

Si usas **Mac (macOS)** o **Linux**, asegúrate de que tu terminal encuentre el comando `templ` ejecutando también esta línea (solo necesitas hacerlo una vez):

🟢 **Copiar y pegar tal cual**:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
grep -qxF 'export PATH="$PATH:$(go env GOPATH)/bin"' ~/.zshrc 2>/dev/null || echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
```

Para comprobar que quedó bien instalado, escribe:
🟢 **Copiar y pegar tal cual**:
```bash
templ version
```

> [!IMPORTANT]
> **Antes de continuar**, abre la aplicación **Docker Desktop** en tu computadora y espera a que el ícono de la ballena indique que está en ejecución (*Engine running*).

---

## 📥 Fase 1: Descargar (Clonar) la Plantilla

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

## 🏷️ Fase 3: Renombrar el Proyecto con tus Datos

Internamente, todos los archivos del código dicen `PlantillaInstitucional`. Ahora vamos a cambiar ese nombre interno por el de tu proyecto y a configurar tus datos básicos.

### Paso 3.1: Cambiar el nombre interno en todo el código

🟡 **Modificar antes de ejecutar**:
En el comando correspondiente a tu sistema operativo, cambia únicamente la palabra **`MiNuevoProyecto`** (que aparece 2 veces) por el nombre de tu proyecto (sin espacios ni acentos, igual que en la Fase 1):

**Si usas Mac (macOS):**
```bash
go mod edit -module MiNuevoProyecto
find . -type f \( -name "*.go" -o -name "*.templ" \) -not -path "./vendor/*" -not -path "./.git/*" -exec sed -i '' 's/PlantillaInstitucional/MiNuevoProyecto/g' {} +
```

**Si usas Linux o Windows (Git Bash / WSL):**
```bash
go mod edit -module MiNuevoProyecto
find . -type f \( -name "*.go" -o -name "*.templ" \) -not -path "./vendor/*" -not -path "./.git/*" -exec sed -i 's/PlantillaInstitucional/MiNuevoProyecto/g' {} +
```

### Paso 3.2: Crear tu archivo de configuración (`.env`)

🟢 **Copiar y pegar tal cual**:
```bash
cp .env.example .env
```

### Paso 3.3: Ajustar los datos de tu proyecto en el archivo `.env`

Abre la carpeta de tu proyecto en **Visual Studio Code** (puedes escribir `code .` en la terminal o abrir la carpeta desde el menú *File -> Open Folder*).

Busca el archivo llamado **`.env`** en la raíz del proyecto y revisa estas líneas al inicio del archivo:

| Línea en `.env` | Qué significa | Qué valor poner |
| :--- | :--- | :--- |
| `AMBIENTE=LOCAL` | Indica que estás trabajando en tu computadora. | **Déjalo en `LOCAL`** |
| `SERVICIO=plantilla` | Nombre corto de tu sistema (para logs y métricas). | Cámbialo por el nombre de tu proyecto en minúsculas, ej.: `SERVICIO=inventario` |
| ` PQ_WRITE_DB=plantilla` | Nombre de la base de datos local. | Puedes dejarlo en `plantilla` o poner el de tu sistema, ej.: `PQ_WRITE_DB=inventario` |
| `PQ_WRITE_USER=usuplantilla` | Usuario de la base de datos local. | Puedes dejarlo en `postgres` |
| `PQ_WRITE_PASSWORD=` | Contraseña de la base de datos local. | Pon una contraseña para tu base local, ej.: `PQ_WRITE_PASSWORD=postgres` |
| `PQ_READ_DB=plantilla` | Debe coincidir con `PQ_WRITE_DB`. | Pon el mismo valor que pusiste en `PQ_WRITE_DB` |
| `PQ_READ_USER=usuplantilla` | Debe coincidir con `PQ_WRITE_USER`. | Pon el mismo valor que pusiste en `PQ_WRITE_USER` |
| `PQ_READ_PASSWORD=` | Debe coincidir con `PQ_WRITE_PASSWORD`. | Pon la misma contraseña que pusiste en `PQ_WRITE_PASSWORD` |

> [!TIP]
> Guarda el archivo `.env` (`Cmd + S` en Mac o `Ctrl + S` en Windows) después de editarlo.

---

## 🚀 Fase 4: Encender tu Proyecto por Primera Vez

Ahora que tu proyecto tiene tu propio nombre y configuración, vamos a levantarlo con Docker.

### Paso 4.1: Construir y encender el sistema
🟢 **Copiar y pegar tal cual**:
```bash
docker compose -f docker-compose.local.yml up --build -d
```
*(La primera vez tardará entre 30 y 60 segundos mientras descarga las herramientas y prepara la base de datos).*

### Paso 4.2: Entrar al sistema desde tu navegador

1. Abre tu navegador (Chrome, Edge, Safari, etc.) y entra a:
   👉 **[http://localhost:8080/login](http://localhost:8080/login)**
2. Inicia sesión con el **usuario administrador de prueba** que ya viene precargado:
   - **Correo electrónico**: `admin@issemym.gob.mx`
   - **Contraseña**: `admin123`

---

## 🧰 Fase 5: Comandos Útiles para tu Día a Día

Guarda esta tabla a la mano; son los únicos comandos que necesitarás mientras trabajas:

| ¿Qué quieres hacer? | Comando (🟢 Copiar y pegar tal cual) |
| :--- | :--- |
| **Ver si el sistema está encendido** | `docker compose -f docker-compose.local.yml ps` |
| **Ver los mensajes o errores del servidor en vivo** | `docker compose -f docker-compose.local.yml logs -f app` *(presiona `Ctrl + C` para salir)* |
| **Actualizar la app después de hacer cambios en el código** | `docker compose -f docker-compose.local.yml up --build -d app` |
| **Apagar el sistema cuando termines de trabajar** | `docker compose -f docker-compose.local.yml down` |
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
