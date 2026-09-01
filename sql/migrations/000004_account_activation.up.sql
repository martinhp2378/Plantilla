-- 000004_account_activation.up.sql
-- One-time account activation and password-reset tokens are stored only as hashes.

ALTER TABLE sesion.usuario
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;

CREATE TABLE IF NOT EXISTS sesion.password_reset_token (
    token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uuid_persona UUID NOT NULL REFERENCES institucion.persona(uuid_persona) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    tenant_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_token_persona ON sesion.password_reset_token(uuid_persona);
CREATE INDEX IF NOT EXISTS idx_password_reset_token_expiry ON sesion.password_reset_token(expires_at);

ALTER TABLE sesion.password_reset_token ENABLE ROW LEVEL SECURITY;
ALTER TABLE sesion.password_reset_token FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_password_reset_token ON sesion.password_reset_token
    FOR ALL
    USING (tenant_id::text = current_setting('app.tenant_id', true))
    WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
