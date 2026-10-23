-- 1. Tabel Roles & Permissions
CREATE TABLE IF NOT EXISTS roles (
    name VARCHAR(20) PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
('admin', 'Akses penuh'),
('staff', 'Boleh melihat dan menambah data, tidak boleh mengubah/menghapus'),
('user', 'Hanya boleh mengelola datanya sendiri')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

INSERT INTO permissions (name, description) VALUES
('student:list', 'Melihat daftar seluruh mahasiswa'),
('student:read:any', 'Melihat data mahasiswa mana pun'),
('student:create', 'Menambah data mahasiswa'),
('student:update:any', 'Mengubah data mahasiswa mana pun'),
('student:delete', 'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO role_permissions (role_name, permission_name) VALUES
('admin', 'student:list'), ('admin', 'student:read:any'),
('admin', 'student:create'), ('admin', 'student:update:any'), ('admin', 'student:delete'),
('staff', 'student:list'), ('staff', 'student:read:any'), ('staff', 'student:create')
ON CONFLICT DO NOTHING;
-- Role 'user' tidak diberi permission (hanya bisa akses miliknya sendiri)

-- 2. Kunci foreign key pada tabel users
UPDATE users SET role = 'user' WHERE role NOT IN (SELECT name FROM roles);
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_fkey;
ALTER TABLE users ADD CONSTRAINT users_role_fkey FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;

-- 3. Tambah owner_id di tabel students
-- HATI-HATI: Kita kosongkan data students lama agar bisa menambahkan NOT NULL constraint
TRUNCATE students RESTART IDENTITY CASCADE;

ALTER TABLE students ADD COLUMN owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students(owner_id);