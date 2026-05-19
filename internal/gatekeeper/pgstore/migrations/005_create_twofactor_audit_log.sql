-- Audit log for Gatekeeper 2FA security events.
CREATE TABLE IF NOT EXISTS twofactor_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    user_id UUID NOT NULL,
    factor_id UUID,
    challenge_id UUID,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    message TEXT NOT NULL,
    source_ip VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for querying by user
CREATE INDEX IF NOT EXISTS idx_twofactor_audit_user ON twofactor_audit_log(user_id, created_at DESC);

-- Index for querying by event type
CREATE INDEX IF NOT EXISTS idx_twofactor_audit_event ON twofactor_audit_log(event_type, created_at DESC);

-- Index for querying by severity
CREATE INDEX IF NOT EXISTS idx_twofactor_audit_severity ON twofactor_audit_log(severity, created_at DESC);
