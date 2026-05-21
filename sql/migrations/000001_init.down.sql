-- 000001_init.down.sql

DROP TABLE IF EXISTS sistema.bitacora_cambios;
DROP TABLE IF EXISTS sesion.sesion;
DROP TABLE IF EXISTS sesion.usuario;
DROP TABLE IF EXISTS sesion.prelogin;
DROP TABLE IF EXISTS sesion.politica;
DROP TABLE IF EXISTS sesion.perfil;
DROP TABLE IF EXISTS sesion.recurso;
DROP TABLE IF EXISTS sesion.cat_accion;
DROP TABLE IF EXISTS institucion.persona;
DROP TABLE IF EXISTS institucion.cat_tenants;

DROP FUNCTION IF EXISTS sistema.fn_registrar_auditoria;
DROP FUNCTION IF EXISTS sistema.fn_set_fecha_m;

DROP SCHEMA IF EXISTS sistema CASCADE;
DROP SCHEMA IF EXISTS sesion CASCADE;
DROP SCHEMA IF EXISTS institucion CASCADE;
