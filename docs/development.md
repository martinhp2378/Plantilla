# 🚀 Guía de Desarrollo

Esta sección detalla cómo trabajar en el día a día con la plantilla.

## 🛠️ Herramientas Recomendadas

*   **Go**: Versión 1.22+.
*   **Templ CLI**: Para generar código Go a partir de archivos `.templ`.
*   **Air**: Para recarga en vivo del servidor Go.
*   **Docker Desktop/OrbStack**: Para gestionar la base de datos local.
*   **VS Code / GoLand**: Con plugins de Go y Templ activados.

## 🏃 Flujo de Trabajo Local

### 1. Variables de Entorno
Copia el archivo de ejemplo y ajusta los valores si es necesario:
```bash
cp .env.example .env
```

### 2. Base de Datos
Levanta el contenedor de PostgreSQL (esquemas y seeds se cargan automáticamente):
```bash
docker-compose -f docker-compose.local.yml up -d db
```

### 3. Generación de Vistas
Si modificas archivos `.templ`, debes regenerar el código Go:
```bash
templ generate
```
También puedes usar el modo "watch": `templ generate --watch`.

### 4. Ejecución con Live Reload
Usa Air para que el servidor se reinicie automáticamente al detectar cambios:
```bash
air
```

## 🧪 Pruebas (Testing)

El proyecto fomenta el uso de pruebas unitarias. Ejecuta todas las pruebas con:
```bash
go test ./...
```

## 📦 Gestión de Dependencias

Usamos `go mod`. Si agregas una nueva dependencia, ejecuta:
```bash
go mod tidy
go mod vendor
```
*(Nota: El proyecto prefiere mantener las dependencias en `vendor/` para compilaciones consistentes).*

---

## 📝 Convenciones de Código

*   **Idiomas**: Código e identificadores en inglés. Comentarios y documentación en español (o inglés según el equipo).
*   **Naming**: 
    *   Interfaces: `Service`, `Repository`, `Handler`.
    *   Funciones: PascalCase para exportadas, camelCase para privadas.
*   **Errores**: Siempre envolver errores con contexto: `fmt.Errorf("error al hacer X: %w", err)`.
