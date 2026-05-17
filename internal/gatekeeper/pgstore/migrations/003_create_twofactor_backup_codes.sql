-- Migration: 003_create_twofactor_backup_codes
-- Creates the twofactor_backup_codes table for one-time recovery codes
-- Backup codes are generated as a set during factor enrollment

CREATE TABLE IF NOT EXISTS twofactor_backup_codes (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    factor_id UUID NOT NULL REFERENCES twofactor_factors(id) ON DELETE CASCADE,
    
    -- Argon2id / bcrypt hash of the plaintext code
    code_hash BYTEA NOT NULL,
    
    -- Timestamp of consumption; null if unused
    used_at TIMESTAMP NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes for common queries
    CONSTRAINT backup_codes_single_use CHECK (used_at IS NULL OR used_at IS NOT NULL)
);

CREATE INDEX idx_twofactor_backup_codes_user_id ON twofactor_backup_codes(user_id);
CREATE INDEX idx_twofactor_backup_codes_factor_id ON twofactor_backup_codes(factor_id);
CREATE INDEX idx_twofactor_backup_codes_unused ON twofactor_backup_codes(user_id) WHERE used_at IS NULL;