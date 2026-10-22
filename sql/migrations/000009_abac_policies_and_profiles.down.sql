-- 000009_abac_policies_and_profiles.down.sql

DELETE FROM sesion.politica 
WHERE nombre IN (
    'policy-superadmin-all-routes',
    'policy-admin-dashboard',
    'policy-admin-configuracion',
    'policy-admin-institucion',
    'policy-admin-parametros',
    'policy-admin-crud',
    'policy-operativo-dashboard',
    'policy-ejecutivo-dashboard'
);

ALTER TABLE sesion.usuario 
ALTER COLUMN perfil SET DEFAULT 'paciente';

DELETE FROM sesion.perfil
WHERE nombre IN ('admin', 'operativo', 'coordinacion', 'medico', 'recepcion', 'adminclinica');
