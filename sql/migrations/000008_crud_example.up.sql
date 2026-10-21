-- 000008_crud_example.up.sql
-- Módulo y tabla de ejemplo CRUD con uno de cada tipo de datos representativo

CREATE TABLE IF NOT EXISTS institucion.crud (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    texto_corto VARCHAR(150) NOT NULL,
    texto_largo TEXT,
    entero      INTEGER NOT NULL DEFAULT 0,
    decimal     NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    booleano    BOOLEAN NOT NULL DEFAULT true,
    fecha       DATE NOT NULL,
    hora        TIME NOT NULL,
    fecha_hora  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    seleccion   VARCHAR(50) NOT NULL,
    correo      VARCHAR(150) NOT NULL,
    telefono    VARCHAR(20),
    url         VARCHAR(255),
    color       VARCHAR(7) DEFAULT '#691C32',
    tenant_id   UUID DEFAULT NULLIF(current_setting('app.tenant_id', true), '')::UUID,
    razon_i     VARCHAR(40),
    razon_m     VARCHAR(40),
    fecha_i     TIMESTAMPTZ DEFAULT NOW(),
    fecha_m     TIMESTAMPTZ
);

-- Triggers de auditoría y fecha de modificación
CREATE TRIGGER tr_set_fecha_m_crud
BEFORE UPDATE ON institucion.crud
FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();

CREATE TRIGGER tr_auditoria_crud
AFTER INSERT OR UPDATE OR DELETE ON institucion.crud
FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('id');

-- Índice de aislamiento multi-tenant
CREATE INDEX IF NOT EXISTS idx_crud_tenant_id ON institucion.crud(tenant_id);

-- Row Level Security (RLS) mandatorio
ALTER TABLE institucion.crud ENABLE ROW LEVEL SECURITY;
ALTER TABLE institucion.crud FORCE ROW LEVEL SECURITY;

-- Política de aislamiento por tenant
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = 'institucion' AND tablename = 'crud' AND policyname = 'tenant_isolation_crud'
    ) THEN
        CREATE POLICY tenant_isolation_crud ON institucion.crud
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true))
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
    END IF;
END $$;

-- Permisos para el rol de ejecución en runtime si existe
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_runtime_tenant') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON institucion.crud TO app_runtime_tenant;
    END IF;
END $$;

-- Registrar recurso y política de acceso para superadmin
INSERT INTO sesion.recurso (nombre, descripcion, razon_i) VALUES
('crud', 'Módulo de ejemplo CRUD con todos los tipos de datos', 'Carga Inicial')
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO sesion.politica (nombre, descripcion, perfil, recurso, accion, reglas, razon_i) VALUES
('Acceso CRUD', 'Ejemplo de CRUD completo', 'superadmin', 'crud', 'consultar', '{"path_prefix": "/crud"}', 'Carga Inicial')
ON CONFLICT (nombre) DO NOTHING;

-- Registro de demostración inicial para el tenant administrador por defecto
DO $$
DECLARE
    v_default_tenant_id UUID;
BEGIN
    SELECT tenant_id INTO v_default_tenant_id FROM institucion.cat_tenants WHERE subdomain = 'admin' LIMIT 1;
    IF v_default_tenant_id IS NOT NULL THEN
        INSERT INTO institucion.crud (
            id,
            texto_corto,
            texto_largo,
            entero,
            decimal,
            booleano,
            fecha,
            hora,
            fecha_hora,
            seleccion,
            correo,
            telefono,
            url,
            color,
            tenant_id,
            razon_i
        ) VALUES (
            gen_random_uuid(),
            'Registro de Ejemplo Inicial',
            'Este es un texto largo explicativo que sirve como muestra de un campo TEXT en la plantilla institucional.',
            42,
            1250.75,
            true,
            CURRENT_DATE,
            '12:30:00',
            NOW(),
            'opcion_a',
            'contacto.ejemplo@issemym.gob.mx',
            '7221234567',
            'https://issemym.gob.mx',
            '#691C32',
            v_default_tenant_id,
            'Carga Inicial'
        ) ON CONFLICT DO NOTHING;
    END IF;
END $$;
