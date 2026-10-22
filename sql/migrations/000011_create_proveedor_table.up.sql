-- 000011_create_proveedor_table.up.sql
-- Tabla y politicas para el modulo Proveedor

CREATE TABLE IF NOT EXISTS sistema.proveedor (
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
CREATE TRIGGER tr_set_fecha_m_proveedor
BEFORE UPDATE ON sistema.proveedor
FOR EACH ROW EXECUTE FUNCTION sistema.fn_set_fecha_m();

CREATE TRIGGER tr_auditoria_proveedor
AFTER INSERT OR UPDATE OR DELETE ON sistema.proveedor
FOR EACH ROW EXECUTE FUNCTION sistema.fn_registrar_auditoria('id');

-- Indice de aislamiento multi-tenant
CREATE INDEX IF NOT EXISTS idx_proveedor_tenant_id ON sistema.proveedor(tenant_id);

-- Row Level Security (RLS)
ALTER TABLE sistema.proveedor ENABLE ROW LEVEL SECURITY;
ALTER TABLE sistema.proveedor FORCE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE schemaname = 'sistema' AND tablename = 'proveedor' AND policyname = 'tenant_isolation_proveedor'
    ) THEN
        CREATE POLICY tenant_isolation_proveedor ON sistema.proveedor
            FOR ALL
            USING (tenant_id::text = current_setting('app.tenant_id', true))
            WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_runtime_tenant') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON sistema.proveedor TO app_runtime_tenant;
    END IF;
END $$;

-- Registro de recurso y politicas ABAC
INSERT INTO sesion.recurso (nombre, descripcion, razon_i) VALUES
('proveedor', 'Modulo Proveedor', 'Carga Inicial')
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO sesion.politica (nombre, descripcion, perfil, recurso, accion, reglas, razon_i) VALUES
('Acceso Proveedor Superadmin', 'Acceso al modulo proveedor', 'superadmin', 'proveedor', 'consultar', '{"path_prefix": "/sistema/proveedores", "routes": [{"method": "*", "prefix": "/sistema/proveedores"}]}', 'Carga Inicial'),
('Acceso Proveedor Admin', 'Acceso al modulo proveedor', 'admin', 'proveedor', 'consultar', '{"path_prefix": "/sistema/proveedores", "routes": [{"method": "*", "prefix": "/sistema/proveedores"}]}', 'Carga Inicial'),
('Acceso Proveedor Operador', 'Acceso al modulo proveedor', 'operador', 'proveedor', 'consultar', '{"path_prefix": "/sistema/proveedores", "routes": [{"method": "*", "prefix": "/sistema/proveedores"}]}', 'Carga Inicial')
ON CONFLICT (nombre) DO NOTHING;
