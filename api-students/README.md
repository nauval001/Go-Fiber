# Kontrak API Students

| Metode | Endpoint | Parameter / Query | Contoh Body Request | Status | Contoh Respons (Sukses) |
|---|---|---|---|---|---|
| GET | `/api/v1/students` | `?page=1&limit=10&search=budi&sort=name&is_active=true` | - | 200 | `{"success": true, "data": [...], "meta": {"page":1, "limit":10, "total": 1, "total_pages": 1}}` |
| GET | `/api/v1/students/:id` | `:id` (Path) | - | 200, 400, 404 | `{"success": true, "data": {"id": 1, "nim": "123", ...}}` |
| POST | `/api/v1/students` | - | `{"nim":"123", "name":"Budi", "grade":85.5}` | 201, 400, 409, 415, 422 | `{"success": true, "data": {"id": 1, ...}}` (Plus Header Location) |
| PUT | `/api/v1/students/:id` | `:id` (Path) | `{"nim":"123", "name":"Budi", "grade":90, "is_active":false}` | 200, 400, 404, 409, 422 | `{"success": true, "data": {...}}` |
| PATCH | `/api/v1/students/:id` | `:id` (Path) | `{"grade": 95}` | 200, 400, 404, 409, 422 | `{"success": true, "data": {...}}` |
| DELETE| `/api/v1/students/:id` | `:id` (Path) | - | 204, 400, 404 | *(Tidak ada body)* |