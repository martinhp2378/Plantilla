-- Keep the bootstrap tenant outside PostgreSQL's shared public schema.
CREATE SCHEMA IF NOT EXISTS admin;

UPDATE institucion.cat_tenants
SET schema_name = 'admin'
WHERE subdomain = 'admin'
  AND schema_name = 'public';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_runtime_tenant') THEN
        GRANT USAGE ON SCHEMA admin TO app_runtime_tenant;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_migrator') THEN
        GRANT USAGE ON SCHEMA admin TO app_migrator;
    END IF;
END $$;
