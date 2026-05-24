# 📚 Documentación de la Plantilla Institucional

Bienvenido a la documentación detallada de la **Plantilla Institucional**. Este espacio está diseñado para ayudarte a entender la estructura, los patrones y las herramientas que hacen que esta base sea robusta y escalable.

## 🧭 Tabla de Contenidos

1.  **[Arquitectura](./architecture.md)**
    *   Arquitectura Hexagonal (Puertos y Adaptadores).
    *   Inyección de Dependencias.
    *   Multi-tenancy basada en esquemas.
2.  **[Guía de Desarrollo](./development.md)**
    *   Configuración del entorno local.
    *   Comandos útiles (Make, Templ, Docker).
    *   Flujo de trabajo con Air (Live Reload).
3.  **[Frontend y Componentes](./ui-components.md)**
    *   Uso de Templ y HTMX.
    *   Componentes v2 (Alpine.js + Tailwind).
    *   Estructura de Layouts.
4.  **[Autenticación y Autorización](./auth.md)**
    *   Gestión de Sesiones y JWT.
    *   Middleware de Políticas.
    *   Auditoría de contexto.
5.  **[Base de Datos y Migraciones](./database.md)**
    *   Estructura de esquemas.
    *   Gestión de migraciones con golang-migrate.
    *   Repositores y Pgxpool.
6.  **[Despliegue en Cloud Run](./cloud_run_deployment.md)**
    *   Diseño de compatibilidad serverless (puerto dinámico, stateless).
    *   Variables de entorno en producción.
    *   Estrategia de migraciones desacopladas (SKIP_MIGRATIONS).
    *   Securización del contenedor no-root.

---

## 🎯 Objetivo de la Documentación

Esta documentación no solo explica *qué* hace la plantilla, sino *por qué* se eligieron ciertos patrones (como HTMX sobre React) para entornos institucionales donde la mantenibilidad y la velocidad de carga son críticas.
