-- 000003_tenant_isolation_expand.down.sql
-- Non-destructive rollback for expand phase: drop policies and disable RLS without dropping data columns.

CREATE OR REPLACE FUNCTION sistema.fn_registrar_auditoria()
RETURNS TRIGGER AS $$
DECLARE
    v_uuid_persona UUID;
    v_registro_pk TEXT;
    v_razon TEXT;
BEGIN
    BEGIN
        v_uuid_persona := current_setting('session_variables.uuid_persona', true)::UUID;
    EXCEPTION WHEN OTHERS THEN
        v_uuid_persona := NULL;
    END;

    IF (TG_OP = 'INSERT') THEN
        EXECUTE format('SELECT $1.%I::text', TG_ARGV[0]) INTO v_registro_pk USING NEW;
        v_razon := (to_jsonb(NEW) ->> 'razon_i');
        INSERT INTO sistema.bitacora_cambios (uuid_persona, accion, nombre_tabla, registro_pk, datos_nuevos, razon)
        VALUES (v_uuid_persona, 'INSERT', TG_TABLE_NAME, v_registro_pk, to_jsonb(NEW), v_razon);
        RETURN NEW;

    ELSIF (TG_OP = 'UPDATE') THEN
        EXECUTE format('SELECT $1.%I::text', TG_ARGV[0]) INTO v_registro_pk USING NEW;
        v_razon := (to_jsonb(NEW) ->> 'razon_m');
        INSERT INTO sistema.bitacora_cambios (uuid_persona, accion, nombre_tabla, registro_pk, datos_viejos, datos_nuevos, razon)
        VALUES (v_uuid_persona, 'UPDATE', TG_TABLE_NAME, v_registro_pk, to_jsonb(OLD), to_jsonb(NEW), v_razon);
        RETURN NEW;

    ELSIF (TG_OP = 'DELETE') THEN
        EXECUTE format('SELECT $1.%I::text', TG_ARGV[0]) INTO v_registro_pk USING OLD;
        INSERT INTO sistema.bitacora_cambios (uuid_persona, accion, nombre_tabla, registro_pk, datos_viejos, razon)
        VALUES (v_uuid_persona, 'DELETE', TG_TABLE_NAME, v_registro_pk, to_jsonb(OLD), 'DELETE operation');
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

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
