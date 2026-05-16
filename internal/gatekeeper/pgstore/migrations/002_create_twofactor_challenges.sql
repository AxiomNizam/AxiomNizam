-- Migration: 002_create_twofactor_challenges
-- Creates the twofactor_challenges table for authentication challenges
-- Like K8s Jobs, challenges are ephemeral with TTL and terminal phases

CREATE TABLE IF NOT EXISTS twofactor_challenges (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    factor_id UUID NOT NULL REFERENCES twofactor_factors(id) ON DELETE CASCADE,
    
    -- Challenge state machine
    phase VARCHAR(20) NOT NULL,  -- Waiting, Verified, Expired, Failed, Rejected
    
    -- OTP value for TOTP; empty for WebAuthn
    nonce TEXT,
    
    -- Attempt tracking and enforcement
    attempts INT NOT NULL DEFAULT 0,
    
    -- Hard TTL for automatic expiry
    expires_at TIMESTAMP NOT NULL,
    
    -- Resolved timestamp (non-null = terminal phase)
    resolved_at TIMESTAMP NULL,
    
    -- Risk context for adaptive authentication
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    
    -- Optimistic concurrency
    resource_version BIGINT NOT NULL DEFAULT 1,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes for common queries
    CONSTRAINT challenges_not_expired CHECK (expires_at > CURRENT_TIMESTAMP OR phase IN ('Verified', 'Expired', 'Failed', 'Rejected'))
);

CREATE INDEX idx_twofactor_challenges_user_id ON twofactor_challenges(user_id);
CREATE INDEX idx_twofactor_challenges_factor_id ON twofactor_challenges(factor_id);
CREATE INDEX idx_twofactor_challenges_expires_at ON twofactor_challenges(expires_at) WHERE phase NOT IN ('Verified', 'Expired', 'Failed', 'Rejected');
CREATE INDEX idx_twofactor_challenges_phase ON twofactor_challenges(phase);