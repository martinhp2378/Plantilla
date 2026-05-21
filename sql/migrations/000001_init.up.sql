-- 000001_init.up.sql

-- 1. SCHEMAS
CREATE SCHEMA IF NOT EXISTS institucion;
CREATE SCHEMA IF NOT EXISTS sesion;
CREATE SCHEMA IF NOT EXISTS sistema;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 2. AUDIT FUNCTIONS
CREATE OR REPLACE FUNCTION sistema.fn_set_fecha_m()
RETURNS TRIGGER AS $$
BEGIN
    NEW.fecha_m = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

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

-- 3. TABLES - INSTITUCION
CREATE TABLE institucion.cat_tenants (
    tenant_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subdomain    VARCHAR(50) UNIQUE NOT NULL,
    schema_name  VARCHAR(50) UNIQUE NOT NULL,
    nombre_depto VARCHAR(100) NOT NULL,
    is_active    BOOLEAN DEFAULT true,
    fecha_i      TIMESTAMPTZ DEFAULT NOW(),
    fecha_m      TIMESTAMPTZ
);

CREATE TABLE institucion.persona (
    uuid_persona  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre        VARCHAR(100) NOT NULL,
    apellido1     VARCHAR(100), 
    apellido2     VARCHAR(100),
    clave_issemym VARCHAR(15)  NOT NULL UNIQUE,
    curp          VARCHAR(18)  NOT NULL UNIQUE,
    rfc           VARCHAR(13)  NOT NULL UNIQUE,
    imagen_perfil VARCHAR(40)  DEFAULT 'placeholder.png',
    razon_i       VARCHAR(40), 
    razon_m       VARCHAR(40),
    fecha_i       TIMESTAMPTZ DEFAULT NOW(), 
    fecha_m       TIMESTAMPTZ
);

-- 4. TABLES - SESION
CREATE TABLE sesion.cat_accion (
    nombre VARCHAR(50) PRIMARY KEY, 
    razon_i VARCHAR(40), razon_m VARCHAR(40), fecha_i TIMESTAMPTZ DEFAULT NOW(), fecha_m TIMESTAMPTZ
);

CREATE TABLE sesion.recurso (
    nombre VARCHAR(50) PRIMARY KEY, descripcion TEXT, 
    razon_i VARCHAR(40), razon_m VARCHAR(40), fecha_i TIMESTAMPTZ DEFAULT NOW(), fecha_m TIMESTAMPTZ
);

CREATE TABLE sesion.perfil (
    nombre VARCHAR(50) PRIMARY KEY,
    razon_i VARCHAR(40), razon_m VARCHAR(40), fecha_i TIMESTAMPTZ DEFAULT NOW(), fecha_m TIMESTAMPTZ
);

CREATE TABLE sesion.prelogin (
    uuid_persona   UUID PRIMARY KEY REFERENCES institucion.persona(uuid_persona),
    uuid_sustituto UUID,
    razon_i        VARCHAR(40), 
    razon_m        VARCHAR(40), 
    fecha_i        TIMESTAMPTZ DEFAULT NOW(), 
    fecha_m        TIMESTAMPTZ
);

CREATE TABLE sesion.politica (
    nombre VARCHAR(150) PRIMARY KEY, descripcion TEXT DEFAULT '',
    perfil VARCHAR(50) NOT NULL REFERENCES sesion.perfil(nombre),
    recurso VARCHAR(50) NOT NULL REFERENCES sesion.recurso(nombre),
    accion VARCHAR(50) NOT NULL REFERENCES sesion.cat_accion(nombre),
    reglas JSONB NOT NULL,
    razon_i VARCHAR(40), razon_m VARCHAR(40), fecha_i TIMESTAMPTZ DEFAULT NOW(), fecha_m TIMESTAMPTZ
);

CREATE TABLE sesion.usuario (
    uuid_persona UUID PRIMARY KEY REFERENCES institucion.persona(uuid_persona),
    email VARCHAR(200) UNIQUE NOT NULL, 
    password TEXT NOT NULL,
    perfil VARCHAR(50) DEFAULT 'paciente' REFERENCES sesion.perfil(nombre),
    razon_i VARCHAR(40), razon_m VARCHAR(40), fecha_i TIMESTAMPTZ DEFAULT NOW(), fecha_m TIMESTAMPTZ
);

CREATE TABLE sesion.sesion (
    uuid_sesion           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uuid_persona          UUID REFERENCES institucion.persona(uuid_persona),
    tipo_sesion           VARCHAR(20) NOT NULL DEFAULT 'cookie' CHECK (tipo_sesion IN ('cookie', 'jwt')),
    refresh_token_hash    BYTEA,
    mapa_structs          JSONB NOT NULL,
    perfil                VARCHAR(50) NOT NULL REFERENCES sesion.perfil(nombre),
    fecha_creacion        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    fecha_actualizacion   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    fecha_ultimo_acceso   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    fecha_expiracion      TIMESTAMPTZ,
    direccion_ip          VARCHAR(45), 
    agente_usuario        TEXT,
    mensajes_flash        JSONB[], 
    huella                VARCHAR(36) NOT NULL DEFAULT gen_random_uuid(),
    info_dispositivo      JSONB DEFAULT '{}'::JSONB,
    recuerdame            BOOL DEFAULT FALSE, 
    total_peticiones      INT DEFAULT 0,
    fecha_ultima_peticion TIMESTAMPTZ DEFAULT NOW(),
    razon_i               VARCHAR(40), 
    razon_m               VARCHAR(40), 
    fecha_i               TIMESTAMPTZ DEFAULT NOW(), 
    fecha_m               TIMESTAMPTZ
);

-- 5. AUDIT - BITACORA
CREATE TABLE sistema.bitacora_cambios (
    uuid_bitacora UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uuid_persona UUID,
    accion VARCHAR(10) NOT NULL,
    nombre_tabla VARCHAR(100) NOT NULL,
    registro_pk TEXT NOT NULL,
    datos_viejos JSONB,
    datos_nuevos JSONB,
    razon TEXT,
    fecha_i TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE sistema.cfg_parametros (
    parametro_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clave          VARCHAR(100) UNIQUE NOT NULL, 
    valor          TEXT NOT NULL, 
    descripcion    TEXT,
    tipo_dato      VARCHAR(20) NOT NULL CHECK (tipo_dato IN ('STRING', 'INT', 'BOOLEAN', 'JSON')),
    is_active      BOOLEAN NOT NULL DEFAULT true,
    razon_i        VARCHAR(40), 
    razon_m        VARCHAR(40), 
    fecha_i        TIMESTAMPTZ DEFAULT NOW(), 
    fecha_m        TIMESTAMPTZ
);

-- 6. TRIGGERS - AUTO UPDATED AT (fecha_m)
CREATE TRIGGER tr_set_fecha_m_tenants BEFORE UPDATE ON institucion.cat_tenants FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();
CREATE TRIGGER tr_set_fecha_m_persona BEFORE UPDATE ON institucion.persona FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();
CREATE TRIGGER tr_set_fecha_m_cat_accion BEFORE UPDATE ON sesion.cat_accion FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();
CREATE TRIGGER tr_set_fecha_m_recurso BEFORE UPDATE ON sesion.recurso FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();
CREATE TRIGGER tr_set_fecha_m_perfil BEFORE UPDATE ON sesion.perfil FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();
CREATE TRIGGER tr_set_fecha_m_prelogin BEFORE UPDATE ON sesion.prelogin FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();
CREATE TRIGGER tr_set_fecha_m_politica BEFORE UPDATE ON sesion.politica FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();
CREATE TRIGGER tr_set_fecha_m_usuario BEFORE UPDATE ON sesion.usuario FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();
CREATE TRIGGER tr_set_fecha_m_sesion BEFORE UPDATE ON sesion.sesion FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();
CREATE TRIGGER tr_set_fecha_m_parametros BEFORE UPDATE ON sistema.cfg_parametros FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();

-- 7. TRIGGERS - AUDIT (bitacora_cambios)
CREATE TRIGGER tr_auditoria_persona AFTER INSERT OR UPDATE OR DELETE ON institucion.persona FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('uuid_persona');
CREATE TRIGGER tr_auditoria_usuario AFTER INSERT OR UPDATE OR DELETE ON sesion.usuario FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('uuid_persona');
CREATE TRIGGER tr_auditoria_perfil AFTER INSERT OR UPDATE OR DELETE ON sesion.perfil FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('nombre');
CREATE TRIGGER tr_auditoria_politica AFTER INSERT OR UPDATE OR DELETE ON sesion.politica FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('nombre');
CREATE TRIGGER tr_auditoria_tenants AFTER INSERT OR UPDATE OR DELETE ON institucion.cat_tenants FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('tenant_id');
CREATE TRIGGER tr_auditoria_parametros AFTER INSERT OR UPDATE OR DELETE ON sistema.cfg_parametros FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('parametro_id');
