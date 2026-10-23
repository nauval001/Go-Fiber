-- Tabel students
-- NIM dijaga unik oleh BASIS DATA (bukan oleh kode Go), supaya dua request
-- paralel tidak bisa menyelipkan NIM yang sama.

CREATE TABLE IF NOT EXISTS students (
    id         SERIAL PRIMARY KEY,
    nim        VARCHAR(20)  NOT NULL,
    name       VARCHAR(100) NOT NULL,
    grade      NUMERIC(5,2) NOT NULL CHECK (grade >= 0 AND grade <= 100),
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Keunikan NIM. UNIQUE sudah cukup karena NIM bukan teks bebas
-- (tidak perlu case-insensitive seperti username).
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
    ON students (nim);

-- Indeks untuk pencarian nama (ILIKE '%keyword%').
-- PostgreSQL tidak memakai indeks btree untuk ILIKE berawalan wildcard,
-- tetapi untuk data yang tumbuh besar, indeks trigram (pg_trgm) bisa membantu.
-- Untuk sekarang, indeks pada is_active lebih berguna untuk filter.
CREATE INDEX IF NOT EXISTS students_is_active_idx
    ON students (is_active);

-- Indeks untuk penyaringan rentang nilai (mis. filter min_grade/max_grade).
CREATE INDEX IF NOT EXISTS students_grade_idx
    ON students (grade);