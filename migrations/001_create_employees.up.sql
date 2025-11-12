CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name TEXT NOT NULL CHECK (char_length(full_name) >= 2 AND char_length(full_name) <= 200),
    phone TEXT NOT NULL CHECK (phone ~ '^\+[1-9]\d{1,14}$'),
    city TEXT NOT NULL CHECK (char_length(city) >= 2 AND char_length(city) <= 120),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_employees_phone ON employees(phone);
CREATE INDEX idx_employees_full_name_trgm ON employees USING gin(full_name gin_trgm_ops);
CREATE INDEX idx_employees_city ON employees(city);
