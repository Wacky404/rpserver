CREATE TABLE users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    admin BOOLEAN NOT NULL DEFAULT false,
    token VARCHAR(36),
    created_at TIMESTAMP DEFAULT now(),
    last_updated TIMESTAMP DEFAULT now()
);
