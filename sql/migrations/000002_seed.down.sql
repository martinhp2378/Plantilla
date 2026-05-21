-- 000002_seed.down.sql

DELETE FROM sesion.usuario WHERE email = 'admin@issemym.gob.mx';
DELETE FROM institucion.persona WHERE clave_issemym = '0000000';
DELETE FROM institucion.cat_tenants WHERE subdomain = 'admin';
DELETE FROM sesion.perfil WHERE nombre IN ('admin', 'operador', 'ejecutivo');
DELETE FROM sesion.recurso WHERE nombre IN ('usuarios', 'dashboard', 'configuracion', 'institucion');
DELETE FROM sesion.cat_accion WHERE nombre IN ('buscar', 'consultar', 'modificar', 'crear', 'eliminar', 'aprobar');
