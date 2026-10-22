# 📚 Documentación de la Plantilla Institucional

Bienvenido a la documentación detallada de la **Plantilla Institucional**. Este espacio está diseñado para ayudarte a entender la estructura, los patrones y las herramientas que hacen que esta base sea robusta y escalable.

## 🧭 Tabla de Contenidos

1.  **[Arquitectura](./architecture.md)**
    *   Arquitectura Hexagonal (Puertos y Adaptadores).
    *   Registro modular autónomo e Inyección de Dependencias (`app.App`).
    *   Rutas descentralizadas y Multi-tenancy con RLS.
    *   Ciclo de vida y retención de reportes privados.
2.  **[Generación de Módulos (Scaffolding)](./scaffolding.md)**
    *   Creación automática de módulos CRUD con `make new-module`.
    *   Normalización de nombres (`snake_case` y `PascalCase`).
    *   Exportación integrada de reportes PDF (con soporte de acentos) y Excel.
    *   Guía de personalización y reversión de módulos de prueba.
3.  **[Guía de Desarrollo](./development.md)**
    *   Configuración del entorno local y comandos `make`.
    *   Flujo de trabajo con Docker, Templ y Air (Live Reload).
    *   Convenciones críticas de código y base de datos.
4.  **[Frontend y Componentes](./ui-components.md)**
    *   Uso de Templ y HTMX.
    *   Componentes v2 (Tablas dinámicas, Modales y Notificaciones OOB).
    *   Estructura de Layouts.
5.  **[Autenticación y Autorización](./auth.md)**
    *   Gestión de Sesiones y JWT.
    *   Middleware de Políticas ABAC.
    *   Auditoría de contexto.
6.  **[Base de Datos y Migraciones](./database.md)**
    *   Estructura de esquemas y aislamiento RLS.
    *   Gestión de migraciones con `golang-migrate`.
    *   Repositorios sobre `TenantTxRunner`.
7.  **[Despliegue en Cloud Run](./cloud_run_deployment.md)**
    *   Diseño de compatibilidad serverless (puerto dinámico, stateless).
    *   Variables de entorno en producción.
    *   Estrategia de migraciones desacopladas (`SKIP_MIGRATIONS`).
    *   Securización del contenedor no-root.
8.  **[Contrato de Autoridad de Tenant](./tenant-authority-contract.md)**
    *   Perfiles de confianza, denegación segura y límites de evidencia de despliegue.

---

## 🎯 Objetivo de la Documentación

Esta documentación no solo explica *qué* hace la plantilla, sino *por qué* se eligieron ciertos patrones (como HTMX sobre React) para entornos institucionales donde la mantenibilidad y la velocidad de carga son críticas.
