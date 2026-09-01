-- 000003_tenant_isolation_expand.up.sql
-- Expand phase of tenant isolation: add tenant_id columns, indexes, FORCE RLS, and isolation policies.

-- 1. Add tenant_id columns to tenant-scoped tables
ALTER TABLE institucion.persona ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE sesion.usuario ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE sesion.sesion ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE sesion.prelogin ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE sistema.bitacora_cambios ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE sistema.cfg_parametros ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- 2. Backfill existing seed data with initial tenant ID if available
DO $$
DECLARE
    v_default_tenant_id UUID;
BEGIN
    SELECT tenant_id INTO v_default_tenant_id FROM institucion.cat_tenants WHERE subdomain = 'admin' LIMIT 1;
    IF v_default_tenant_id IS NOT NULL THEN
        UPDATE institucion.persona SET tenant_id = v_default_tenant_id WHERE tenant_id IS NULL;
        UPDATE sesion.usuario SET tenant_id = v_default_tenant_id WHERE tenant_id IS NULL;
        UPDATE sesion.sesion SET tenant_id = v_default_tenant_id WHERE tenant_id IS NULL;
        UPDATE sesion.prelogin SET tenant_id = v_default_tenant_id WHERE tenant_id IS NULL;
        UPDATE sistema.bitacora_cambios SET tenant_id = v_default_tenant_id WHERE tenant_id IS NULL;
        UPDATE sistema.cfg_parametros SET tenant_id = v_default_tenant_id WHERE tenant_id IS NULL;
    END IF;
END $$;

-- 3. Create indexes on tenant_id
CREATE INDEX IF NOT EXISTS idx_persona_tenant_id ON institucion.persona(tenant_id);
CREATE INDEX IF NOT EXISTS idx_usuario_tenant_id ON sesion.usuario(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sesion_tenant_id ON sesion.sesion(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bitacora_cambios_tenant_id ON sistema.bitacora_cambios(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cfg_parametros_tenant_id ON sistema.cfg_parametros(tenant_id);

-- 4. Enable and FORCE Row Level Security (RLS) on tenant-scoped tables
ALTER TABLE institucion.persona ENABLE ROW LEVEL SECURITY;
ALTER TABLE institucion.persona FORCE ROW LEVEL SECURITY;

ALTER TABLE sesion.usuario ENABLE ROW LEVEL SECURITY;
ALTER TABLE sesion.usuario FORCE ROW LEVEL SECURITY;

ALTER TABLE sesion.sesion ENABLE ROW LEVEL SECURITY;
ALTER TABLE sesion.sesion FORCE ROW LEVEL SECURITY;

ALTER TABLE sesion.prelogin ENABLE ROW LEVEL SECURITY;
ALTER TABLE sesion.prelogin FORCE ROW LEVEL SECURITY;

ALTER TABLE sistema.bitacora_cambios ENABLE ROW LEVEL SECURITY;
ALTER TABLE sistema.bitacora_cambios FORCE ROW LEVEL SECURITY;

ALTER TABLE sistema.cfg_parametros ENABLE ROW LEVEL SECURITY;
ALTER TABLE sistema.cfg_parametros FORCE ROW LEVEL SECURITY;

-- 5. Record the transaction-local tenant identity in audit rows. An absent or
-- invalid setting becomes NULL, which the FORCE RLS WITH CHECK policy rejects.
CREATE OR REPLACE FUNCTION sistema.fn_registrar_auditoria()
RETURNS TRIGGER AS $$
DECLARE
    v_uuid_persona UUID;
    v_tenant_id UUID;
    v_registro_pk TEXT;
    v_razon TEXT;
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
        v_razon := (to_jsonb(NEW) ->> 'razon_i');
        INSERT INTO sistema.bitacora_cambios (tenant_id, uuid_persona, accion, nombre_tabla, registro_pk, datos_nuevos, razon)
        VALUES (v_tenant_id, v_uuid_persona, 'INSERT', TG_TABLE_NAME, v_registro_pk, to_jsonb(NEW), v_razon);
        RETURN NEW;

    ELSIF (TG_OP = 'UPDATE') THEN
        EXECUTE format('SELECT $1.%I::text', TG_ARGV[0]) INTO v_registro_pk USING NEW;
        v_razon := (to_jsonb(NEW) ->> 'razon_m');
        INSERT INTO sistema.bitacora_cambios (tenant_id, uuid_persona, accion, nombre_tabla, registro_pk, datos_viejos, datos_nuevos, razon)
        VALUES (v_tenant_id, v_uuid_persona, 'UPDATE', TG_TABLE_NAME, v_registro_pk, to_jsonb(OLD), to_jsonb(NEW), v_razon);
        RETURN NEW;

    ELSIF (TG_OP = 'DELETE') THEN
        EXECUTE format('SELECT $1.%I::text', TG_ARGV[0]) INTO v_registro_pk USING OLD;
        INSERT INTO sistema.bitacora_cambios (tenant_id, uuid_persona, accion, nombre_tabla, registro_pk, datos_viejos, razon)
        VALUES (v_tenant_id, v_uuid_persona, 'DELETE', TG_TABLE_NAME, v_registro_pk, to_jsonb(OLD), 'DELETE operation');
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 6. Create tenant isolation policies
DO $$
BEGIN
    -- institucion.persona
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = 'institucion' AND tablename = 'persona' AND policyname = 'tenant_isolation_persona'
    ) THEN
        CREATE POLICY tenant_isolation_persona ON institucion.persona
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true))
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
    END IF;

    -- sesion.usuario
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = 'sesion' AND tablename = 'usuario' AND policyname = 'tenant_isolation_usuario'
    ) THEN
        CREATE POLICY tenant_isolation_usuario ON sesion.usuario
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true))
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
    END IF;

    -- sesion.sesion
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = 'sesion' AND tablename = 'sesion' AND policyname = 'tenant_isolation_sesion'
    ) THEN
        CREATE POLICY tenant_isolation_sesion ON sesion.sesion
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true))
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
    END IF;

    -- sesion.prelogin
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = 'sesion' AND tablename = 'prelogin' AND policyname = 'tenant_isolation_prelogin'
    ) THEN
        CREATE POLICY tenant_isolation_prelogin ON sesion.prelogin
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true))
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
    END IF;

    -- sistema.bitacora_cambios
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = 'sistema' AND tablename = 'bitacora_cambios' AND policyname = 'tenant_isolation_bitacora'
    ) THEN
        CREATE POLICY tenant_isolation_bitacora ON sistema.bitacora_cambios
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true))
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
    END IF;

    -- sistema.cfg_parametros
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = 'sistema' AND tablename = 'cfg_parametros' AND policyname = 'tenant_isolation_parametros'
    ) THEN
        CREATE POLICY tenant_isolation_parametros ON sistema.cfg_parametros
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true))
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
    END IF;
END $$;
