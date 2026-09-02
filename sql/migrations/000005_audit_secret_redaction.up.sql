-- 000005_audit_secret_redaction.up.sql
-- Remove credential and configuration secrets from new and historical audit snapshots.

CREATE OR REPLACE FUNCTION sistema.fn_registrar_auditoria()
RETURNS TRIGGER AS $$
DECLARE
    v_uuid_persona UUID;
    v_tenant_id UUID;
    v_registro_pk TEXT;
    v_razon TEXT;
    v_datos_viejos JSONB;
    v_datos_nuevos JSONB;
BEGIN
    BEGIN
        v_uuid_persona := current_setting('session_variables.uuid_persona', true)::UUID;
    EXCEPTION WHEN OTHERS THEN
        v_uuid_persona := NULL;
    END;

    BEGIN
        v_tenant_id := NULLIF(current_setting('app.tenant_id', true), '')::UUID;
    EXCEPTION WHEN OTHERS THEN
        v_tenant_id := NULL;
    END;

    IF (TG_OP = 'INSERT') THEN
        EXECUTE format('SELECT $1.%I::text', TG_ARGV[0]) INTO v_registro_pk USING NEW;
        v_datos_nuevos := to_jsonb(NEW);
        IF TG_TABLE_SCHEMA = 'sesion' AND TG_TABLE_NAME = 'usuario' THEN
            v_datos_nuevos := v_datos_nuevos - 'password';
        ELSIF TG_TABLE_SCHEMA = 'sistema' AND TG_TABLE_NAME = 'cfg_parametros' THEN
            v_datos_nuevos := v_datos_nuevos - 'valor';
        END IF;
        v_razon := v_datos_nuevos ->> 'razon_i';
        INSERT INTO sistema.bitacora_cambios (tenant_id, uuid_persona, accion, nombre_tabla, registro_pk, datos_nuevos, razon)
        VALUES (v_tenant_id, v_uuid_persona, 'INSERT', TG_TABLE_NAME, v_registro_pk, v_datos_nuevos, v_razon);
        RETURN NEW;

    ELSIF (TG_OP = 'UPDATE') THEN
        EXECUTE format('SELECT $1.%I::text', TG_ARGV[0]) INTO v_registro_pk USING NEW;
        v_datos_viejos := to_jsonb(OLD);
        v_datos_nuevos := to_jsonb(NEW);
        IF TG_TABLE_SCHEMA = 'sesion' AND TG_TABLE_NAME = 'usuario' THEN
            v_datos_viejos := v_datos_viejos - 'password';
            v_datos_nuevos := v_datos_nuevos - 'password';
        ELSIF TG_TABLE_SCHEMA = 'sistema' AND TG_TABLE_NAME = 'cfg_parametros' THEN
            v_datos_viejos := v_datos_viejos - 'valor';
            v_datos_nuevos := v_datos_nuevos - 'valor';
        END IF;
        v_razon := v_datos_nuevos ->> 'razon_m';
        INSERT INTO sistema.bitacora_cambios (tenant_id, uuid_persona, accion, nombre_tabla, registro_pk, datos_viejos, datos_nuevos, razon)
        VALUES (v_tenant_id, v_uuid_persona, 'UPDATE', TG_TABLE_NAME, v_registro_pk, v_datos_viejos, v_datos_nuevos, v_razon);
        RETURN NEW;

    ELSIF (TG_OP = 'DELETE') THEN
        EXECUTE format('SELECT $1.%I::text', TG_ARGV[0]) INTO v_registro_pk USING OLD;
        v_datos_viejos := to_jsonb(OLD);
        IF TG_TABLE_SCHEMA = 'sesion' AND TG_TABLE_NAME = 'usuario' THEN
            v_datos_viejos := v_datos_viejos - 'password';
        ELSIF TG_TABLE_SCHEMA = 'sistema' AND TG_TABLE_NAME = 'cfg_parametros' THEN
            v_datos_viejos := v_datos_viejos - 'valor';
        END IF;
        INSERT INTO sistema.bitacora_cambios (tenant_id, uuid_persona, accion, nombre_tabla, registro_pk, datos_viejos, razon)
        VALUES (v_tenant_id, v_uuid_persona, 'DELETE', TG_TABLE_NAME, v_registro_pk, v_datos_viejos, 'DELETE operation');
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- The exclusive lock prevents writes while the table owner temporarily leaves
-- FORCE mode to scrub every tenant in one migration. RLS remains enabled for
-- non-owner roles, and FORCE mode is restored before the migration completes.
LOCK TABLE sistema.bitacora_cambios IN ACCESS EXCLUSIVE MODE;
ALTER TABLE sistema.bitacora_cambios NO FORCE ROW LEVEL SECURITY;

UPDATE sistema.bitacora_cambios
SET datos_viejos = CASE
        WHEN nombre_tabla = 'usuario' THEN datos_viejos - 'password'
        WHEN nombre_tabla = 'cfg_parametros' THEN datos_viejos - 'valor'
        ELSE datos_viejos
    END,
    datos_nuevos = CASE
        WHEN nombre_tabla = 'usuario' THEN datos_nuevos - 'password'
        WHEN nombre_tabla = 'cfg_parametros' THEN datos_nuevos - 'valor'
        ELSE datos_nuevos
    END
WHERE (nombre_tabla = 'usuario' AND
       (COALESCE(datos_viejos, '{}'::jsonb) ? 'password' OR COALESCE(datos_nuevos, '{}'::jsonb) ? 'password'))
   OR (nombre_tabla = 'cfg_parametros' AND
       (COALESCE(datos_viejos, '{}'::jsonb) ? 'valor' OR COALESCE(datos_nuevos, '{}'::jsonb) ? 'valor'));

ALTER TABLE sistema.bitacora_cambios FORCE ROW LEVEL SECURITY;
