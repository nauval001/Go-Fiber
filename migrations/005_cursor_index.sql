-- Indeks ini harus persis sama dengan ORDER BY di query cursor.
-- Di Modul 7, kita akan mengurutkan berdasarkan created_at DESC, id DESC
CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
    ON students (created_at DESC, id DESC);