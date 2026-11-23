# Urutan Pengerjaan Agent BE

Panduan langkah demi langkah untuk agent BE. Detail tiap task (tabel, endpoint, AC) ada di [`be.md`](be.md); dokumen ini menjawab **urutan**, **langkah di dalam task**, dan **kapan FE boleh mulai**.

## Cara memakai dokumen ini
- Kerjakan milestone **berurutan**. Di dalam milestone, ikuti urutan langkah (sudah memperhitungkan dependensi).
- Centang `[x]` setiap langkah yang sudah di-merge.
- Setiap milestone ditutup dengan **GATE**: daftar yang wajib ada sebelum FE diberi tahu bahwa fitur itu siap. Setelah GATE lolos, perbarui kolom status di tabel paling bawah.
- Ukuran: **S** ≈ setengah hari, **M** ≈ 1–2 hari, **L** ≈ 3–4 hari kerja agent (perkiraan kasar untuk perencanaan, bukan target).
- Satu langkah bertanda ⟂ boleh dikerjakan paralel dengan langkah sebelumnya.

## Pola yang dipakai di SETIAP fitur (urutan baku)
Ikuti urutan ini di setiap milestone M2 ke atas, supaya fitur selalu lengkap dari DB sampai kontrak:
1. **Migration** (`migrations/0000NN_<fitur>.up/down.sql`) — tabel, index, partial unique `WHERE deleted_at IS NULL`
2. **Domain** (`internal/domain/<fitur>/`) — entity, value object, aturan/transisi, interface repository, error domain. Unit test di sini (tanpa DB).
3. **Repository** (`internal/infrastructure/postgres/<fitur>_repository.go`) — implementasi GORM. Integration test ke Postgres asli.
4. **Application service** (`internal/application/<fitur>/`) — use case, cek kepemilikan (user_id), memakai Clock. Unit test dengan repository fake.
5. **Handler + route** (`internal/interfaces/http/handlers/<fitur>_handler.go`) — DTO request/response, validasi, mapping error. Integration test HTTP (register user → panggil endpoint).
6. **OpenAPI** — tambahkan path & schema di `docs/api/openapi.yaml`, lint lolos.
7. **Seed dev** (opsional) — data contoh di `scripts/seed` agar FE bisa mencoba.

Aturan umum:
- Semua query milik user **wajib** difilter `user_id` di repository (bukan hanya di handler).
- Jangan pernah `time.Now()` di domain/application — pakai `Clock`.
- Logika aturan bisnis di **domain/application**, bukan di handler, karena BE-12 (sync) akan memanggil logika yang sama.

---

## M0 — Fondasi (BE-00) · total ≈ M–L

- [ ] **0.1 Bersihkan scaffold lama** (S)
  - Hapus: `internal/domain/entry.go`, `internal/application/entry_service.go`, `internal/application/time_helper.go`, `internal/infrastructure/postgres/entry_repository.go`, `internal/interfaces/http/handlers/entry_handler.go`, `migrations/000001_init.*`.
  - Pindahkan file tersisa ke pola per fitur (`internal/domain/user/`, `internal/domain/workspace/`, dst). Kode auth & workspace lama boleh jadi referensi tapi akan ditulis ulang di M1/M2.
  - Ganti module path ke `github.com/pradanna/jurnal-todo`.
  - Buat `migrations/000001_init.up.sql` berisi `CREATE EXTENSION IF NOT EXISTS pgcrypto; CREATE EXTENSION IF NOT EXISTS citext;`
- [ ] **0.2 Base model & tipe bersama** (S) — `internal/domain/shared`: `BaseModel` (UUID, timestamps, `gorm.DeletedAt`), tipe `Date` (tanggal tanpa zona, marshal `YYYY-MM-DD`, scan/value ke kolom `DATE`).
- [ ] **0.3 Clock & zona waktu** (S) — interface `Clock`, `RealClock`, `FakeClock`; helper `TodayIn(tz)`, `LocalDate(t, tz)`. Unit test termasuk pergantian hari di zona +07:00.
- [ ] **0.4 Config** (S) — semua env (DB, JWT, Google client IDs, CORS, VAPID, APP_ENV) divalidasi saat start; gagal start kalau env wajib kosong.
- [ ] **0.5 Infrastruktur HTTP** (M) — prefix `/api/v1`, `GET /health`, error handler global + package `apperror`, validator, middleware (request ID, slog, recover, CORS, rate limit), graceful shutdown.
- [ ] **0.6 Tooling & CI** (M) ⟂ — Makefile, `docker-compose.dev.yml` (Postgres), testcontainers untuk integration test, golangci-lint, GitHub Actions (lint, test, build), cek "tidak ada `time.Now()` di domain/application".
- [ ] **0.7 OpenAPI kerangka** (S) — `docs/api/openapi.yaml` dengan komponen bersama + `/health`; Scalar/Swagger UI di `/docs` saat development; lint OpenAPI di CI.

**GATE G0 → FE-00 boleh mulai**
- `make run` menjalankan API kosong + `/health` OK
- `openapi.yaml` valid & bisa di-generate tipe TypeScript
- CI hijau

---

## M1 — Auth dasar (BE-01) · total ≈ L

- [ ] **1.1 Migration** — `users`, `refresh_tokens` (lihat be.md BE-01.1).
- [ ] **1.2 Domain user** — entity User, validasi email/password/timezone (IANA), error `EMAIL_TAKEN`, `INVALID_CREDENTIALS`.
- [ ] **1.3 Token service** — JWT access 15 menit; refresh token acak (hash sha256 di DB), rotasi + `replaced_by`, deteksi pemakaian ulang → cabut seluruh rantai. Unit test rotasi & reuse.
- [ ] **1.4 Register & login password** — bcrypt, respons token, perilaku `X-Client` (web → cookie HttpOnly, android → body).
- [ ] **1.5 Refresh, logout, logout-all** — termasuk cookie path `/api/v1/auth`.
- [ ] **1.6 Login Google** — verifikasi `id_token` (multi audience), `email_verified`, 3 cabang penautan akun (penautan mengosongkan password & mencabut token lama). Test ketiga cabang dengan verifier palsu.
- [ ] **1.7 Middleware auth & `/me`** — muat user (termasuk timezone) ke context; `GET /me`, `PATCH /me`.
- [ ] **1.8 Rate limit auth** — 10/menit per IP untuk `/auth/*`.
- [ ] **1.9 OpenAPI auth** + contoh request di dokumentasi.

**GATE G1 → FE-01 boleh mulai**
- Register, login, Google, refresh (cookie), logout, `/me` terdokumentasi & teruji
- CORS dengan `credentials` jalan dari `http://localhost:5173`

---

## M2 — Workspace (BE-02) · total ≈ M

- [ ] **2.1 Migration `workspaces`**
- [ ] **2.2 Domain** — validasi nama/warna/ikon, aturan arsip, error `WORKSPACE_ARCHIVED`.
- [ ] **2.3 Repository + service** — CRUD, arsip/batal arsip, urutan (`PUT /workspaces/order`).
- [ ] **2.4 Handler + OpenAPI** — tanpa DELETE.
- [ ] **2.5 Helper kepemilikan** — `workspace.EnsureActiveOwned(userID, id)` yang akan dipakai todo, timer, catatan.

**GATE G2 → FE-02 boleh mulai**

---

## M3 — Todo dasar + sub-task (BE-03) · total ≈ L

- [ ] **3.1 Migration `todos`** — semua kolom termasuk kolom recurring (`recurrence_rule`, `recurrence_parent_id`, `generated_until`) supaya M6 tidak perlu ALTER besar.
- [ ] **3.2 Domain todo** — status & transisi (rules §2), aturan sub-task 1 level, induk selesai → sub-task ikut selesai, turunan `IsMissed(today, now)`, `CommitmentDate()`, `CompletedLate()`. Unit test tabel lengkap.
- [ ] **3.3 Repository** — termasuk ambil todo + sub-task dalam satu query (preload), filter list.
- [ ] **3.4 Service + handler CRUD** — `POST/GET/PATCH/DELETE /todos`, `GET /todos?view=inbox|all`.
- [ ] **3.5 Aksi status** — `/start`, `/complete`, `/reopen`, `/skip`, `/cancel`.
- [ ] **3.6 OpenAPI todo** — schema Todo dipakai ulang oleh semua view berikutnya.

**GATE G3 → FE-03 boleh mulai**

---

## M4 — Layar Hari Ini + auto-prioritas (BE-04) · total ≈ M

- [ ] **4.1 Test vector prioritas** — tulis `docs/spec/test-vectors/priority.json` **lebih dulu** (≥15 kasus).
- [ ] **4.2 Domain `PriorityScore`** — rules §5, lolos semua vektor.
- [ ] **4.3 `TodayQuery`** — rules §3 (hari ini menurut timezone user), hook `OccurrenceGenerator` no-op, hitung `counts.missed` & `counts.inbox`.
- [ ] **4.4 `GET /today` + OpenAPI**
- [ ] **4.5 `GET /todos?view=missed` versi awal** — todo biasa yang terlewat (recurring menyusul di M6). Membantu FE menyiapkan halaman Terlewat lebih awal.

**GATE G4 → FE-04 boleh mulai**

## M4.5 — Staging (sebagian BE-13) · total ≈ M
- [ ] Dockerfile final, `deploy/docker-compose.prod.yml` (api + postgres + Traefik), deploy ke VPS sebagai `api-staging.<domain>`.
- [ ] Migrasi otomatis saat start, backup harian sederhana.
- Tujuan: FE & Pradana bisa mencoba alur Login → Workspace → Todo → Hari Ini di server sungguhan sejak dini.

---

## M5 — Timer + TimeLog (BE-07) · total ≈ M

- [ ] **5.1 Migration `time_logs`** + partial unique "satu timer berjalan per user".
- [ ] **5.2 Domain** — start (hentikan yang berjalan), stop, entri manual, durasi, flag `suspicious` (>16 jam).
- [ ] **5.3 Service** — start dengan todo → workspace dari todo, todo pending → in_progress (dalam satu transaksi).
- [ ] **5.4 Handler + OpenAPI** — `/timer/start|stop|current`, CRUD `/time-logs`.
- [ ] **5.5 Test konkurensi** — dua `start` bersamaan menghasilkan tepat satu timer berjalan.

**GATE G5 → FE-07 boleh mulai**

---

## M6 — Todo berulang / Rutin (BE-05) · total ≈ L (milestone paling rumit)

- [ ] **6.1 Test vector recurrence** — `recurrence.json` (≥20 kasus) **lebih dulu**, termasuk tgl 31 → akhir bulan, Februari kabisat, `1MO`, `-1FR`, UNTIL, COUNT.
- [ ] **6.2 Parser subset RRULE** — validasi (tolak di luar subset: `422 RECURRENCE_UNSUPPORTED`), ringkasan teks Indonesia ("Setiap Senin & Kamis").
- [ ] **6.3 Iterator tanggal** — bungkus `rrule-go` + penyimpangan BYMONTHDAY 29–31; lolos semua vektor.
- [ ] **6.4 UUIDv5 helper** — `OccurrenceID(templateID, date)`, `ChildID(occurrenceID, sourceID)` dengan namespace dari rules §6.
- [ ] **6.5 OccurrenceGenerator** — backfill maks 30 hari, `generated_until`, salin field + sub-task, `ON CONFLICT DO NOTHING`; ganti no-op di `TodayQuery`.
- [ ] **6.6 Template API** — create via `POST /todos` dengan `recurrence_rule`, `GET /routines` (+ `next_dates` virtual), edit template (hanya kejadian berikutnya), `POST /routines/{id}/stop`.
- [ ] **6.7 Test konkurensi generator** — 2 goroutine paralel → hasil sama dengan 1 kali jalan.
- [ ] **6.8 OpenAPI rutin**

**GATE G6 → FE-05 boleh mulai**

---

## M7 — Terlewat + kepatuhan (BE-06) · total ≈ L

- [ ] **7.1 Test vector kepatuhan** — `compliance.json` (≥15 kasus) **lebih dulu**, termasuk selesai terlambat, lewati, batal, masih terbuka hari ini, dan baris `todo_misses`.
- [ ] **7.2 Migration `todo_misses`**
- [ ] **7.3 Domain riwayat terlewat** — perubahan `scheduled_date`/`due_at` pada todo yang sedang terlewat → buat miss (UUIDv5, idempotent). Dipasang di domain todo sehingga `PATCH`, `/reschedule`, dan sync semuanya tercakup.
- [ ] **7.4 `POST /todos/{id}/reschedule`**
- [ ] **7.5 `view=missed` lengkap** — todo biasa + kejadian recurring, `commitment_date`, terbaru dulu; respons complete menandai `completed_late`.
- [ ] **7.6 `ComplianceCalculator`** — satu query agregat per lingkup (user / workspace / rutin), termasuk `todo_misses`; lolos vektor.
- [ ] **7.7 Integrasi** — `GET /compliance`, field `compliance_percent` di `/routines` dan `/workspaces`.
- [ ] **7.8 OpenAPI**

**GATE G7 → FE-06 boleh mulai**

---

## M8 — Jurnal harian (BE-09) · total ≈ M–L

- [ ] **8.1 Migration** `journal_entries`, `journal_notes`.
- [ ] **8.2 Domain** — catatan (label, workspace opsional), mood 1–5, entry dibuat otomatis.
- [ ] **8.3 `JournalRecap`** — rules §9: todo selesai (tandai terlambat), terlewat/dilewati per tanggal komitmen + `todo_misses`, menit per workspace, kepatuhan 1 hari. Pakai ulang `ComplianceCalculator`.
- [ ] **8.4 Endpoint** — `GET /journal/{date}`, `PUT .../mood`, CRUD catatan, `POST /journal-notes/{id}/to-todo`, `GET /journal?from&to&workspace_id`.
- [ ] **8.5 OpenAPI**

**GATE G8 → FE-09 dan FE-08 (catat cepat) boleh mulai**

---

## M9 — Laporan per workspace (BE-10) · total ≈ M

- [ ] **9.1 Query agregat** — menit, selesai tepat/terlambat, terlewat, kepatuhan per workspace untuk rentang; deret waktu day/week/month (`date_trunc` di zona user).
- [ ] **9.2 Endpoint + validasi rentang (maks 366 hari)**
- [ ] **9.3 Test dengan data seed** — total menit = jumlah TimeLog.
- [ ] **9.4 OpenAPI**

**GATE G9 → FE-10 boleh mulai**

---

## M10 — Reminder / Web Push (BE-11) · total ≈ M–L

- [ ] **10.1 Migration** `reminders`, `push_subscriptions`.
- [ ] **10.2 Domain** — reminder absolut vs offset (template), aturan tidak mengirim untuk todo selesai/batal/dilewati.
- [ ] **10.3 Endpoint** — CRUD reminder per todo, langganan push, `GET /push/vapid-public-key`.
- [ ] **10.4 Generator** — salin offset template → reminder absolut saat kejadian dibuat (rules §10).
- [ ] **10.5 Worker** — ticker 30 detik, `FOR UPDATE SKIP LOCKED`, kirim Web Push, set `sent_at`, hapus subscription 404/410, berhenti rapi saat shutdown.
- [ ] **10.6 OpenAPI**

**GATE G10 → FE-11 boleh mulai**

---

## M11 — Produksi web MVP (BE-13) · total ≈ M

- [ ] Compose produksi lengkap: `api.<domain>`, `app.<domain>` (static dari FE-00.8), Postgres, Traefik + Let's Encrypt.
- [ ] Backup harian `pg_dump` + rotasi 7 hari + uji restore sekali.
- [ ] Log & monitoring minimal (log JSON, uptime check `/health`).
- [ ] `deploy/README.md`.

**GATE G11 → Web MVP live** 🎉

---

## M12 — Sync API untuk mobile (BE-12) · total ≈ L

- [ ] **12.1 Spesifikasi** — tulis `docs/spec/sync.md` dulu (cursor, tombstone, LWW, jepit waktu, urutan apply) dan minta review Pradana.
- [ ] **12.2 `updated_at` dijamin server** — trigger DB atau hook di semua tabel yang di-sync.
- [ ] **12.3 `GET /sync/pull`** — semua entity (termasuk `todo_misses`), tombstone, batch 500, `has_more`.
- [ ] **12.4 `POST /sync/push`** — memanggil application service yang sama dengan REST; recurring UUIDv5 `ON CONFLICT DO NOTHING`.
- [ ] **12.5 Test skenario** — dua device offline edit todo sama, hapus vs edit, push ulang (idempotent), jam HP salah.
- [ ] **12.6 OpenAPI**

**GATE G12 → Fase Mobile (MB-02) boleh mulai**

---

## Ringkasan urutan & status

| # | Milestone | Task | Ukuran | Membuka FE | Status |
|---|---|---|---|---|---|
| M0 | Fondasi | BE-00 | M–L | FE-00 | ⬜ |
| M1 | Auth dasar | BE-01 | L | FE-01 | ⬜ |
| M2 | Workspace | BE-02 | M | FE-02 | ⬜ |
| M3 | Todo dasar | BE-03 | L | FE-03 | ⬜ |
| M4 | Hari Ini | BE-04 | M | FE-04 | ⬜ |
| M4.5 | Staging | BE-13 (sebagian) | M | — | ⬜ |
| M5 | Timer | BE-07 | M | FE-07 | ⬜ |
| M6 | Rutin | BE-05 | L | FE-05 | ⬜ |
| M7 | Terlewat + kepatuhan | BE-06 | L | FE-06 | ⬜ |
| M8 | Jurnal | BE-09 | M–L | FE-09, FE-08 | ⬜ |
| M9 | Laporan | BE-10 | M | FE-10 | ⬜ |
| M10 | Reminder | BE-11 | M–L | FE-11 | ⬜ |
| M11 | Produksi | BE-13 | M | — (web live) | ⬜ |
| M12 | Sync API | BE-12 | L | MB-02 | ⬜ |

Status: ⬜ belum · 🟡 dikerjakan · ✅ GATE lolos
