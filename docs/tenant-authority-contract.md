# 🏢 Configuración de Dominios e Instituciones (Autoridad de Tenant)

> [!TIP]
> **¿Necesitas configurar esto para trabajar en tu computadora?**
> **No.** Mientras tu archivo `.env` tenga `AMBIENTE=LOCAL` y `TENANT_AUTHORITY_CONFIG=` vacío, la aplicación configura automáticamente `localhost` para que puedas trabajar sin hacer nada extra.
>
> Este documento solo aplica cuando el sistema se despliega en servidores de **`DESARROLLO`** o **`PRODUCCION`**.

---

## 🛡️ 1. ¿Para qué sirve la Autoridad de Tenant?

Antes de procesar cualquier petición web, el sistema verifica qué dominio o subdominio escribió el usuario en su navegador (por ejemplo, `inventario.issemym.gob.mx`) para saber a qué institución o área (`tenant`) pertenece y asignarle su propio esquema de base de datos.

Si alguien intenta entrar desde un dominio desconocido o manipulando cabeceras HTTP, el sistema rechaza la petición automáticamente con el código HTTP **`421 Misdirected Request`** antes de tocar la base de datos.

---

## ⚙️ 2. Perfiles de Despliegue Soportados

Dependiendo de dónde esté instalado el sistema, se elige uno de estos 3 perfiles dentro del JSON:

| Perfil (`deployment_profile`) | Cuándo se usa | Cómo valida el dominio |
| :--- | :--- | :--- |
| **`local_direct`** | Automático en tu computadora (`AMBIENTE=LOCAL`). | Acepta directamente `localhost`. No se permite en producción. |
| **`gcp_cloud_run_alb`** | Despliegue en **Google Cloud Run** detrás de un Balanceador de Carga (ALB). | Toma el dominio del header `Host` enviado por el balanceador bajo `HTTPS`. |
| **`onprem_nginx_single_hop`** | Despliegue en servidores propios detrás de un proxy **NGINX**. | Valida que la petición venga de la IP autorizada del NGINX (`nginx_trusted_cidrs`) y lee `X-Forwarded-Host`. |

---

## 📋 3. Ejemplo de JSON para `TENANT_AUTHORITY_CONFIG` (Listo para adaptar)

En los ambientes `DESARROLLO` y `PRODUCCION`, debes definir la variable de entorno `TENANT_AUTHORITY_CONFIG` con un JSON como el siguiente (en una sola línea o inyectado desde Secret Manager):

🟡 **Modificar antes de usar en Desarrollo/Producción**:
- Cambia `sistema.issemym.gob.mx` por el dominio real de tu aplicación.
- Cambia `tenant_id` por el UUID real registrado en la tabla `institucion.cat_tenants`.
- Ajusta las fechas de vigencia `valid_from` y `valid_until` en formato ISO (`YYYY-MM-DDTHH:MM:SSZ`).

```json
{
  "deployment_profile": "gcp_cloud_run_alb",
  "allowed_schemes": ["https"],
  "global_access_policy_id": "global-readers-v1",
  "canonical_hosts": [
    {
      "host": "sistema.issemym.gob.mx",
      "tenant_id": "6163cc43-be78-428c-8fd6-7ebc3e961f3f",
      "schema": "admin",
      "active": true
    }
  ],
  "catalog": {
    "source": "tenant-catalog-db",
    "version": "2026-10-01.1",
    "valid_from": "2026-01-01T00:00:00Z",
    "valid_until": "2027-12-31T23:59:59Z",
    "entries": [
      {
        "host": "sistema.issemym.gob.mx",
        "tenant_id": "6163cc43-be78-428c-8fd6-7ebc3e961f3f",
        "schema": "admin",
        "institution_owner": "issemym-ti",
        "platform_owner": "plataforma-cloud",
        "source": "tenant-catalog-db",
        "version": "2026-10-01.1",
        "lifecycle": "active",
        "institution_approved": true,
        "platform_approved": true,
        "evidence_approved": true,
        "valid_from": "2026-01-01T00:00:00Z",
        "valid_until": "2027-12-31T23:59:59Z"
      }
    ]
  }
}
```

> [!IMPORTANT]
> Si usas el perfil **`onprem_nginx_single_hop`**, agrega también la propiedad `"nginx_trusted_cidrs": ["10.0.0.0/8"]` en la raíz del JSON con el rango IP de tu servidor NGINX.
