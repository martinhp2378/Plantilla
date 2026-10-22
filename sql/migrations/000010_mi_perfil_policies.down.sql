-- Rollback migration 000010: ABAC policies for Mi Perfil
DELETE FROM sesion.politica 
WHERE nombre IN (
    'policy-superadmin-mi-perfil',
    'policy-admin-mi-perfil',
    'policy-operativo-mi-perfil',
    'policy-ejecutivo-mi-perfil',
    'policy-coordinacion-mi-perfil'
);

DELETE FROM sesion.recurso WHERE nombre = 'mi_perfil';
