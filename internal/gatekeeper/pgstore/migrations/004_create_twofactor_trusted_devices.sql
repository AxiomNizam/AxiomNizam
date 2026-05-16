-- Migration: 004_create_twofactor_trusted_devices
-- Creates the twofactor_trusted_devices table for device trust tokens
-- Once a user verifies MFA on a device, they can mark it trusted to skip MFA

CREATE TABLE IF NOT EXISTS twofactor_trusted_devices (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    
    -- bcrypt/argon2id hash of the device token
    token_hash BYTEA NOT NULL,
    
    -- Device fingerprinting (browser/OS identifier)
    fingerprint VARCHAR(255) NOT NULL,
    
    -- User agent for tracking and invalidation
    user_agent TEXT NOT NULL,
    
    -- IP address at registration time
    ip_address VARCHAR(45) NOT NULL,
    
    -- TTL for device trust (configurable per policy)
    expires_at TIMESTAMP NOT NULL,
    
    -- Explicit revocation timestamp
    revoked_at TIMESTAMP NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes for common queries
    CONSTRAINT trusted_device_not_revoked CHECK (revoked_at IS NULL OR revoked_at IS NOT NULL)
);

CREATE INDEX idx_twofactor_trusted_devices_user_id ON twofactor_trusted_devices(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_twofactor_trusted_devices_fingerprint ON twofactor_trusted_devices(user_id, fingerprint) WHERE revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP;
CREATE INDEX idx_twofactor_trusted_devices_expires_at ON twofactor_trusted_devices(expires_at) WHERE revoked_at IS NULL;