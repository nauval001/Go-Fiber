CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade DECIMAL(5,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Menjaga keunikan NIM
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key ON students (nim);

-- Indeks tambahan untuk mempercepat pencarian (ILIKE) pada nama
CREATE INDEX IF NOT EXISTS students_name_lower_idx ON students (LOWER(name));