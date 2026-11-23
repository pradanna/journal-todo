# jurnal-be (nama sementara)

Backend Go untuk app jurnal kerja harian freelancer. Arsitektur DDD-style:
domain → application (use case) → infrastructure (postgres, oauth, jwt) → interfaces/http.

## Setup lokal

1. Install Go 1.22+ (sudah ada di PC ini di folder `go`)
2. `cp .env.example .env` lalu isi `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET` (random string panjang),
   dan `GOOGLE_CLIENT_ID` (dari Google Cloud Console > OAuth consent screen)
3. `go mod tidy` — ambil semua dependency & generate go.sum (butuh koneksi internet, belum dijalankan dari sini)
4. Jalankan Postgres lokal (atau `docker compose up postgres -d`)
5. Jalankan migration: install `golang-migrate` CLI, lalu
   `migrate -path migrations -database "postgres://jurnal:change_me@localhost:5432/jurnal_freelance?sslmode=disable" up`
6. `go run ./cmd/api`

## Struktur folder

- `internal/domain` — entity (User, Workspace, JournalEntry, EntryItem, Tag, RefreshToken) + interface repository
- `internal/application` — use case (AuthService, EntryService)
- `internal/infrastructure/postgres` — implementasi repository pakai GORM
- `internal/infrastructure/oauth` — verifikasi Google ID token
- `internal/infrastructure/authutil` — JWT access token + refresh token (pengganti Laravel Sanctum)
- `internal/interfaces/http` — handler, middleware, routing (Fiber)
- `migrations/` — SQL migration (golang-migrate)

## Belum dikerjakan / TODO

- Endpoint update entry & update workspace (baru ada create/list/get/delete)
- Endpoint tags (CRUD + attach ke entry)
- Validasi input yang lebih ketat (sekarang masih minimal)
- Rate limiting di endpoint auth
- Dockerfile & docker-compose sudah ada tapi belum pernah dites build
