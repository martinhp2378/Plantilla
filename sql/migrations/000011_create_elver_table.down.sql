-- 000011_create_elver_table.down.sql

DELETE FROM sesion.politica WHERE recurso = 'elver';
DELETE FROM sesion.recurso WHERE nombre = 'elver';

DROP TRIGGER IF EXISTS tr_auditoria_elver ON sistema.elver;
DROP TRIGGER IF EXISTS tr_set_fecha_m_elver ON sistema.elver;
DROP TABLE IF EXISTS sistema.elver;
