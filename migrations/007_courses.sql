CREATE TABLE IF NOT EXISTS courses (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    credits SMALLINT NOT NULL CHECK (credits BETWEEN 1 AND 6),
    quota INTEGER NOT NULL CHECK (quota > 0),
    enrolled_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT courses_enrolled_within_quota CHECK (enrolled_count >= 0 AND enrolled_count <= quota)
);

CREATE TABLE IF NOT EXISTS enrollments (
    id BIGSERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT enrollments_course_user_unique UNIQUE (course_id, user_id)
);

CREATE INDEX IF NOT EXISTS enrollments_user_idx ON enrollments (user_id);

-- Isi data dummy kelas
INSERT INTO courses (code, name, credits, quota) VALUES
('IF101', 'Pemrograman Backend Lanjut', 3, 40),
('IF102', 'Basis Data Lanjut', 3, 40),
('IF301', 'Kecerdasan Buatan', 3, 40),
('IF302', 'Keamanan Informasi', 3, 40),
('IF399', 'Kelas Rebutan', 3, 5)
ON CONFLICT (code) DO NOTHING;

-- Tambah permission untuk admin (jika belum ada)
INSERT INTO permissions (name, description) VALUES
('course:manage', 'Mengubah kuota dan data kelas')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
('admin', 'course:manage')
ON CONFLICT DO NOTHING;