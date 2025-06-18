CREATE TABLE users_sessions (
    session_id UUID PRIMARY KEY,
    user_id UUID,
    ip TEXT NOT NULL,
    ua TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    last_updated TIMESTAMP DEFAULT now()
);
