# AGENTS.md — Aturan tetap Backend (Go)

Berlaku untuk **setiap** perubahan di `backend/`, termasuk perbaikan kecil di luar milestone. Roadmap & detail task ada di `docs/agents/be.md`; aturan bisnis di `docs/spec/rules.md`; konvensi API di `docs/spec/api-conventions.md`.

## Sebelum mulai
- Baca milestone yang sedang dikerjakan di `docs/agents/be.md` dan bagian `rules.md` yang dirujuknya.
- Cek draft PR agent lain — satu milestone hanya dikerjakan satu agent. Buka draft PR `[BE-A] …` / `[BE-B] …` saat mulai.
- Kalau instruksi, `be.md`, `rules.md`, dan plan saling bertentangan: **berhenti dan tanya**, jangan menebak.

## Perintah (Task — sama di Windows & Linux)
Development utama di **Windows** (PowerShell); server & CI di Linux. Pakai `task`, bukan `make`.

| Perintah | Kapan |
|---|---|
| `task tools` | Sekali di awal / saat versi tool berubah — cek & beri tahu cara install |
| `task db:up` / `task db:down` | Menyalakan/mematikan Postgres dev (Docker Desktop) |
| `task generate` | Setelah mengubah kontrak (`docs/api/src/`) atau query SQL (`queries/`) |
| `task migrate:create -- <nama>` | Membuat migration baru (nama timestamp) |
| `task migrate:up` / `task migrate:down` | Mencoba migration (down wajib diuji) |
| `task run` | Menjalankan API lokal |
| `task seed` | Mengisi data contoh |
| `task lint` / `task test` / `task test-integration` | Sebelum push |

**Sebelum menyatakan task selesai, wajib: `task generate`, `task lint`, `task test`, `task test-integration` lolos dan `git status` bersih dari hasil generate yang belum di-commit.**

### Kompatibilitas Windows
- Isi `Taskfile.yml` tidak boleh memakai perintah khusus bash (`grep`, `sed`, `rm -rf`, `xargs`, dll). Butuh logika → program Go kecil.
- Tidak ada skrip `.sh` di `backend/` (kecuali skrip yang hanya jalan di server, di `deploy/`).
- Path di kode memakai `filepath.Join`, bukan string dengan `/` atau `\`.
- Line ending LF dipaksa lewat `.gitattributes` — jangan diubah.

## Struktur
```
backend/
  api/oapi-codegen.yaml            konfigurasi generator
  cmd/api/main.go                  hanya bootstrap; wiring per fitur di internal/app/
  internal/
    api/gen/                       HASIL GENERATE (oapi-codegen) — jangan diedit
    app/<fitur>.go                 wiring dependensi per fitur
    domain/<fitur>/                entity, aturan, interface repository, error domain
    domain/shared/                 Date, NewID (UUID v7), UUIDv5, Clock
  cmd/migrate/  cmd/seed/          alat bantu (Go, lintas OS)
    application/<fitur>/           use case
    infrastructure/postgres/
      queries/<fitur>.sql          SQL untuk sqlc — satu file per fitur
      sqlcgen/                     HASIL GENERATE (sqlc) — jangan diedit
      <fitur>_repository.go
    interfaces/http/
      server.go                    struct Server meng-embed handler tiap fitur
      <fitur>_handler.go
  migrations/                      YYYYMMDDHHMMSS_<nama>.up.sql / .down.sql
docs/api/src/                      SUMBER kontrak (diedit di sini)
docs/api/openapi.yaml              HASIL BUNDLE — jangan diedit
```

## Kontrak & generate
- Kontrak ditulis **lebih dulu** di `docs/api/src/paths/<fitur>.yaml` + `schemas/<fitur>.yaml`, lalu `task generate`.
- Operasi baru tanpa implementasi → stub yang membalas `501 NOT_IMPLEMENTED`, supaya `main` selalu compile.
- Jangan mengedit `internal/api/gen/`, `sqlcgen/`, atau `docs/api/openapi.yaml` secara manual.
- Mengubah kontrak yang sudah dipakai FE (rename/hapus field, ubah tipe) = **breaking change**: sebutkan jelas di deskripsi PR.

## Data & query
- **Setiap** query data user difilter `user_id` di SQL. Resource milik user lain → 404, bukan 403.
- **Soft delete manual**: setiap SELECT wajib `deleted_at IS NULL`; hapus = `UPDATE … SET deleted_at = now()`. Hard delete hanya untuk kasus yang disebut eksplisit di `be.md`.
- Semua timestamp di DB `timestamptz` (UTC). Konversi ke zona user hanya di tepi: respons dan perhitungan "hari ini". Kolom "hari" bertipe `DATE`.
- ID baru: `shared.NewID()` (UUID v7). ID kejadian rutin & turunannya: helper UUIDv5 — jangan membuat sendiri.
- Endpoint list wajib pagination (default 50, maks 200). Tidak boleh ada N+1 query.
- Query agregat (kepatuhan, laporan, rekap): sertakan hasil `EXPLAIN ANALYZE` dengan data seed di deskripsi PR.

## Migration
- Buat dengan `task migrate:create -- <nama>`.
- Nomor migration di PR harus **lebih besar** dari migration terakhir di `main` — kalau tertinggal karena branch lain merge lebih dulu, rename ke timestamp baru sebelum merge.
- Setiap migration punya `down` yang benar dan diuji.
- Setiap foreign key diberi index. Setiap tabel yang di-sync punya `created_at`, `updated_at` (trigger `set_updated_at`), `deleted_at`.
- Perubahan yang merusak data (rename/drop kolom) dipecah: tambah → pindahkan data → hapus, di PR terpisah.
- Jangan mengedit migration yang sudah di-merge.

## Lapisan & logika
- Aturan bisnis hanya di `domain/` dan `application/`. Handler hanya: terima request → panggil service → petakan hasil/error.
- Transaksi dibuka di application service (`WithTx`); repository menerima `tx`, tidak membuka transaksi sendiri.
- `context.Context` diteruskan ke semua fungsi yang menyentuh DB atau jaringan.
- **Dilarang** `time.Now()` di `domain/` & `application/` — pakai `Clock`. (Dicek golangci-lint `forbidigo`.)
- Tidak ada state global/variabel package yang bisa berubah. Dependensi lewat constructor.
- Formula (`rules.md`) tidak boleh diubah tanpa mengubah `rules.md` + test vector di PR yang sama dan konfirmasi Pradana.

## Error & logging
- Bungkus error dengan konteks: `fmt.Errorf("create todo: %w", err)`. Error domain pakai sentinel/tipe di `domain/<fitur>/errors.go` dengan kode stabil (`TODO_NOT_FOUND`).
- Detail internal (SQL, stack trace, pesan library) **tidak boleh** sampai ke respons — klien hanya menerima format error standar.
- Log sekali di batas HTTP, lewat `slog`, dengan `request_id` dan `user_id`.
- **Jangan pernah** me-log password, token (access/refresh/id_token), cookie, atau body endpoint `/auth/*`.

## Testing
- Table-driven test. Nama: `TestXxx_<kondisi>_<hasil>`.
- Domain & application: coverage ≥ 80%.
- Integration test ke Postgres asli (testcontainers), bukan mock DB.
- **Setiap endpoint** punya test akses lintas user: user B tidak bisa membaca/mengubah data user A (harus 404).
- Tidak ada `time.Sleep` di test — pakai `FakeClock`. Tidak bergantung urutan eksekusi test.
- Formula wajib diuji dengan `docs/spec/test-vectors/*.json`.

## Disiplin perubahan
- Kerjakan hanya scope task. Refactor di luar scope → PR terpisah.
- Perubahan pada kode milik milestone agent lain → PR kecil terpisah, di-review lebih dulu.
- Dependency baru hanya dengan alasan tertulis di PR (kenapa, alternatif yang dipertimbangkan).
- Branch: `feat/BE-<id>-<slug>`, `fix/BE-<slug>`. Commit: `<tipe>(<fitur>): <ringkasan>` — tipe `feat|fix|refactor|test|docs|chore`.
- Tidak ada secret di repo; nilai contoh di `.env.example`.

## Template deskripsi PR
```
## Apa yang berubah
## Milestone & langkah            (mis. M6 — 6.3, 6.4)
## Kontrak                        (endpoint baru/berubah; breaking change?)
## Cara mengetes
## Status GATE                    (GATE-K / GATE penuh lolos?)
## Catatan untuk FE / agent lain
```
