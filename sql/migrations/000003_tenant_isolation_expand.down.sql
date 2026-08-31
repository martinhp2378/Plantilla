-- 000003_tenant_isolation_expand.down.sql
-- Non-destructive rollback for expand phase: drop policies and disable RLS without dropping data columns.

DROP POLICY IF EXISTS tenant_isolation_parametros ON sistema.cfg_parametros;
DROP POLICY IF EXISTS tenant_isolation_bitacora ON sistema.bitacora_cambios;
DROP POLICY IF EXISTS tenant_isolation_prelogin ON sesion.prelogin;
DROP POLICY IF EXISTS tenant_isolation_sesion ON sesion.sesion;
DROP POLICY IF EXISTS tenant_isolation_usuario ON sesion.usuario;
DROP POLICY IF EXISTS tenant_isolation_persona ON institucion.persona;

ALTER TABLE sistema.cfg_parametros NO FORCE ROW LEVEL SECURITY;
ALTER TABLE sistema.cfg_parametros DISABLE ROW LEVEL SECURITY;

ALTER TABLE sistema.bitacora_cambios NO FORCE ROW LEVEL SECURITY;
ALTER TABLE sistema.bitacora_cambios DISABLE ROW LEVEL SECURITY;

ALTER TABLE sesion.prelogin NO FORCE ROW LEVEL SECURITY;
ALTER TABLE sesion.prelogin DISABLE ROW LEVEL SECURITY;

ALTER TABLE sesion.sesion NO FORCE ROW LEVEL SECURITY;
ALTER TABLE sesion.sesion DISABLE ROW LEVEL SECURITY;

ALTER TABLE sesion.usuario NO FORCE ROW LEVEL SECURITY;
ALTER TABLE sesion.usuario DISABLE ROW LEVEL SECURITY;

ALTER TABLE institucion.persona NO FORCE ROW LEVEL SECURITY;
ALTER TABLE institucion.persona DISABLE ROW LEVEL SECURITY;

DROP INDEX IF EXISTS sistema.idx_cfg_parametros_tenant_id;
DROP INDEX IF EXISTS sistema.idx_bitacora_cambios_tenant_id;
DROP INDEX IF EXISTS sesion.idx_sesion_tenant_id;
DROP INDEX IF EXISTS sesion.idx_usuario_tenant_id;
DROP INDEX IF EXISTS institucion.idx_persona_tenant_id;
