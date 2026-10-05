-- 000007_report_retention.up.sql
-- Durable tenant-aware lifecycle metadata for private generated reports.

CREATE TABLE sistema.reportes_generados (
    report_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES institucion.cat_tenants(tenant_id),
    storage_path TEXT NOT NULL UNIQUE CHECK (storage_path LIKE 'reportes/%'),
    state VARCHAR(16) NOT NULL CHECK (state IN ('pending', 'active', 'claimed', 'retry', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    retry_count SMALLINT NOT NULL DEFAULT 0 CHECK (retry_count BETWEEN 0 AND 5),
    next_retry_at TIMESTAMPTZ,
    last_error VARCHAR(500),
    claim_token UUID,
    claimed_at TIMESTAMPTZ,
    CHECK ((state = 'claimed') = (claim_token IS NOT NULL AND claimed_at IS NOT NULL))
);

CREATE INDEX idx_reportes_generados_due
    ON sistema.reportes_generados (tenant_id, state, expires_at, next_retry_at);

ALTER TABLE sistema.reportes_generados ENABLE ROW LEVEL SECURITY;
ALTER TABLE sistema.reportes_generados FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_reportes_generados ON sistema.reportes_generados
    FOR ALL
    USING (tenant_id::text = current_setting('app.tenant_id', true))
    WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
