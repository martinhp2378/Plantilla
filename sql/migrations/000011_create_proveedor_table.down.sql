-- 000011_create_proveedor_table.down.sql

DELETE FROM sesion.politica WHERE recurso = 'proveedor';
DELETE FROM sesion.recurso WHERE nombre = 'proveedor';

DROP TRIGGER IF EXISTS tr_auditoria_proveedor ON sistema.proveedor;
DROP TRIGGER IF EXISTS tr_set_fecha_m_proveedor ON sistema.proveedor;
DROP TABLE IF EXISTS sistema.proveedor;
