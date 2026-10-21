-- 000008_crud_example.down.sql

DROP TRIGGER IF EXISTS tr_auditoria_crud ON institucion.crud;
DROP TRIGGER IF EXISTS tr_set_fecha_m_crud ON institucion.crud;
DROP POLICY IF EXISTS tenant_isolation_crud ON institucion.crud;
DROP TABLE IF EXISTS institucion.crud CASCADE;

DELETE FROM sesion.politica WHERE nombre = 'Acceso CRUD';
DELETE FROM sesion.recurso WHERE nombre = 'crud';
