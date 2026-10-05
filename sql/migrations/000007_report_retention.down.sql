-- 000007_report_retention.down.sql
DROP POLICY IF EXISTS tenant_isolation_reportes_generados ON sistema.reportes_generados;
DROP TABLE IF EXISTS sistema.reportes_generados;
