## Autentikasi

Semua endpoint `/api/v1/students/*` memerlukan header:

    Authorization: Bearer <access_token>

### Endpoint Auth

| Metode | Endpoint | Body | Sukses | Gagal |
|--------|----------|------|--------|-------|
| POST | `/auth/register` | `{username, email, password}` | 201 | 400, 422, 409 |
| POST | `/auth/login` | `{username, password}` | 200 | 400, 401, 422, 429 |
| POST | `/auth/refresh` | `{refresh_token}` | 200 | 400, 401 |
| POST | `/auth/logout` | `{refresh_token}` | 200 | — |
| GET | `/auth/me` | — (butuh Bearer) | 200 | 401 |

### Environment Baru

| Variabel | Default | Keterangan |
|----------|---------|------------|
| `JWT_SECRET` | — | **Wajib**, min 32 karakter acak |
| `JWT_ISSUER` | praktikum-backend | Issuer claim di JWT |
| `JWT_ACCESS_TTL_MINUTES` | 15 | Masa berlaku access token |
| `JWT_REFRESH_TTL_DAYS` | 7 | Masa berlaku refresh token |
| `ALLOWED_ORIGINS` | http://localhost:5173 | Origin CORS yang diizinkan |