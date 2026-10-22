-- Migration 000010: ABAC policies for Mi Perfil and extend avatar filename
-- Author: Antigravity
-- Date: 2026-10-21

ALTER TABLE institucion.persona ALTER COLUMN imagen_perfil TYPE VARCHAR(100);

INSERT INTO sesion.recurso (nombre, descripcion, razon_i)
VALUES ('mi_perfil', 'Módulo de Perfil de Usuario', 'Mi Perfil ABAC')
ON CONFLICT (nombre) DO UPDATE SET
    descripcion = EXCLUDED.descripcion,
    razon_m = 'Mi Perfil ABAC',
    fecha_m = now();

INSERT INTO sesion.politica (nombre, descripcion, perfil, recurso, accion, reglas, razon_i)
VALUES 
    ('policy-superadmin-mi-perfil', 'Acceso a mi perfil para superadmin', 'superadmin', 'mi_perfil', 'consultar', '{"path_prefix":"/perfil","routes":[{"method":"*","prefix":"/perfil","accion":"consultar"}]}'::jsonb, 'Mi Perfil ABAC'),
    ('policy-admin-mi-perfil', 'Acceso a mi perfil para admin', 'admin', 'mi_perfil', 'consultar', '{"path_prefix":"/perfil","routes":[{"method":"*","prefix":"/perfil","accion":"consultar"}]}'::jsonb, 'Mi Perfil ABAC'),
    ('policy-operativo-mi-perfil', 'Acceso a mi perfil para operativo', 'operativo', 'mi_perfil', 'consultar', '{"path_prefix":"/perfil","routes":[{"method":"*","prefix":"/perfil","accion":"consultar"}]}'::jsonb, 'Mi Perfil ABAC'),
    ('policy-ejecutivo-mi-perfil', 'Acceso a mi perfil para ejecutivo', 'ejecutivo', 'mi_perfil', 'consultar', '{"path_prefix":"/perfil","routes":[{"method":"*","prefix":"/perfil","accion":"consultar"}]}'::jsonb, 'Mi Perfil ABAC'),
    ('policy-coordinacion-mi-perfil', 'Acceso a mi perfil para coordinacion', 'coordinacion', 'mi_perfil', 'consultar', '{"path_prefix":"/perfil","routes":[{"method":"*","prefix":"/perfil","accion":"consultar"}]}'::jsonb, 'Mi Perfil ABAC')
ON CONFLICT (nombre) DO UPDATE SET
    descripcion = EXCLUDED.descripcion,
    perfil = EXCLUDED.perfil,
    recurso = EXCLUDED.recurso,
    accion = EXCLUDED.accion,
    reglas = EXCLUDED.reglas,
    razon_m = 'Mi Perfil ABAC',
    fecha_m = now();
