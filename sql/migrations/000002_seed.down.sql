-- 000002_seed.down.sql

-- Delete dependent policies before the catalog rows they reference.
DELETE FROM sesion.politica
WHERE nombre IN ('Acceso Usuarios', 'Acceso Perfiles', 'Acceso Politicas', 'Acceso Dashboard', 'Acceso Personas')
  AND razon_i = 'Carga Inicial';

DELETE FROM sistema.cfg_parametros
WHERE clave IN ('EMAIL_SYSTEM_NAME', 'EMAIL_BRAND_NAME', 'EMAIL_FOOTER_ORGANIZATION', 'EMAIL_SUPPORT_LABEL', 'EMAIL_SUPPORT_PHONE')
  AND razon_i = 'Carga Inicial';

DELETE FROM institucion.cat_tenants WHERE subdomain = 'admin' AND schema_name = 'public';

DELETE FROM sesion.perfil
WHERE nombre IN ('superadmin', 'operador', 'ejecutivo', 'anonimo')
  AND razon_i IN ('Administrador Total del Sistema', 'Personal de Unidad Médica', 'Consulta de Indicadores', 'Perfil para usuarios no autenticados');

DELETE FROM sesion.recurso
WHERE nombre IN ('usuarios', 'dashboard', 'configuracion', 'institucion', 'superadmin', 'parametros')
  AND razon_i = 'Carga Inicial';

DELETE FROM sesion.cat_accion
WHERE nombre IN ('buscar', 'consultar', 'modificar', 'crear', 'eliminar', 'aprobar')
  AND razon_i = 'Carga Inicial';
