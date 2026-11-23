# Agent BE — Backend (Go)

File ini adalah brief lengkap untuk agent Backend: konteks, aturan kerja, urutan pengerjaan, dan detail setiap task sampai Sync API untuk mobile.

Backend dikerjakan oleh **dua agent paralel** — **BE-A (jalur inti)** dan **BE-B (jalur pendukung)** — setelah M0 & M1 selesai. Lihat "Pembagian jalur". Aturan coding yang berlaku terus ada di `backend/AGENTS.md` (wajib dipatuhi di setiap task).

## Peran & batasan
- Kamu mengerjakan **seluruh backend** di folder `backend/` dan menjaga kontrak API di `docs/api/openapi.yaml`.
- Agent lain (FE Web, lalu Mobile) hanya memakai endpoint yang ada di `openapi.yaml`. Setiap fitur yang kamu selesaikan membuka pekerjaan FE lewat **GATE** (lihat tiap milestone).
- Keputusan produk ada di `plan-project.md` (root repo). Poin **[FIX]** tidak boleh diubah tanpa konfirmasi Pradana.

## Konteks produk (ringkas)
App jurnal kerja + todo pintar untuk siapa saja yang bekerja per-task/per-hari. Web dulu (SvelteKit), lalu Android (Flutter, offline-first).
- **Workspace** = klien/project (satu level, hanya bisa diarsip, tidak dihapus).
- **Todo**: prioritas, deadline, tanggal rencana, sub-task 1 level. Semua todo bertanggal adalah *komitmen*: lewat & belum selesai → **Terlewat** (mengurangi % kepatuhan). Tanpa tanggal → **Inbox**. Tidak ada carry-over.
- **Rutin**: todo berulang (harian/mingguan/bulanan) dari template RRULE.
- **% kepatuhan** 30 hari bergulir: tepat waktu 1, terlambat 0.5, terlewat 0, dilewati tidak dihitung.
- **Timer/TimeLog**: sumber tunggal jam kerja per workspace.
- **Jurnal harian**: rekap otomatis (dihitung) + catatan user + mood.
- **Laporan** per workspace, **Reminder** (web push; Android dijadwalkan lokal di HP).

## Wajib dibaca sebelum mulai
1. `docs/spec/rules.md` — **semua** formula & aturan perilaku (hari ini, terlewat, skor prioritas, recurring, kepatuhan, timer, rekap jurnal, reminder). Jangan menulis aturan bisnis yang bertentangan dengannya.
2. `docs/spec/api-conventions.md` — format respons/error, UUID, tanggal, auth web vs android, CORS, rate limit.
3. `plan-project.md` — bagian 1A (fitur) dan 2 (domain model).

## Stack
| Kebutuhan | Pilihan |
|---|---|
| Bahasa | Go 1.24+ (dibutuhkan direktif `tool` di go.mod) |
| HTTP router | **chi v5** (`go-chi/chi`) di atas `net/http` standar + `go-chi/cors`, `go-chi/httprate` |
| Kontrak API | **Spec-first**: `docs/api/openapi.yaml` → **oapi-codegen** (strict server untuk chi + model) |
| Validasi request | Middleware validator OpenAPI (`oapi-codegen/nethttp-middleware` + `kin-openapi`) — request divalidasi terhadap spec |
| Database | PostgreSQL 16 + **pgx v5** (`pgxpool`) |
| Query | **sqlc** — query ditulis sebagai SQL, kode Go type-safe digenerate |
| Migration | golang-migrate |
| ID | **UUID v7** (`google/uuid` → `uuid.NewV7()`); UUID v5 untuk kejadian rutin & turunannya |
| Auth | `golang-jwt/jwt/v5`, `google.golang.org/api/idtoken`, bcrypt |
| Lainnya | `teambition/rrule-go`, `SherClockHolmes/webpush-go`, `log/slog` |
| Test & lint | testcontainers-go (Docker), golangci-lint (aturan `forbidigo` melarang `time.Now()` di domain/application) |
| Task runner | **Task** (`Taskfile.yml`) — jalan sama di Windows (PowerShell), Linux, dan CI |

**Lingkungan development: Windows** (PC Pradana) — produksi & CI: Linux. Semua perintah harus jalan di keduanya:
- Pakai `task <nama>`, bukan `make`. Isi task tidak boleh bergantung pada perintah khusus bash (`grep`, `sed`, `rm -rf`, `&&` berantai panjang) — kalau butuh logika, tulis sebagai program Go kecil di `cmd/` atau `internal/tools/`.
- Tidak ada skrip `.sh`. Seed, migrasi, dan alat bantu lain = program Go (`go run ./cmd/<nama>`).
- `.gitattributes` memaksa line ending LF (`* text=auto eol=lf`) — tanpa ini, cek "hasil generate up-to-date" di CI akan gagal karena CRLF dari Windows.

**Mengapa pilihan ini** (sudah diputuskan Pradana, jangan diganti):
- **Spec-first + generate**: kontrak ditulis dulu, handler wajib memenuhi interface hasil generate → kode yang tidak sesuai kontrak gagal compile. FE bisa mulai dengan mock dari kontrak sebelum endpoint selesai.
- **sqlc + pgx**: query terpenting app ini adalah agregat (kepatuhan, laporan, rekap) — lebih jelas dan akurat sebagai SQL eksplisit.
- **chi**: standar `net/http`, semua library Go kompatibel tanpa adapter.
- **UUID v7**: berurutan menurut waktu → index lebih efisien; tetap bisa dibuat di client untuk offline.

## Aturan kerja
- **Satu langkah/task = satu branch/PR**: `feat/BE-<id>-<slug>`. Commit kecil dan jelas.
- **Kontrak dulu**: setiap fitur dimulai dengan menulis kontraknya di `docs/api/src/` (lihat M0.6), lalu `task generate`. Begitu PR kontrak di-merge, FE boleh mulai dengan mock (**GATE-K**). GATE penuh = endpoint nyata jalan.
- **Kode hasil generate tidak boleh diedit manual** (`internal/api/gen/`, `internal/infrastructure/postgres/sqlcgen/`). CI memeriksa hasil generate sudah up-to-date (`task generate && git diff --exit-code`).
- **Soft delete manual**: tidak ada ORM yang menyaring otomatis — **setiap query** baca wajib `deleted_at IS NULL`, dan "hapus" = `UPDATE ... SET deleted_at = now()`.
- **Definition of Done** setiap task:
  - Acceptance criteria (AC) terpenuhi
  - Unit test domain & application; integration test repository + HTTP ke Postgres asli
  - golangci-lint & build lolos di CI
  - Kontrak (`docs/api/src/`), `docs/api/openapi.yaml` hasil bundle, dan hasil generate (oapi-codegen, sqlc) up-to-date; lint OpenAPI lolos
  - Tidak ada secret di repo
- **Jangan pernah** `time.Now()` di domain/application — pakai `Clock` (dicek otomatis di CI).
- **Semua query data user wajib difilter `user_id` di repository.** Resource milik user lain → 404.
- **Logika bisnis di domain/application, bukan di handler** — Sync API (M12) akan memanggil logika yang sama.
- Formula hanya boleh diubah bersamaan dengan `rules.md` + test vector-nya, dalam PR yang sama, dan dengan konfirmasi.
- Migration: satu file baru per perubahan, **nama berbasis timestamp** (`migrations/20261124093000_<nama>.up/down.sql`). Jangan mengedit migration yang sudah di-merge. Detail anti-bentrok di M0.9.
- Kalau ada kontradiksi antara file ini, `rules.md`, dan plan → berhenti, tulis pertanyaan di PR.
- Identifier kode bahasa Inggris; pesan error untuk user bahasa Indonesia.

## Pola baku per fitur (dipakai mulai M1)
1. **Kontrak** — `docs/api/src/paths/<fitur>.yaml` + `docs/api/src/schemas/<fitur>.yaml` → `task generate` → stub handler `501` untuk operasi baru → PR kontrak (**GATE-K**: FE boleh mulai dengan mock)
2. **Migration** — tabel, index, partial unique `WHERE deleted_at IS NULL`, trigger `set_updated_at`
3. **Domain** (`internal/domain/<fitur>/`) — entity, aturan/transisi, interface repository, error domain + unit test tanpa DB
4. **Query & repository** — SQL di `internal/infrastructure/postgres/queries/<fitur>.sql` → `task generate` (sqlc) → repository tipis di `internal/infrastructure/postgres/<fitur>_repository.go` yang mengubah hasil sqlc ke entity domain + integration test
5. **Application service** (`internal/application/<fitur>/`) — use case, cek kepemilikan, transaksi, pakai Clock + unit test dengan repository fake
6. **Handler** (`internal/interfaces/http/<fitur>_handler.go`) — ganti stub `501` dengan implementasi method `StrictServerInterface`, mapping error domain → respons + integration test HTTP
7. **Seed dev** (opsional) — `cmd/seed` agar FE bisa mencoba

## Urutan & status

| # | ID | Milestone | Jalur | Ukuran | Membuka FE | Status |
|---|---|---|---|---|---|---|
| M0 | BE-00 | Fondasi | A | M–L | FE-00 | ⬜ |
| M1 | BE-01 | Auth dasar | A | L | FE-01 | ⬜ |
| M2 | BE-02 | Workspace | A | M | FE-02 | ⬜ |
| M3 | BE-03 | Todo dasar + sub-task | A | L | FE-03 | ⬜ |
| M4 | BE-04 | Layar Hari Ini + prioritas | A | M | FE-04 | ⬜ |
| M4.5 | BE-13a | Staging lokal (Docker) | B | M | — | ⬜ |
| M5 | BE-07 | Timer + TimeLog | B | M | FE-07 | ⬜ |
| M6 | BE-05 | Rutin (todo berulang) | A | L | FE-05 | ⬜ |
| M7 | BE-06 | Terlewat + kepatuhan | A | L | FE-06 | ⬜ |
| M8a | BE-09a | Jurnal: catatan + mood | B | M | FE-09 (sebagian), FE-08 | ⬜ |
| M8b | BE-09b | Jurnal: rekap otomatis | A | M | FE-09 | ⬜ |
| M9 | BE-10 | Laporan per workspace | B | M | FE-10 | ⬜ |
| M10a | BE-11a | Reminder + Web Push | B | M | FE-11 | ⬜ |
| M10b | BE-11b | Reminder offset untuk rutin | B | S | FE-11 (rutin) | ⬜ |
| M11 | BE-13 | Produksi web MVP | B | M | — (web live) | ⬜ |
| M12 | BE-12 | Sync API | A | L | Mobile MB-02 | ⬜ |

Ukuran: **S** ≈ setengah hari, **M** ≈ 1–2 hari, **L** ≈ 3–4 hari kerja (perkiraan kasar).
Status: ⬜ belum · 🟡 dikerjakan · 🔷 GATE-K (kontrak) lolos · ✅ GATE penuh lolos. Tabel status ini **diperbarui oleh Pradana**, bukan oleh agent (supaya dua agent tidak bentrok di baris yang sama). Agent cukup mencentang `[x]` langkah di bagian milestone miliknya sendiri.

## Pembagian jalur (2 agent BE)

- **BE-A (jalur inti)** mengerjakan jalur kritis: M0 → M1 → M2 → M3 → M4 → M6 → M7 → M8b → M12. Jalur ini menentukan kapan MVP selesai.
- **BE-B (jalur pendukung)** mulai setelah M1: M4.5 → M5 → M8a → M10a → M10b → M9 → M11.

| Urutan | BE-A (jalur inti) | BE-B (jalur pendukung) |
|---|---|---|
| 1 | **M0** Fondasi | — (belum mulai) |
| 2 | **M1** Auth | — |
| 3 | **M2** Workspace | **M4.5** Staging lokal *(mulai setelah M1)* |
| 4 | **M3** Todo | menunggu M3 — boleh menyiapkan seed & dokumentasi deploy |
| 5 | **M4** Hari Ini | **M5** Timer *(butuh M3)* |
| 6 | **M6** Rutin | **M8a** Catatan jurnal + mood *(butuh M2)* |
| 7 | **M7** Terlewat + kepatuhan | **M10a** Reminder + Web Push *(butuh M3)* |
| 8 | **M8b** Rekap jurnal *(butuh M5, M7, M8a)* | **M10b** Reminder offset rutin *(butuh M6, M10a)* |
| 9 | **M12** Sync API *(butuh semua entity)* | **M9** Laporan *(butuh M5, M7)* |
| 10 | — | **M11** Produksi *(setelah semua milestone web)* |

Aturan koordinasi:
- **M0 dan M1 dikerjakan BE-A sendirian.** BE-B baru mulai setelah GATE G1 — tanpa user & middleware login, fitur lain belum bisa diuji.
- Kalau milestone yang ditunggu belum lolos GATE, **jangan menunggu diam**: kerjakan bagian yang tidak bergantung (kontrak, migration, domain, test vector), atau ambil milestone berikutnya di jalurmu yang sudah terbuka.
- Satu milestone hanya dikerjakan **satu agent**. Tanda "sedang dikerjakan" = **draft PR** berjudul `[BE-A] M6 Rutin` / `[BE-B] M5 Timer` dibuka saat mulai. Sebelum mengambil milestone, cek draft PR agent lain.
- Perubahan pada kode milik milestone agent lain (mis. BE-B perlu menambah method di domain todo) → PR kecil terpisah yang di-review lebih dulu, jangan diselipkan di PR fitur.

---

## M0 — Fondasi · `BE-00` · Jalur A

- [ ] **0.1 Bersihkan scaffold lama** (S)
  - Hapus seluruh kode lama berbasis Fiber & GORM: `internal/domain/*.go`, `internal/application/*.go`, `internal/infrastructure/postgres/*.go`, `internal/interfaces/http/**`, `migrations/000001_init.*`, dependensi `gofiber/fiber`, `gorm.io/*`. Kode auth/oauth lama hanya referensi; ditulis ulang di M1.
  - Module path → `github.com/pradanna/jurnal-todo` (sampai nama app final).
  - Tulis ulang `backend/README.md` (masih menjelaskan stack lama) dan sesuaikan `docker-compose.yml` lama → `docker-compose.dev.yml` (Postgres saja).
  - Patuhi `backend/AGENTS.md` mulai langkah ini.
  - AC: `go build ./...` lolos; `go.mod` tidak lagi memuat Fiber/GORM.
- [ ] **0.2 Migration dasar** (S) — `000001_init.up.sql`: extension `pgcrypto`, `citext`; fungsi trigger `set_updated_at()` (dipasang di setiap tabel berikutnya agar `updated_at` selalu diisi server — dibutuhkan Sync di M12).
- [ ] **0.3 Database & tipe bersama** (S)
  - `pgxpool` + helper transaksi `WithTx(ctx, fn)`.
  - `sqlc.yaml`: engine postgresql, schema dari `migrations/`, `sql_package: pgx/v5`, output `internal/infrastructure/postgres/sqlcgen`, override `uuid` → `github.com/google/uuid.UUID`, `date` → tipe `shared.Date`.
  - `internal/domain/shared`: tipe `Date` (tanggal tanpa zona, JSON `YYYY-MM-DD`), helper `NewID()` = UUID v7, helper UUIDv5.
- [ ] **0.4 Clock & zona waktu** (S) — interface `Clock { Now() time.Time }`, `RealClock`, `FakeClock`; helper `TodayIn(tz)`, `LocalDate(t, tz)`. Unit test pergantian hari di +07:00.
- [ ] **0.5 Config** (S) — env: DB, `JWT_ACCESS_SECRET`, `GOOGLE_CLIENT_IDS` (dipisah koma), `CORS_ALLOWED_ORIGINS`, `VAPID_*`, `APP_ENV`, `MIGRATE_ON_START`. Divalidasi saat start; gagal start kalau env wajib kosong.
- [ ] **0.6 OpenAPI kerangka + generator** (M)
  - Kontrak **dipecah per fitur** agar dua agent tidak mengedit file yang sama:
    - `docs/api/src/openapi.yaml` — root: info, servers, security, dan `$ref` ke tiap path
    - `docs/api/src/paths/<fitur>.yaml` — path per fitur (auth, workspaces, todos, …)
    - `docs/api/src/schemas/<fitur>.yaml` + `docs/api/src/schemas/common.yaml` — Error, pagination, UUID, Date, Timestamp
  - `task generate` menjalankan `redocly bundle` → **`docs/api/openapi.yaml`** (satu file hasil bundle, di-commit, **tidak boleh diedit manual**) → oapi-codegen → sqlc. FE & Mobile memakai file hasil bundle ini.
  - Versi OpenAPI 3.0.3 (didukung penuh oapi-codegen).
  - `backend/api/oapi-codegen.yaml`: generate `chi-server` + `strict-server` + `models` + `embedded-spec` ke `internal/api/gen`.
  - Scalar/Swagger UI di `/docs` saat development.
- [ ] **0.7 Infrastruktur HTTP** (M)
  - Router chi, prefix `/api/v1`; `GET /health` (cek koneksi DB) di luar prefix.
  - Middleware: `RequestID`, logging `log/slog` (JSON di produksi), `Recoverer`, `go-chi/cors` (`AllowCredentials: true`), `httprate` (auth 10/menit per IP, lainnya 120/menit per user), validator request OpenAPI.
  - Package `apperror` + error handler strict server & validator → format error standar (`api-conventions.md`): validasi gagal → `422` dengan `fields`.
  - Graceful shutdown (SIGTERM).
  - AC: request yang melanggar spec ditolak `422` sebelum masuk handler; panic → `500` berformat standar tanpa mematikan server.
- [ ] **0.8 Tooling & CI** (M) ⟂ boleh paralel dengan 0.7 — **harus jalan di Windows dan Linux**
  - Hapus `Makefile` lama; buat `backend/Taskfile.yml` dengan task: `tools`, `generate`, `run`, `test`, `test-integration`, `lint`, `migrate:up`, `migrate:down`, `migrate:create -- <nama>`, `seed`, `openapi:lint`, `db:up` / `db:down` (Docker compose dev).
  - Versi tool dikunci:
    - oapi-codegen → direktif `tool` di `go.mod` (`go tool oapi-codegen`)
    - sqlc → binary resmi atau image Docker `sqlc/sqlc:<versi>` (`go install` sqlc di Windows rawan gagal karena butuh compiler C)
    - redocly → `npx @redocly/cli@<versi>` (butuh Node.js)
    - golangci-lint → binary resmi versi tetap
    - Task `tools` mengecek semua tool tersedia & versinya cocok, dan menampilkan cara install di Windows (`winget`/`scoop`) bila belum ada.
  - Migrasi lewat program Go `cmd/migrate` (library golang-migrate): `up`, `down`, `create <nama>` (membuat file bernama timestamp) — tidak butuh CLI `migrate` terpisah, sama di Windows & Linux. API juga bisa migrasi saat start (`MIGRATE_ON_START`).
  - Seed lewat `cmd/seed`.
  - `.gitattributes` (`* text=auto eol=lf`) + `.editorconfig` di **root repo** (berlaku juga untuk `web/` & `mobile/`).
  - `docker-compose.dev.yml` (Postgres); integration test pakai testcontainers (Docker Desktop di PC Pradana sudah tersedia).
  - golangci-lint dengan `forbidigo` untuk melarang `time.Now()` di `domain/` & `application/` (menggantikan cek berbasis grep).
  - GitHub Actions (Ubuntu): `task tools`, lint, cek hasil generate up-to-date, test, build.
  - Opsional untuk server Linux: `Makefile` tipis yang meneruskan ke Task (`%: ; @task $@`) bagi yang terbiasa mengetik `make`.
  - AC: dari clone baru di Windows (PowerShell) maupun Linux, `task tools && task db:up && task generate && task test` berhasil tanpa langkah manual lain.
- [ ] **0.9 Aturan anti-bentrok untuk kerja paralel** (S) — wajib selesai sebelum BE-B mulai
  - **Migration bernama timestamp**: `task migrate:create -- <nama>`. Cek CI: setiap migration baru di PR harus bernomor **lebih besar** dari migration terakhir di `main`; kalau tidak (karena branch lain merge lebih dulu), rename ke timestamp baru sebelum merge. Tanpa cek ini, migration yang tertinggal bisa terlewat diam-diam di database yang sudah berjalan.
  - **Server tersusun per fitur**: `internal/interfaces/http/server.go` berisi struct `Server` yang meng-embed handler tiap fitur (`*AuthHandler`, `*TodoHandler`, …) sehingga menambah fitur = satu baris. Wiring dependensi per fitur di `internal/app/<fitur>.go`, bukan menumpuk di `main.go`.
  - **Stub 501**: karena interface hasil generate memuat semua operasi, PR kontrak wajib menambahkan stub handler yang membalas `501 NOT_IMPLEMENTED` untuk operasi barunya, supaya `main` selalu bisa di-compile.
  - **Query per fitur**: `queries/<fitur>.sql` — satu file per fitur, tidak berbagi file.

**GATE G0 → FE-00 boleh mulai:** `task run` jalan + `/health` OK · `task generate` bersih · `docs/api/openapi.yaml` (bundle) valid dan bisa di-generate ke TypeScript · aturan anti-bentrok (0.9) aktif di CI · CI hijau.

---

## M1 — Auth dasar · `BE-01` · Jalur A
Detail lanjutan (verifikasi email, lupa password, hapus akun) **ditunda** — lihat plan Bagian 3.

- [ ] **1.0 Kontrak dulu** — tulis `docs/api/src/paths/<fitur>.yaml` + schema (termasuk kode error), `task generate`, stub `501` untuk operasi baru, merge → **GATE-K**: FE boleh mulai dengan mock
- [ ] **1.1 Migration**
  - `users`: id, email (`citext`), password_hash (null), google_id (null), name, timezone (default `Asia/Jakarta`), timestamps, deleted_at. Unique `email` & `google_id` `WHERE deleted_at IS NULL`.
  - `refresh_tokens`: id, user_id, token_hash (sha256), client (`web|android`), device_info, expires_at (30 hari), revoked_at, replaced_by, created_at.
- [ ] **1.2 Domain user** — validasi email, password (min 8), timezone (IANA); error `EMAIL_TAKEN`, `INVALID_CREDENTIALS`.
- [ ] **1.3 Token service** — access JWT HS256 15 menit (`sub` = user id); refresh token acak, simpan hash; **rotasi** tiap dipakai (`revoked_at` + `replaced_by`); pemakaian ulang token yang sudah dirotasi → cabut seluruh rantai token user. Unit test rotasi & reuse.
- [ ] **1.4 Register & login password**
  - `POST /auth/register` {email, password, name, timezone} → buat user + langsung login.
  - `POST /auth/login` {email, password, timezone?}. Bcrypt. Respons gagal identik untuk email tidak ada & password salah.
  - Header `X-Client`: `web` → refresh token di cookie `refresh_token` (`HttpOnly; Secure; SameSite=Lax; Path=/api/v1/auth`), **tidak** di body; `android` → di body.
  - AC: email duplikat → `409 EMAIL_TAKEN`.
- [ ] **1.5 Refresh & logout** — `POST /auth/refresh` (web: cookie, android: body), `POST /auth/logout`, `POST /auth/logout-all`.
- [ ] **1.6 Login Google** — `POST /auth/google` {id_token, timezone?}; verifikasi `idtoken`, audience salah satu `GOOGLE_CLIENT_IDS`, wajib `email_verified = true`. Penautan akun:
  - `google_id` sudah ada → login
  - email sama tanpa `google_id` → tautkan Google, **kosongkan `password_hash`**, cabut semua refresh token lama (pemilik password belum terbukti memiliki email itu)
  - tidak ada → buat user baru
  - AC: test ketiga cabang dengan verifier palsu; password lama tidak bisa dipakai setelah penautan.
- [ ] **1.7 Middleware auth & profil** — muat user (termasuk timezone) ke context; `GET /me`, `PATCH /me` {name?, timezone?}.

**GATE G1 → FE-01 + BE-B boleh mulai:** register, login, Google, refresh via cookie, logout, `/me` terdokumentasi & teruji · CORS + cookie jalan dari `http://localhost:5173`.

---

## M2 — Workspace · `BE-02` · Jalur A

- [ ] **2.0 Kontrak dulu** — tulis `docs/api/src/paths/<fitur>.yaml` + schema (termasuk kode error), `task generate`, stub `501` untuk operasi baru, merge → **GATE-K**: FE boleh mulai dengan mock
- [ ] **2.1 Migration `workspaces`**: id, user_id, name (1–60), color (hex), icon (key), sort_order, archived_at, timestamps, deleted_at.
- [ ] **2.2 Domain** — validasi nama/warna/ikon, aturan arsip, error `WORKSPACE_ARCHIVED`.
- [ ] **2.3 Repository + service**
- [ ] **2.4 Endpoint** — `GET /workspaces?include_archived=false`, `POST /workspaces`, `GET/PATCH /workspaces/{id}`, `POST /workspaces/{id}/archive`, `POST /workspaces/{id}/unarchive`, `PUT /workspaces/order` {ids}. **Tidak ada DELETE.**
- [ ] **2.5 Helper kepemilikan** — `EnsureActiveOwned(userID, workspaceID)` dipakai todo, timer, catatan. Workspace arsip tidak bisa dipakai untuk data baru (`422 WORKSPACE_ARCHIVED`) tapi tetap muncul di laporan & histori.
- AC: user A tidak bisa membaca/mengubah workspace user B (404).

**GATE G2 → FE-02.**

---

## M3 — Todo dasar + sub-task · `BE-03` · Jalur A

- [ ] **3.0 Kontrak dulu** — tulis `docs/api/src/paths/<fitur>.yaml` + schema (termasuk kode error), `task generate`, stub `501` untuk operasi baru, merge → **GATE-K**: FE boleh mulai dengan mock
- [ ] **3.1 Migration `todos`**: id, user_id, workspace_id (null), parent_id (null), title (1–200), notes, status, priority (1–4, default 3), scheduled_date (DATE null), due_at (null), estimate_minutes (null), completed_at, source (`manual|journal|ai`), source_note_id (null), recurrence_rule (null), recurrence_parent_id (null), generated_until (DATE null), timestamps, deleted_at.
  - Index: (user_id, status), (user_id, scheduled_date), (user_id, due_at), (recurrence_parent_id).
  - Kolom recurring dibuat sekarang supaya M6 tidak perlu ALTER besar. **Tidak ada** kolom carry-over atau status `missed`.
- [ ] **3.2 Domain todo** (rules §2) — transisi status (tolak yang tidak valid: `422 INVALID_STATUS_TRANSITION`); sub-task maks 1 level, bukan di bawah template (`422 SUBTASK_DEPTH`), mewarisi workspace induk; induk `done` → sub-task terbuka ikut `done`; turunan `IsMissed(today, now)`, `CommitmentDate()`, `CompletedLate()`. Unit test tabel lengkap.
- [ ] **3.3 Query & repository** — todo + sub-task dalam satu query (sqlc); filter list dengan parameter opsional (`sqlc.narg`).
- [ ] **3.4 CRUD** — `POST /todos` (boleh kirim `id` UUID v7 dari client, idempotent), `GET /todos/{id}`, `PATCH /todos/{id}`, `DELETE /todos/{id}` (soft, ikut sub-task), `GET /todos?view=inbox|all&workspace_id=&status=&from=&to=`.
- [ ] **3.5 Aksi status** — `POST /todos/{id}/start|complete|reopen|skip|cancel`. `completed_at` terisi/terhapus sesuai transisi.

**GATE G3 → FE-03.**

---

## M4 — Layar Hari Ini + auto-prioritas · `BE-04` · Jalur A

- [ ] **4.0 Kontrak dulu** — tulis `docs/api/src/paths/<fitur>.yaml` + schema (termasuk kode error), `task generate`, stub `501` untuk operasi baru, merge → **GATE-K**: FE boleh mulai dengan mock
- [ ] **4.1 Test vector prioritas** — tulis `docs/spec/test-vectors/priority.json` **lebih dulu** (≥15 kasus: kombinasi prioritas, due hari ini/besok/≤3/≤7 hari, rutin, tie-break). Format: input todo + "hari ini" + timezone → skor & urutan yang diharapkan.
- [ ] **4.2 Domain `PriorityScore`** (rules §5) — lolos semua vektor.
- [ ] **4.3 `TodayQuery`** (rules §3) — hari ini menurut timezone user; hook `OccurrenceGenerator` (no-op dulu, diisi M6); `counts.missed` & `counts.inbox`. Satu round-trip DB untuk daftar (tanpa N+1).
- [ ] **4.4 `GET /today`** → `{ date, items: [{...todo, score, is_due_today}], counts: { missed, inbox } }`, sudah terurut.
- [ ] **4.5 `GET /todos?view=missed` versi awal** — todo biasa terlewat (recurring menyusul di M7), dengan `commitment_date`.

**GATE G4 → FE-04.**

## M4.5 — Staging lokal · `BE-13a` · Jalur B
VPS belum dipilih, jadi staging berjalan **di PC lokal dengan Docker** dulu.
- [ ] Dockerfile multi-stage final (image < 30 MB).
- [ ] `deploy/docker-compose.prod.yml` yang mirip produksi: api + postgres (+ web nanti), dijalankan lokal (`http://localhost`). Konfigurasi Traefik & Let's Encrypt disiapkan tapi baru aktif di M11.
- [ ] Migrasi otomatis saat start (`MIGRATE_ON_START=true`).
- [ ] `cmd/seed` mengisi akun demo + workspace + todo contoh.
- Tujuan: alur Login → Workspace → Todo → Hari Ini bisa dicoba dari build "produksi", bukan hanya mode dev.
---

## M5 — Timer + TimeLog (rules §8) · `BE-07` · Jalur B

- [ ] **5.0 Kontrak dulu** — tulis `docs/api/src/paths/<fitur>.yaml` + schema (termasuk kode error), `task generate`, stub `501` untuk operasi baru, merge → **GATE-K**: FE boleh mulai dengan mock
- [ ] **5.1 Migration `time_logs`**: id, user_id, workspace_id (wajib), todo_id (null), started_at, ended_at (null), duration_minutes (null), note, timestamps, deleted_at. Partial unique: satu baris berjalan (`ended_at IS NULL AND duration_minutes IS NULL AND deleted_at IS NULL`) per user.
- [ ] **5.2 Domain** — start (hentikan yang berjalan), stop, entri manual (`duration_minutes` wajib), durasi, flag `suspicious` (> 16 jam) di respons.
- [ ] **5.3 Service** — start dengan todo → workspace dari todo, todo `pending` → `in_progress`, dalam satu transaksi.
- [ ] **5.4 Endpoint** — `POST /timer/start` {todo_id?, workspace_id?}, `POST /timer/stop`, `GET /timer/current` (null kalau tidak ada), `POST /time-logs` (manual), `GET /time-logs?from=&to=&workspace_id=&todo_id=`, `PATCH /time-logs/{id}`, `DELETE /time-logs/{id}`.
- [ ] **5.5 Test konkurensi** — dua `start` bersamaan → tepat satu timer berjalan.

**GATE G5 → FE-07.**

---

## M6 — Rutin / todo berulang (rules §6) — milestone paling rumit · `BE-05` · Jalur A

- [ ] **6.0 Kontrak dulu** — tulis `docs/api/src/paths/<fitur>.yaml` + schema (termasuk kode error), `task generate`, stub `501` untuk operasi baru, merge → **GATE-K**: FE boleh mulai dengan mock
- [ ] **6.1 Test vector** — `docs/spec/test-vectors/recurrence.json` **lebih dulu** (≥20 kasus): harian interval, Senin–Jumat, mingguan multi-hari, bulanan tgl 31 → akhir bulan (Februari & kabisat), `BYDAY=1MO`, `-1FR`, UNTIL, COUNT.
- [ ] **6.2 Parser subset RRULE** — hanya subset rules §6 (`422 RECURRENCE_UNSUPPORTED` untuk lainnya); ringkasan teks Indonesia ("Setiap Senin & Kamis").
- [ ] **6.3 Iterator tanggal** — bungkus `rrule-go` + penyimpangan BYMONTHDAY 29–31 → hari terakhir bulan; lolos semua vektor.
- [ ] **6.4 Helper UUIDv5** — `OccurrenceID(templateID, date)` dengan `NS_OCCURRENCE`, `ChildID(parentID, sourceID)` dengan `NS_CHILD` (nilai namespace di rules §6).
- [ ] **6.5 `OccurrenceGenerator`** — buat kejadian dari `max(generated_until + 1, DTSTART, hari ini − 30)` s/d hari ini; salin judul/catatan/workspace/prioritas/estimasi + sub-task; `INSERT ... ON CONFLICT (id) DO NOTHING`; perbarui `generated_until`. Ganti no-op di `TodayQuery`, juga dipanggil di `view=missed` dan `/routines`.
- [ ] **6.6 Template API** — buat template via `POST /todos` dengan `recurrence_rule` + `scheduled_date` (DTSTART); `GET /routines` (template + ringkasan pola + `next_dates` 5 berikutnya, virtual); `PATCH` template = hanya untuk kejadian berikutnya; `POST /routines/{id}/stop` (UNTIL = hari ini).
- [ ] **6.7 Test konkurensi generator** — 2 goroutine paralel → jumlah kejadian sama dengan sekali jalan, tanpa error.

**GATE G6 → FE-05.**

---

## M7 — Terlewat + % kepatuhan (rules §2, §7) · `BE-06` · Jalur A

- [ ] **7.0 Kontrak dulu** — tulis `docs/api/src/paths/<fitur>.yaml` + schema (termasuk kode error), `task generate`, stub `501` untuk operasi baru, merge → **GATE-K**: FE boleh mulai dengan mock
- [ ] **7.1 Test vector** — `docs/spec/test-vectors/compliance.json` **lebih dulu** (≥15 kasus): tepat waktu, terlambat, terlewat, dilewati, dibatalkan, masih terbuka hari ini, baris `todo_misses`.
- [ ] **7.2 Migration `todo_misses`**: id (UUIDv5), user_id, todo_id, workspace_id, commitment_date, created_at, deleted_at.
- [ ] **7.3 Domain riwayat terlewat** — perubahan `scheduled_date`/`due_at` pada todo yang **sedang terlewat** → buat miss (`UUIDv5(NS_CHILD, "<todo_id>:miss:<YYYY-MM-DD>")`, idempotent). Dipasang di domain todo sehingga `PATCH`, `/reschedule`, dan sync semuanya tercakup.
- [ ] **7.4 `POST /todos/{id}/reschedule`** {scheduled_date?, due_at?}.
- [ ] **7.5 `view=missed` lengkap** — todo biasa + kejadian recurring, `commitment_date`, terbaru dulu; complete pada todo terlewat → respons `completed_late: true`.
- [ ] **7.6 `ComplianceCalculator`** — satu query agregat per lingkup (user / workspace / rutin), termasuk `todo_misses`; lolos vektor.
- [ ] **7.7 Integrasi** — `GET /compliance?scope=user|workspace|routine&id=` → `{ percent|null, window_from, window_to, on_time, late, missed, excluded }`; field `compliance_percent` di `/routines` dan `/workspaces`.
- AC: tidak ada kolom status `missed` di DB; edit tanggal todo terlewat lewat `PATCH` biasa juga mencatat miss.

**GATE G7 → FE-06.**

---

## M8a — Jurnal: catatan + mood · `BE-09a` · Jalur B
Butuh: M2. Tidak bergantung pada todo/kepatuhan.

- [ ] **8a.0 Kontrak dulu** — `docs/api/src/paths/journal.yaml` + schema untuk semua endpoint jurnal (termasuk `recap` yang nanti diisi M8b), `task generate`, stub `501`, merge → **GATE-K**
- [ ] **8a.1 Migration**
  - `journal_entries`: id, user_id, entry_date, mood (1–5 null), timestamps, deleted_at. Unique (user_id, entry_date) `WHERE deleted_at IS NULL`.
  - `journal_notes`: id, user_id, entry_date, content, label (`note|blocker|decision`), workspace_id (null), timestamps, deleted_at.
- [ ] **8a.2 Domain** — catatan, mood, entry dibuat otomatis saat pertama ditulis.
- [ ] **8a.3 Endpoint**
  - `GET /journal/{date}` → `{ date, mood, notes, recap }` — `recap` sementara kosong (`null`) sampai M8b
  - `PUT /journal/{date}/mood`; `POST /journal/{date}/notes`; `PATCH/DELETE /journal-notes/{id}`
  - `POST /journal-notes/{id}/to-todo` {scheduled_date?, workspace_id?} (`source = journal`, `source_note_id`) — butuh M3
  - `GET /journal?from=&to=&workspace_id=` → ringkasan per hari (sementara: mood & jumlah catatan; angka todo/menit ditambah di M8b)
- AC: hari tanpa data → struktur kosong, bukan 404.

**GATE G8a → FE-08 (catat cepat) dan FE-09 bagian catatan.**

---

## M8b — Jurnal: rekap otomatis (rules §9) · `BE-09b` · Jalur A
Butuh: M5, M7, M8a.

- [ ] **8b.1 `JournalRecap`** — todo selesai pada hari X (tandai terlambat), terlewat/dilewati dengan tanggal komitmen X + `todo_misses` hari X ("dijadwalkan ulang ke …"), menit per workspace (TimeLog), kepatuhan 1 hari (pakai ulang `ComplianceCalculator`).
- [ ] **8b.2 Isi `recap`** di `GET /journal/{date}` dan lengkapi ringkasan `GET /journal` (jumlah selesai, menit kerja) — tanpa mengubah kontrak dari M8a.
- [ ] **8b.3 Test** — rekap hari campuran (selesai tepat/terlambat, terlewat, dijadwalkan ulang, beberapa workspace).

**GATE G8b → FE-09 lengkap.**
---

## M9 — Laporan per workspace · `BE-10` · Jalur B

- [ ] **9.0 Kontrak dulu** — tulis `docs/api/src/paths/<fitur>.yaml` + schema (termasuk kode error), `task generate`, stub `501` untuk operasi baru, merge → **GATE-K**: FE boleh mulai dengan mock
- [ ] **9.1 Query agregat** — per workspace: total menit, selesai tepat/terlambat, terlewat, kepatuhan rentang; deret waktu `day|week|month` (`date_trunc` di zona user).
- [ ] **9.2 Endpoint** — `GET /reports/workspaces?from=&to=`, `GET /reports/workspaces/{id}?from=&to=&group_by=`; rentang maks 366 hari.
- [ ] **9.3 Test dengan data seed** — total menit = jumlah TimeLog rentang tsb.

**GATE G9 → FE-10.**

---

## M10a — Reminder + Web Push (rules §10) · `BE-11a` · Jalur B
Butuh: M3.

- [ ] **10a.0 Kontrak dulu** — `docs/api/src/paths/reminders.yaml` + `push.yaml`, `task generate`, stub `501`, merge → **GATE-K**
- [ ] **10a.1 Migration** — `reminders`: id, user_id, todo_id, remind_at (null untuk template), offset_minutes (null), sent_at, timestamps, deleted_at. `push_subscriptions`: id, user_id, platform (`web`), endpoint, p256dh, auth, user_agent, timestamps.
- [ ] **10a.2 Domain** — reminder absolut; tidak dikirim untuk todo `done/cancelled/skipped`.
- [ ] **10a.3 Endpoint** — `GET/POST /todos/{id}/reminders`, `DELETE /reminders/{id}`, `POST /push/subscriptions`, `DELETE /push/subscriptions/{id}`, `GET /push/vapid-public-key`.
- [ ] **10a.4 Worker** — goroutine ticker 30 detik di proses API, `FOR UPDATE SKIP LOCKED`, kirim Web Push, set `sent_at`, hapus subscription yang dibalas 404/410, berhenti rapi saat shutdown.
- AC: restart server tidak mengirim ulang reminder yang sudah `sent_at`. Android tidak butuh FCM (dijadwalkan lokal di HP).

**GATE G10a → FE-11.**

---

## M10b — Reminder offset untuk rutin · `BE-11b` · Jalur B
Butuh: M6, M10a.

- [ ] **10b.1 Reminder template** — reminder pada template rutin disimpan sebagai `offset_minutes` (rules §10); validasi: template hanya boleh reminder offset.
- [ ] **10b.2 Generator** — saat kejadian rutin dibuat (M6), offset template → reminder absolut (ID `UUIDv5(NS_CHILD, ...)`). Perubahan kecil di `OccurrenceGenerator` milik BE-A → PR terpisah (lihat aturan koordinasi).
- [ ] **10b.3 Test** — kejadian yang dibuat dua kali (paralel/backfill) tidak menggandakan reminder.

**GATE G10b → FE-11 bagian rutin.**
---

## M11 — Produksi web MVP · `BE-13` · Jalur B

- [ ] Pilih & siapkan VPS (masih OPEN di plan Bagian 5) — konfirmasi ke Pradana.
- [ ] Compose produksi: `api.<domain>`, `app.<domain>` (static dari FE), Postgres, Traefik + Let's Encrypt.
- [ ] Backup harian `pg_dump` + rotasi 7 hari + uji restore sekali (di server Linux; skrip boleh pakai shell/cron karena hanya jalan di server).
- [ ] Log JSON + uptime check `/health`.
- [ ] `deploy/README.md` langkah deploy.
- AC: `docker compose up` di VPS bersih → API HTTPS lolos `/health`.

**GATE G11 → Web MVP live.**

---

## M12 — Sync API untuk mobile · `BE-12` · Jalur A

- [ ] **12.0 Spesifikasi lebih dulu** — tulis `docs/spec/sync.md` (cursor, tombstone, last-write-wins, penjepitan waktu, urutan apply) dan minta review Pradana sebelum coding.
- [ ] **12.1 Kontrak** — setelah `sync.md` disetujui: `docs/api/src/paths/sync.yaml` + schema, `task generate`, stub `501`, merge → **GATE-K**: agent Mobile boleh mulai dengan mock
- [ ] **12.2 Cek `updated_at`** — pastikan trigger `set_updated_at` (M0.2) terpasang di semua tabel yang di-sync; tambah test.
- [ ] **12.3 `GET /sync/pull?cursor=`** — perubahan sejak cursor untuk workspaces, todos, todo_misses, time_logs, journal_entries, journal_notes, reminders; **termasuk tombstone**; batch maks 500 + `has_more`; cursor = (updated_at, id).
- [ ] **12.4 `POST /sync/push`** {changes: [{entity, op: upsert|delete, data, client_updated_at}]} — memanggil application service yang sama dengan REST; konflik last-write-wins berdasarkan `client_updated_at` yang dijepit ke `now()`; kejadian recurring UUIDv5 → `ON CONFLICT DO NOTHING`.
- [ ] **12.5 Test skenario** — dua device offline mengedit todo sama, hapus vs edit, push ulang (idempotent), jam HP salah.

**GATE G12 → Mobile MB-02 boleh mulai.**
