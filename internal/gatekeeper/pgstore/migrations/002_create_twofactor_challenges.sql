CREATE TABLE mfa_challenges (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,

    factor_id UUID NOT NULL,
    code_hash TEXT,

    expires_at TIMESTAMP,
    status VARCHAR(20),

    created_at TIMESTAMP DEFAULT NOW()
);