# 🚀 Despliegue en Google Cloud Run

Este documento detalla las configuraciones, adaptaciones y mejores prácticas para desplegar y operar la **Plantilla Institucional** de forma exitosa en el entorno serverless de **Google Cloud Run**.

---

## 🏛️ Diseño de Compatibilidad Serverless

La aplicación ha sido adaptada para cumplir con los principios de las aplicaciones de 12 factores en entornos de contenedores efímeros y auto-escalables:

1.  **Puerto Dinámico**: Cloud Run asigna dinámicamente el puerto mediante la variable de entorno `PORT`. La aplicación está configurada para mapear este puerto directamente y resolver `ServerListen = ":" + PORT` si no se configura `LISTEN_ADDR`.
2.  **Stateless (Sin Estado)**: Toda persistencia de archivos físicos se delega a **Google Cloud Storage (GCS)** a través del paquete interno de `archivos`, evitando asumir que el disco local es persistente.
3.  **Hardenización de Seguridad**: El contenedor corre bajo un usuario y grupo de sistema sin privilegios (`appuser` con UID `10001`), evitando la ejecución como `root`.
4.  **Desacople de Migraciones**: Para evitar race conditions y locks competitivos al escalar a múltiples instancias en paralelo, se proporciona soporte para evitar correr migraciones al arrancar la app.

---

## ⚙️ Variables de Entorno en Producción

Al desplegar en Cloud Run, debés configurar las siguientes variables de entorno a través del panel de GCP o tus scripts de Terraform/gcloud:

| Variable | Propósito | Valor Sugerido / Ejemplo |
| :--- | :--- | :--- |
| `AMBIENTE` | Determina las optimizaciones de ejecución | `"PRODUCCION"` o `"DESARROLLO"` |
| `PORT` | Puerto de escucha dinámico (inyectado por Cloud Run) | `8080` (automático) |
| `SKIP_MIGRATIONS` | Bandera para desactivar migraciones automáticas al bootear | `true` |
| `TENANT_AUTHORITY_CONFIG` | Perfil de confianza, reglas canónicas y snapshot autoritativo vigente. Es obligatorio en `DESARROLLO` y `PRODUCCION`; el arranque falla si falta o no es válido. | JSON validado; consultar `docs/tenant-authority-contract.md` |
| `PQ_WRITE_HOST` | Host para conexión de base de datos de escritura | `"/cloudsql/tu-proyecto:us-central1:instancia"` (Unix Socket) o IP Privada |
| `PQ_WRITE_USER` | Usuario de PostgreSQL | `"usuplantilla"` |
| `PQ_WRITE_PASSWORD`| Contraseña de PostgreSQL | *(Secrets Manager / Secret Env)* |
| `PQ_WRITE_DB` | Base de datos principal | `"plantilla"` |

---

## 🗄️ Estrategia de Migraciones en Producción

Correr migraciones concurrentemente en sistemas serverless genera fallos de conexión y locks competitivos. La estrategia recomendada es **desacoplar las migraciones del ciclo de vida del servidor web**:

1.  **Desactivar en la App**: Seteá la variable de entorno `SKIP_MIGRATIONS=true` en tu configuración de Cloud Run.
2.  **Correr en Cloud Build**: Antes de desplegar la nueva versión del servicio de Cloud Run, ejecutá un paso en tu pipeline de integración continua que levante de forma aislada la migración:
    ```yaml
    # Ejemplo de paso en cloudbuild.yaml
    - name: 'gcr.io/tu-proyecto/plantilla:latest'
      entrypoint: './main'
      args: ['migrate']
      env:
        - 'SKIP_MIGRATIONS=false'
        - 'PQ_WRITE_HOST=IP_O_SOCKET_DE_BD'
        # ... credenciales ...
     ```
3.  **Correr como Cloud Run Job**: Alternativamente, se puede invocar un **Cloud Run Job** aislado con el comando `./main migrate` y `SKIP_MIGRATIONS=false`. Este comando ejecuta únicamente las migraciones y termina correctamente al finalizar la actualización de la base de datos.

---

## 🐳 Contenedor Seguro (Dockerfile)

El `Dockerfile` del proyecto está optimizado con una estructura multi-etapa:

*   **Etapa 1 (Build)**: Genera los archivos estáticos de Templ (`templ generate`) y compila el binario binario estático de Go.
*   **Etapa 2 (Runtime)**: Utiliza `alpine:latest` como base ligera, instala certificados de CA para conexiones HTTPS seguras externas, crea el usuario `appuser:appuser` (UID `10001`), asigna la propiedad de los directorios clave (`static/` y `sql/`), y cambia el contexto de ejecución a este usuario.
*   *Nota de Seguridad*: No se incluye ningún archivo `.env` dentro de la imagen. La configuración completa se obtiene del entorno inyectado por el runtime.
