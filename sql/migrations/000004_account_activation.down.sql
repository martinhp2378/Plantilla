-- 000004_account_activation.down.sql
DROP POLICY IF EXISTS tenant_isolation_password_reset_token ON sesion.password_reset_token;
ALTER TABLE IF EXISTS sesion.password_reset_token NO FORCE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS sesion.password_reset_token DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS sesion.password_reset_token;
ALTER TABLE sesion.usuario DROP COLUMN IF EXISTS is_active;
