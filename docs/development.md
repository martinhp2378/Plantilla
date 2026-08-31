# 🚀 Guía de Desarrollo

Esta sección detalla cómo utilizar la plantilla para iniciar tu propio proyecto y cómo trabajar en el día a día.

## 🌟 Cómo Empezar tu Nuevo Proyecto

Si vas a usar esta plantilla para un proyecto nuevo, sigue estos pasos iniciales:

1.  **Clona el repositorio** (o usa el botón "Use this template" si está en GitHub).
2.  **Renombra el módulo Go**: Usa `find` y `sed` (o buscar y reemplazar en tu IDE) para cambiar `PlantillaInstitucional` por la ruta de tu nuevo módulo en `go.mod` y en todos los `.go` y `.templ`.
3.  **Configura tu entorno**:
    ```bash
    cp .env.example .env
    ```
    *(Ajusta las variables de base de datos, credenciales SMTP y el `DOMAIN` base de tu proyecto).*
4.  **Registra un Tenant (Institución)**:
    El sistema exige multi-tenancy. Debes crear el esquema y un registro inicial en la BD (o utilizar un script seed) antes de poder acceder por el navegador a tu subdominio (ej. `midominio.localhost`).

## 🛠️ Herramientas Recomendadas

*   **Go**: Versión 1.22+.
*   **Templ CLI**: Para generar código Go a partir de archivos `.templ`.
*   **Docker Desktop/OrbStack**: Para gestionar el entorno completo.
*   **VS Code / GoLand**: Con plugins de Go y Templ activados.

## 🏃 Flujo de Trabajo Local

### 1. Entorno Docker (Base de Datos y Servicios)
Para levantar el sistema con todos sus servicios (incluyendo generadores de PDF y PostgreSQL):
```bash
docker-compose -f docker-compose.local.yml up --build -d
```

### 2. Live Reload y Generación
Ejecuta la app usando [Air](https://github.com/cosmtrek/air) para recompilación automática:
```bash
air
```
Si modificas archivos `.templ`, asegúrate de correr `templ generate` o configúralo en tu proceso de build.

## 📊 Reportes PDF y Assets

El sistema de reportes utiliza la librería de archivos interna para localizar recursos.
*   **Fuentes**: Se encuentran en `static/public/font/Gotham`.
    *   *Nota*: Se requiere formato **.ttf** para soporte UTF-8 completo en PDF.
*   **Imágenes**: Logos institucionales en `static/public/img/`.

## 🧪 Pruebas (Testing)

Ejecuta todas las pruebas con validación exhaustiva:
```bash
go test -count=1 ./...
```

## 📦 Gestión de Dependencias

El proyecto utiliza el directorio `vendor/` para garantizar compilaciones consistentes.
```bash
go mod tidy
go mod vendor
```

---

## 📝 Convenciones Críticas

1.  **Transacciones Multi-tenant Obligatorias**: NUNCA accedas directamente a un pool de base de datos. Todos los repositorios DEBEN recibir el contexto e inyectar operaciones a través del closure `r.Tenant.Read(ctx, func(tx pgx.Tx) error)` o `r.Tenant.Write(...)`.
2.  **Cero prefijos de esquema**: NUNCA uses prefijos de esquema explícitos (ej. `institucion.persona`) en las consultas de los repositorios a tablas de negocio. RLS y `search_path` se encargan del aislamiento.
3.  **Rutas**: NUNCA hardcodear rutas como strings manuales. Usa siempre los helpers de `pagina.GlobalRutas`.
4.  **Naming**: PascalCase para campos exportados, camelCase para privados.
