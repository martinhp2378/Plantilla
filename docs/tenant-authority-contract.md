# Contrato de autoridad de tenant

La plantilla resuelve el tenant antes de cualquier handler. La autoridad no comprobada se deniega con `421 Misdirected Request`; la respuesta HTTP es genérica y no expone host, tenant, esquema, IP, URL, headers, secretos ni PII.

## Contrato aplicable

| Perfil | Autoridad permitida | Rechazos obligatorios |
|---|---|---|
| `gcp_cloud_run_alb` | `Host` preservado por el ALB; `X-Forwarded-Proto` sólo valida el esquema. | `X-Forwarded-Host`, valores conflictivos, multihop o acceso directo no probado. No se infiere confianza por CIDR. |
| `onprem_nginx_single_hop` | Un `X-Forwarded-Host` de un peer incluido en `NginxTrustedCIDRs`; `Host` upstream válido y un `X-Forwarded-Proto` permitido. | Peer fuera de CIDR, header ausente, duplicado, conflictivo, `Forwarded` o cadena de proxies. |
| `local_direct` | `Host` directo de una fixture local explícita. | Cualquier header `Forwarded` o `X-Forwarded-*`, o política productiva. |

`X-Forwarded-For` nunca selecciona un tenant. Un único valor puede existir para atribución de cliente sin alterar la autoridad; una cadena multihop se deniega. Sólo una entrada de catálogo `active` y asociada exactamente a una regla canónica puede continuar. No hay fallback a `public`.

## Diagnóstico y efectos laterales

Los diagnósticos internos usan reason codes estables y de baja cardinalidad: `authority_untrusted_peer`, `authority_conflicting_headers`, `authority_multihop`, `authority_invalid`, `tenant_unknown` y `tenant_inactive`. La respuesta pública conserva el estado determinista `421` y no revela el reason code ni datos de entrada.

Un rechazo ocurre antes de invocar el handler downstream. Por ello no debe abrir una operación tenant-scoped ni seleccionar un esquema de base de datos. Las pruebas locales ejercitan esta matriz mediante handlers centinela; no son evidencia de infraestructura externa.

## Fixtures e inventario

Las fixtures locales son exclusivamente para `local_direct` y no sustituyen el snapshot DB autoritativo de producción. La referencia `plantilla-institucional.issemym.gob.mx` permanece en estado `provisioning`, no enrutable: cada sistema nuevo debe sustituirla o aprobarla explícitamente con ownership, evidencia, versión y ventana válidos.

## Límites de runtime y despliegue

La aplicación valida el contrato recibido, pero no puede demostrar desde una solicitud la topología de Cloud Run, ALB, serverless NEG, URL `run.app`, CIDRs reales ni la configuración efectiva de NGINX. Esas pruebas son evidencia late-bound y bloquean el cutover hasta que plataforma y el owner institucional las aprueben.

El rollout inicia con un canary limitado a un host `active` aprobado. Debe observar rechazos redactados, errores y latencia dentro de umbrales aprobados. Si falla autoridad, catálogo o observabilidad, el rollback retira el mapping o el host rule y conserva la denegación segura; nunca restaura `public` ni habilita una ruta de bypass.
## Aislamiento Transaccional (Base de Datos)

El cumplimiento del tenant context no se detiene en la capa HTTP. La plantilla fuerza una arquitectura **fail-closed** a nivel transaccional y de motor de base de datos (PostgreSQL).

1. **Aislamiento en Memoria**: `TenantTxRunner` requiere una instancia de contexto inmutable válida para cada unidad de trabajo. Inyecta la resolución (`search_path` y `app.tenant_id`) a través de un set transaccional (`SET LOCAL`), garantizando que la configuración sea efímera y se elimine automáticamente en cada `commit` o `rollback`. No existe un fallback a la red pública (schema `public`).
2. **Defensa en Profundidad (RLS)**: La identidad inyectada en tiempo de transacción se evalúa internamente a través de restricciones Row Level Security (RLS) habilitadas y forzadas (`FORCE ROW LEVEL SECURITY`) en cada tabla, validando y requiriendo validación en escrituras (`WITH CHECK`).
3. **Mínimo Privilegio**: La aplicación web se conecta usando un rol `NOBYPASSRLS` (`app_runtime_tenant`). Se impide programáticamente el acceso cruzado de datos.
