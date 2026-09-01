-- 000002_seed.up.sql

SET search_path TO institucion, sesion, sistema, public;

-- 1. ACCIONES
INSERT INTO cat_accion (nombre, razon_i) VALUES 
('buscar', 'Carga Inicial'), 
('consultar', 'Carga Inicial'), 
('modificar', 'Carga Inicial'), 
('crear', 'Carga Inicial'), 
('eliminar', 'Carga Inicial'),
('aprobar', 'Carga Inicial')
ON CONFLICT DO NOTHING;

-- 2. RECURSOS
INSERT INTO recurso (nombre, descripcion, razon_i) VALUES 
('usuarios', 'Gestión de Usuarios y Perfiles', 'Carga Inicial'),
('dashboard', 'Acceso al dashboard administrativo', 'Carga Inicial'),
('configuracion', 'Acceso a módulos de configuración', 'Carga Inicial'),
('institucion', 'Acceso a módulos de institución', 'Carga Inicial'),
('superadmin', 'Acceso total a funciones de super administrador', 'Carga Inicial'),
('parametros', 'Gestión de parámetros del sistema', 'Carga Inicial')
ON CONFLICT DO NOTHING;

-- 3. PARAMETROS INICIALES
INSERT INTO sistema.cfg_parametros (clave, valor, descripcion, tipo_dato, razon_i) VALUES 
('EMAIL_SYSTEM_NAME', 'Plantilla Institucional', 'Nombre del sistema que aparece en los correos', 'STRING', 'Carga Inicial'),
('EMAIL_BRAND_NAME', 'ISSEMYM', 'Marca institucional', 'STRING', 'Carga Inicial'),
('EMAIL_FOOTER_ORGANIZATION', 'ISSEMYM - Unidad de Tecnologías', 'Organización en el pie de página del correo', 'STRING', 'Carga Inicial'),
('EMAIL_SUPPORT_LABEL', '¿Necesitas ayuda?', 'Etiqueta de soporte', 'STRING', 'Carga Inicial'),
('EMAIL_SUPPORT_PHONE', '722 226 1900 ext. 1911', 'Teléfono de soporte', 'STRING', 'Carga Inicial')
ON CONFLICT DO NOTHING;

-- 4. PERFILES
INSERT INTO perfil (nombre, razon_i) VALUES 
('superadmin', 'Administrador Total del Sistema'), 
('operador', 'Personal de Unidad Médica'), 
('ejecutivo', 'Consulta de Indicadores'),
('anonimo', 'Perfil para usuarios no autenticados')
ON CONFLICT DO NOTHING;

-- 4. POLITICAS INICIALES (Para que el superadmin pueda entrar y navegar)
INSERT INTO politica (nombre, descripcion, perfil, recurso, accion, reglas, razon_i) VALUES 
('Acceso Usuarios', 'Gestión de usuarios', 'superadmin', 'usuarios', 'consultar', '{"path_prefix": "/configuracion/usuarios"}', 'Carga Inicial'),
('Acceso Perfiles', 'Gestión de perfiles', 'superadmin', 'superadmin', 'consultar', '{"path_prefix": "/configuracion/perfiles"}', 'Carga Inicial'),
('Acceso Politicas', 'Gestión de políticas', 'superadmin', 'superadmin', 'consultar', '{"path_prefix": "/configuracion/politicas"}', 'Carga Inicial'),
('Acceso Dashboard', 'Acceso al dashboard', 'superadmin', 'dashboard', 'consultar', '{"path_prefix": "/dashboard"}', 'Carga Inicial'),
('Acceso Personas', 'Catálogo de personas', 'superadmin', 'institucion', 'consultar', '{"path_prefix": "/institucion/personas"}', 'Carga Inicial')
ON CONFLICT DO NOTHING;

-- 5. TENANT INICIAL
INSERT INTO cat_tenants (subdomain, schema_name, nombre_depto, is_active) VALUES
('admin', 'public', 'Administración Central', true)
ON CONFLICT DO NOTHING;
