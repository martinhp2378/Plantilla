# 📚 Documentación de la Plantilla Institucional

Bienvenido a la documentación detallada de la **Plantilla Institucional**. Este espacio está diseñado para ayudarte a entender la estructura, los patrones y las herramientas que hacen que esta base sea robusta y escalable.

## 🌱 Guías para Empezar (Paso a Paso y Sin Experiencia Previa)

1.  **[Guía Paso a Paso: De Cero a tu Propio Proyecto](./primeros-pasos.md)**
    *   Qué programas instalar en tu computadora.
    *   Cómo descargar (clonar) la plantilla y desvincularla del repositorio original.
    *   Comandos para copiar y pegar para renombrar el proyecto con los datos de tu sistema.
    *   Cómo encender el proyecto en Docker, iniciar sesión y guardar tus avances en Git.
2.  **[Guía Amigable: Cómo Crear y Personalizar un Nuevo Módulo CRUD](./scaffolding.md)**
    *   Cómo crear una pantalla completa con `jrmp new-module` (tabla, buscador, modales y reportes PDF/Excel).
    *   Cómo borrar un módulo de prueba sin dejar residuos en la base de datos.
    *   Cómo agregar campos nuevos paso a paso.

---

## 🧭 Referencia Técnica y Arquitectura

3.  **[Arquitectura](./architecture.md)**
    *   Arquitectura Hexagonal (Puertos y Adaptadores).
    *   Registro modular autónomo e Inyección de Dependencias (`app.App`).
    *   Rutas descentralizadas y Multi-tenancy con RLS.
    *   Ciclo de vida y retención de reportes privados.
4.  **[Guía de Desarrollo](./development.md)**
    *   Configuración del entorno local y comandos `make`.
    *   Flujo de trabajo con Docker, Templ y Air (Live Reload).
    *   Convenciones críticas de código y base de datos.
5.  **[Frontend y Componentes](./ui-components.md)**
    *   Uso de Templ y HTMX.
    *   Componentes v2 (Tablas dinámicas, Modales y Notificaciones OOB).
    *   Estructura de Layouts.
6.  **[Autenticación y Autorización](./auth.md)**
    *   Gestión de Sesiones y JWT.
    *   Middleware de Políticas ABAC.
    *   Auditoría de contexto.
7.  **[Base de Datos y Migraciones](./database.md)**
    *   Estructura de esquemas y aislamiento RLS.
    *   Gestión de migraciones con `golang-migrate`.
    *   Repositorios sobre `TenantTxRunner`.
8.  **[Despliegue en Cloud Run](./cloud_run_deployment.md)**
    *   Diseño de compatibilidad serverless (puerto dinámico, stateless).
    *   Variables de entorno en producción.
    *   Estrategia de migraciones desacopladas (`SKIP_MIGRATIONS`).
    *   Securización del contenedor no-root.
9.  **[Contrato de Autoridad de Tenant](./tenant-authority-contract.md)**
    *   Perfiles de confianza, denegación segura y límites de evidencia de despliegue.

---

## 🎯 Objetivo de la Documentación

Esta documentación no solo explica *qué* hace la plantilla, sino *por qué* se eligieron ciertos patrones (como HTMX sobre React) para entornos institucionales donde la mantenibilidad y la velocidad de carga son críticas.
