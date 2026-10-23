-- Tabel users untuk autentikasi (terpisah dari students).
-- Students adalah RESOURCE yang dilindungi; users adalah AKTOR yang login.
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(50)  NOT NULL,
    email      VARCHAR(255) NOT NULL,
    password   VARCHAR(255) NOT NULL,   -- bcrypt hash, bukan plain text
    role       VARCHAR(20)  NOT NULL DEFAULT 'user',
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Keunikan username tanpa memandang huruf besar/kecil.
CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_key
    ON users (LOWER(username));

CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_idx
    ON users (LOWER(email));

-- Refresh token disimpan sebagai HASH, bukan nilai aslinya.
-- Alasannya sama seperti password: kalau table bocor, penyerang
-- tetap tidak memiliki token yang dapat dipakai.
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGSERIAL  PRIMARY KEY,
    user_id    INTEGER    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT       NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_token_hash_idx ON refresh_tokens (token_hash);