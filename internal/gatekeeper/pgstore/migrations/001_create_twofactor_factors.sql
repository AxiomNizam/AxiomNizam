CREATE TABLE mfa_factors (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,

    type VARCHAR(20) NOT NULL,  -- totp, sms, webauthn
    secret TEXT,

    status VARCHAR(20) NOT NULL, -- pending, active, disabled

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_mfa_user ON mfa_factors(user_id);