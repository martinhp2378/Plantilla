-- 000011_create_elver_table.up.sql
-- Tabla y politicas para el modulo Elver

CREATE TABLE IF NOT EXISTS sistema.elver (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre      VARCHAR(150) NOT NULL,
    codigo      VARCHAR(50) NOT NULL,
    descripcion TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    tenant_id   UUID,
    razon_i     VARCHAR(40),
    razon_m     VARCHAR(40),
    fecha_i     TIMESTAMPTZ DEFAULT NOW(),
    fecha_m     TIMESTAMPTZ
);

-- Triggers de auditoria y fecha de modificacion
DROP TRIGGER IF EXISTS tr_set_fecha_m_elver ON sistema.elver;
CREATE TRIGGER tr_set_fecha_m_elver
BEFORE UPDATE ON sistema.elver
FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();

DROP TRIGGER IF EXISTS tr_auditoria_elver ON sistema.elver;
CREATE TRIGGER tr_auditoria_elver
AFTER INSERT OR UPDATE OR DELETE ON sistema.elver
FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('id');

-- Indice de aislamiento multi-tenant
CREATE INDEX IF NOT EXISTS idx_elver_tenant_id ON sistema.elver(tenant_id);

-- Row Level Security (RLS)
ALTER TABLE sistema.elver ENABLE ROW LEVEL SECURITY;
ALTER TABLE sistema.elver FORCE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = 'sistema' AND tablename = 'elver' AND policyname = 'tenant_isolation_elver'
    ) THEN
        CREATE POLICY tenant_isolation_elver ON sistema.elver
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true) OR current_setting('app.tenant_id', true) = 'local')
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true) OR current_setting('app.tenant_id', true) = 'local');
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_runtime_tenant') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON sistema.elver TO app_runtime_tenant;
    END IF;
END $$;

-- Registro de recurso y politicas ABAC
INSERT INTO sesion.recurso (nombre, descripcion, razon_i) VALUES
('elver', 'Modulo Elver', 'Carga Inicial')
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO sesion.politica (nombre, descripcion, perfil, recurso, accion, reglas, razon_i) VALUES
('Acceso Elver Superadmin', 'Acceso al modulo elver', 'superadmin', 'elver', 'consultar', '{"path_prefix": "/sistema/elveres", "routes": [{"method": "*", "prefix": "/sistema/elveres"}, {"method": "GET", "path": "/sistema/elveres", "accion": "consultar"}]}', 'Carga Inicial'),
('Acceso Elver Admin', 'Acceso al modulo elver', 'admin', 'elver', 'consultar', '{"path_prefix": "/sistema/elveres", "routes": [{"method": "*", "prefix": "/sistema/elveres"}, {"method": "GET", "path": "/sistema/elveres", "accion": "consultar"}]}', 'Carga Inicial'),
('Acceso Elver Operador', 'Acceso al modulo elver', 'operador', 'elver', 'consultar', '{"path_prefix": "/sistema/elveres", "routes": [{"method": "*", "prefix": "/sistema/elveres"}, {"method": "GET", "path": "/sistema/elveres", "accion": "consultar"}]}', 'Carga Inicial')
ON CONFLICT (nombre) DO NOTHING;
