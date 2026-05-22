# 🚀 Guía de Desarrollo

Esta sección detalla cómo trabajar en el día a día con la plantilla.

## 🛠️ Herramientas Recomendadas

*   **Go**: Versión 1.22+.
*   **Templ CLI**: Para generar código Go a partir de archivos `.templ`.
*   **Docker Desktop/OrbStack**: Para gestionar el entorno completo.
*   **VS Code / GoLand**: Con plugins de Go y Templ activados.

## 🏃 Flujo de Trabajo Local

### 1. Variables de Env
Copia el archivo de ejemplo:
```bash
cp .env.example .env
```

### 2. Entorno Docker (Visualización)
Para ver el sistema con todos sus servicios (incluyendo generadores de PDF y DB):
```bash
docker-compose -f docker-compose.local.yml up --build -d
```

### 3. Generación de Vistas
Si modificas archivos `.templ`, debes regenerar el código Go:
```bash
templ generate
```

## 📊 Reportes PDF y Assets

El sistema de reportes utiliza la librería de archivos interna para localizar recursos.
*   **Fuentes**: Se encuentran en `static/public/font/Gotham`.
    *   *Nota*: Se requiere formato **.ttf** para soporte UTF-8 completo en PDF.
*   **Imágenes**: Logos institucionales en `static/public/img/`.

## 🧪 Pruebas (Testing)

Ejecuta todas las pruebas con:
```bash
go test ./...
```

## 📦 Gestión de Dependencias

El proyecto utiliza el directorio `vendor/` para garantizar compilaciones consistentes.
```bash
go mod tidy
go mod vendor
```

---

## 📝 Convenciones

*   **Rutas**: NUNCA hardcodear rutas como strings manuales. Usa siempre `pagina.GlobalRutas`.
*   **SQL**: NUNCA usar prefijos de esquema explícitos en las queries de los repositorios para no romper el multi-tenant.
*   **Naming**: PascalCase para exportadas, camelCase para privadas.
