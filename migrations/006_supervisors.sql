CREATE TABLE IF NOT EXISTS supervisors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    quota INTEGER NOT NULL CHECK (quota > 0),
    supervised_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT supervisors_count_within_quota CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

-- Tabel supervisions: 1 mahasiswa hanya boleh punya 1 pembimbing (PRIMARY KEY student_id)
CREATE TABLE IF NOT EXISTS supervisions (
    student_id INTEGER PRIMARY KEY REFERENCES students(id) ON DELETE CASCADE,
    supervisor_id INTEGER NOT NULL REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert data dummy
INSERT INTO supervisors (name, quota) VALUES 
('Dr. Budi (Kuota 3)', 3), 
('Prof. Sari (Kuota 5)', 5), 
('Ir. Andi (Kuota 2)', 2)
ON CONFLICT DO NOTHING;