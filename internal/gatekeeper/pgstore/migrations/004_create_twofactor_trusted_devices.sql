CREATE TABLE trusted_devices (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,

    fingerprint TEXT NOT NULL,
    ip TEXT,

    trusted_until TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW()
);