CREATE TABLE mfa_backup_codes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,

    code_hash TEXT NOT NULL,
    used BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP DEFAULT NOW()
);