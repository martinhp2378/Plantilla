# ☁️ Guía de Despliegue en Producción (Google Cloud Run)

> [!NOTE]
> **¿Cuándo necesitas esta guía?**
> Si estás desarrollando o probando en tu computadora (`AMBIENTE=LOCAL`), **no necesitas configurar nada de lo que aparece aquí**. Esta guía es únicamente para el momento en que el proyecto se sube a los servidores de **Desarrollo** o **Producción** en Google Cloud Run.

---

## 🏛️ 1. Cómo Funciona la Plantilla en la Nube

La aplicación ya viene preparada de fábrica para funcionar en contenedores de **Google Cloud Run**:

1. **Puerto Automático (`PORT`)**: Cloud Run asigna el puerto automáticamente mediante la variable `PORT`, y la aplicación lo toma sin que tengas que cambiar el código.
2. **Almacenamiento de Archivos en Cloud Storage (GCS)**: En producción, las fotos y los reportes privados en PDF/Excel se guardan en cubetas (*buckets*) de Google Cloud Storage en lugar del disco local.
3. **Contenedor Seguro Sin Permisos de Root**: El `Dockerfile` compila la aplicación y la ejecuta con un usuario restringido (`appuser`, UID `10001`) para cumplir con las auditorías de seguridad institucional.

---

## ⚙️ 2. Variables de Entorno Obligatorias en Producción

Cuando configures el servicio en el panel de Google Cloud Run (o mediante Terraform / `gcloud`), asegúrate de definir estas variables:

| Variable | Qué valor poner | Para qué sirve |
| :--- | :--- | :--- |
| `AMBIENTE` | `DESARROLLO` o `PRODUCCION` | Activa las validaciones estrictas de seguridad y almacenamiento en la nube. |
| `PORT` | *(Lo asigna Cloud Run solo)* | Puerto en el que escucha el contenedor (usualmente `8080`). |
| `SKIP_MIGRATIONS` | `true` | Evita que varias instancias en la nube intenten correr migraciones al mismo tiempo al encenderse. |
| `TENANT_AUTHORITY_CONFIG` | JSON de configuración de dominios | Define qué dominios institucionales están autorizados (ver [Contrato de Autoridad de Tenant](./tenant-authority-contract.md)). |
| `PQ_WRITE_HOST` | Socket de Cloud SQL o IP privada | Dirección de la base de datos de escritura (ej. `/cloudsql/proyecto:region:instancia`). |
| `PQ_WRITE_USER` | Usuario de BD productiva | Usuario con permisos de aplicación (`app_runtime_tenant`). |
| `PQ_WRITE_PASSWORD` | *(Desde Secret Manager)* | Contraseña de la base de datos de escritura. |
| `PQ_WRITE_DB` | Nombre de la base de datos | Nombre de la base de datos del sistema en Cloud SQL. |
| `COOKIE_NOMBRE` | Nombre único para tu sistema | Ej. `inventario_prod_v1` (para que la cookie de sesión sea exclusiva de tu sistema). |
| `COOKIE_ES_SECURE` | `true` | Exige que la cookie de sesión viaje únicamente por conexiones seguras `HTTPS`. |

---

## 🗄️ 3. Cómo Ejecutar las Migraciones en Producción

En tu computadora local las migraciones corren solas al encender la app, pero en producción (donde Cloud Run puede levantar varios contenedores al mismo tiempo) lo recomendable es correr las migraciones en un paso previo antes de actualizar el servidor web:

1. Configura `SKIP_MIGRATIONS=true` en las variables del servicio web de Cloud Run.
2. En tu pipeline de despliegue (**Cloud Build** o un **Cloud Run Job**), ejecuta el binario pasándole el argumento `migrate` con `SKIP_MIGRATIONS=false`:

```yaml
# Ejemplo para copiar y adaptar en cloudbuild.yaml
- name: 'gcr.io/TU-PROYECTO-GCP/TU-SISTEMA:latest'
  entrypoint: './main'
  args: ['migrate']
  env:
    - 'SKIP_MIGRATIONS=false'
    - 'PQ_WRITE_HOST=/cloudsql/TU-PROYECTO:REGION:INSTANCIA'
```
*(El comando `./main migrate` aplica las tablas nuevas en PostgreSQL y se apaga inmediatamente al terminar).*
