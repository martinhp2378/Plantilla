-- 000009_abac_policies_and_profiles.up.sql
-- ABAC granular policies, canonical application profiles and backward compatibility aliases

SET search_path TO sesion, institucion, sistema, public;

-- 1. Ensure required actions exist
INSERT INTO cat_accion (nombre, razon_i) VALUES 
('buscar', 'ABAC Granular'), 
('consultar', 'ABAC Granular'), 
('modificar', 'ABAC Granular'), 
('crear', 'ABAC Granular'), 
('eliminar', 'ABAC Granular'),
('aprobar', 'ABAC Granular')
ON CONFLICT (nombre) DO UPDATE SET
    razon_m = 'ABAC Granular',
    fecha_m = NOW();

-- 2. Ensure resources exist
INSERT INTO recurso (nombre, descripcion, razon_i) VALUES 
('usuarios', 'Gestión de Usuarios y Perfiles', 'ABAC Granular'),
('perfiles', 'Catálogo y Configuración de Perfiles', 'ABAC Granular'),
('politicas', 'Políticas de Seguridad y Acceso', 'ABAC Granular'),
('dashboard', 'Acceso al Dashboard Administrativo', 'ABAC Granular'),
('configuracion', 'Módulos Generales de Configuración', 'ABAC Granular'),
('institucion', 'Módulos de Institución y Personas', 'ABAC Granular'),
('parametros', 'Gestión de Parámetros del Sistema', 'ABAC Granular'),
('crud', 'Módulo de Ejemplo CRUD', 'ABAC Granular'),
('todas_las_rutas', 'Acceso Global a Rutas del Sistema', 'ABAC Granular')
ON CONFLICT (nombre) DO UPDATE SET
    descripcion = EXCLUDED.descripcion,
    razon_m = 'ABAC Granular',
    fecha_m = NOW();

-- 3. Canonical application profiles and compatibility aliases for sesiones v0.4.0
INSERT INTO perfil (nombre, razon_i) VALUES 
('superadmin', 'Administrador Total del Sistema'), 
('admin', 'Administrador de Configuración y Operación'), 
('operativo', 'Personal Operativo de Unidad'), 
('ejecutivo', 'Consulta de Indicadores y Reportes'),
('anonimo', 'Perfil Público no Autenticado'),
('coordinacion', 'Personal de Coordinación'),
('medico', 'Compat sesiones v0.4.0'),
('recepcion', 'Compat sesiones v0.4.0'),
('adminclinica', 'Compat sesiones v0.4.0')
ON CONFLICT (nombre) DO UPDATE SET
    razon_m = 'ABAC Granular',
    fecha_m = NOW();

-- 4. Default profile for users
ALTER TABLE sesion.usuario 
ALTER COLUMN perfil SET DEFAULT 'anonimo';

-- 5. Migrate any existing policies that lack 'routes' into the structured JSONB format
UPDATE sesion.politica
SET reglas = jsonb_build_object(
    'path_prefix', COALESCE(reglas->>'path_prefix', ''),
    'routes', jsonb_build_array(
        jsonb_build_object(
            'method', '*',
            'prefix', COALESCE(reglas->>'path_prefix', '/'),
            'accion', accion
        ),
        jsonb_build_object(
            'method', 'GET',
            'path', COALESCE(reglas->>'path_prefix', '/'),
            'accion', 'consultar'
        )
    )
)
WHERE NOT (reglas ? 'routes') AND reglas ? 'path_prefix';

-- 6. Seed / Upsert granular policies for canonical profiles
INSERT INTO sesion.politica (nombre, descripcion, perfil, recurso, accion, reglas, razon_i) VALUES 
('policy-superadmin-all-routes', 'Acceso global superadmin a todas las rutas', 'superadmin', 'todas_las_rutas', 'consultar', '{"path_prefix":"/","routes":[{"method":"*","prefix":"/","accion":"consultar"}]}'::jsonb, 'ABAC Granular'),
('policy-superadmin-landing-menu', 'Acceso a rutas de navegación superadmin', 'superadmin', 'todas_las_rutas', 'consultar', '{"path_prefix":"/","routes":[{"method":"GET","path":"/dashboard","accion":"consultar"},{"method":"GET","path":"/institucion/personas","accion":"consultar"},{"method":"GET","path":"/configuracion/usuarios","accion":"consultar"},{"method":"GET","path":"/configuracion/perfiles","accion":"consultar"},{"method":"GET","path":"/configuracion/politicas","accion":"consultar"},{"method":"GET","path":"/parametros","accion":"consultar"},{"method":"GET","path":"/crud","accion":"consultar"}]}'::jsonb, 'ABAC Granular'),

('policy-admin-dashboard', 'Acceso admin al dashboard', 'admin', 'dashboard', 'consultar', '{"path_prefix":"/dashboard","routes":[{"method":"*","prefix":"/dashboard","accion":"consultar"},{"method":"GET","path":"/dashboard","accion":"consultar"}]}'::jsonb, 'ABAC Granular'),
('policy-admin-configuracion', 'Acceso admin a configuración', 'admin', 'configuracion', 'consultar', '{"path_prefix":"/configuracion","routes":[{"method":"*","prefix":"/configuracion","accion":"consultar"}]}'::jsonb, 'ABAC Granular'),
('policy-admin-institucion', 'Acceso admin a módulo de institución', 'admin', 'institucion', 'consultar', '{"path_prefix":"/institucion","routes":[{"method":"*","prefix":"/institucion","accion":"consultar"}]}'::jsonb, 'ABAC Granular'),
('policy-admin-parametros', 'Acceso admin a parámetros del sistema', 'admin', 'parametros', 'consultar', '{"path_prefix":"/parametros","routes":[{"method":"*","prefix":"/parametros","accion":"consultar"}]}'::jsonb, 'ABAC Granular'),
('policy-admin-crud', 'Acceso admin a módulo CRUD de ejemplo', 'admin', 'crud', 'consultar', '{"path_prefix":"/crud","routes":[{"method":"*","prefix":"/crud","accion":"consultar"},{"method":"GET","path":"/crud","accion":"consultar"}]}'::jsonb, 'ABAC Granular'),

('policy-operativo-dashboard', 'Acceso operativo al dashboard', 'operativo', 'dashboard', 'consultar', '{"path_prefix":"/dashboard","routes":[{"method":"GET","path":"/dashboard","accion":"consultar"},{"method":"*","prefix":"/dashboard","accion":"consultar"}]}'::jsonb, 'ABAC Granular'),
('policy-ejecutivo-dashboard', 'Acceso ejecutivo al dashboard', 'ejecutivo', 'dashboard', 'consultar', '{"path_prefix":"/dashboard","routes":[{"method":"GET","path":"/dashboard","accion":"consultar"},{"method":"*","prefix":"/dashboard","accion":"consultar"}]}'::jsonb, 'ABAC Granular')
ON CONFLICT (nombre) DO UPDATE SET
    descripcion = EXCLUDED.descripcion,
    perfil = EXCLUDED.perfil,
    recurso = EXCLUDED.recurso,
    accion = EXCLUDED.accion,
    reglas = EXCLUDED.reglas,
    razon_m = 'ABAC Granular',
    fecha_m = NOW();
