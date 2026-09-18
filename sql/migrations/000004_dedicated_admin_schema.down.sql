UPDATE institucion.cat_tenants
SET schema_name = 'public'
WHERE subdomain = 'admin'
  AND schema_name = 'admin';

DROP SCHEMA IF EXISTS admin;
