-- Keep the bootstrap tenant outside PostgreSQL's shared public schema.
CREATE SCHEMA IF NOT EXISTS admin;

UPDATE institucion.cat_tenants
SET schema_name = 'admin'
WHERE subdomain = 'admin'
  AND schema_name = 'public';

GRANT USAGE ON SCHEMA admin TO app_runtime_tenant, app_migrator;
