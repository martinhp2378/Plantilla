-- sql/bootstrap/roles.sql
-- Bootstrap role setup for multi-tenant isolation
-- Separates privileges across app_bootstrap, app_migrator, and app_runtime_tenant

-- 1. Create roles if they do not exist
DO $$
BEGIN
    -- app_migrator: dedicated DDL role for running schema migrations
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_migrator') THEN
        CREATE ROLE app_migrator WITH LOGIN NOCREATEDB NOCREATEROLE NOBYPASSRLS;
    END IF;

    -- app_runtime_tenant: least-privileged application runtime role
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_runtime_tenant') THEN
        CREATE ROLE app_runtime_tenant WITH LOGIN NOCREATEDB NOCREATEROLE NOBYPASSRLS;
    END IF;
END $$;

-- 2. Revoke all default privileges from PUBLIC on public schema
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM PUBLIC;

-- 3. Restrict runtime role on public schema
REVOKE ALL ON SCHEMA public FROM app_runtime_tenant;

-- 4. Grant schema usage and DML permissions to app_runtime_tenant on application schemas
GRANT USAGE ON SCHEMA institucion, sesion, sistema TO app_runtime_tenant;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA institucion, sesion, sistema TO app_runtime_tenant;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA institucion, sesion, sistema TO app_runtime_tenant;

-- 5. Grant DDL permissions to app_migrator on application schemas
GRANT ALL ON SCHEMA institucion, sesion, sistema TO app_migrator;
GRANT ALL ON ALL TABLES IN SCHEMA institucion, sesion, sistema TO app_migrator;
GRANT ALL ON ALL SEQUENCES IN SCHEMA institucion, sesion, sistema TO app_migrator;
GRANT ALL ON ALL FUNCTIONS IN SCHEMA institucion, sesion, sistema TO app_migrator;

-- 6. Configure default privileges for future objects created by app_migrator
ALTER DEFAULT PRIVILEGES FOR ROLE app_migrator IN SCHEMA institucion, sesion, sistema
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_runtime_tenant;

ALTER DEFAULT PRIVILEGES FOR ROLE app_migrator IN SCHEMA institucion, sesion, sistema
    GRANT USAGE, SELECT ON SEQUENCES TO app_runtime_tenant;
