-- Migration: 001_create_twofactor_factors
-- Creates the twofactor_factors table to store MFA factors
-- Spec/Status pattern similar to K8s Custom Resources

CREATE TABLE IF NOT EXISTS twofactor_factors (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    
    -- Desired state (user-provided)
    spec JSONB NOT NULL,  -- FactorSpec: {type, phone_number, email, encrypted_secret, issuer}
    
    -- Observed state (reconciler-owned)
    status JSONB NOT NULL, -- FactorStatus: {phase, conditions, last_verified_at, activated_at, disabled_at, revoked_at, observed_generation}
    
    -- Optimistic concurrency and K8s-style versioning
    resource_version BIGINT NOT NULL DEFAULT 1,
    
    -- Soft delete
    deleted_at TIMESTAMP NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes for common queries
    CONSTRAINT factors_active CHECK (deleted_at IS NULL)
);

CREATE INDEX idx_twofactor_factors_user_id ON twofactor_factors(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_twofactor_factors_phase ON twofactor_factors USING GIN(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_twofactor_factors_created_at ON twofactor_factors(created_at) WHERE deleted_at IS NULL;

-- Automatic updated_at trigger
CREATE TRIGGER update_twofactor_factors_updated_at
    BEFORE UPDATE ON twofactor_factors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();