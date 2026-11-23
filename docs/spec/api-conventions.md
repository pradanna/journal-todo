# Konvensi API

## Umum
- Base path: `/api/v1`. Auth di `/api/v1/auth/*`. Health: `GET /health` (di luar /api).
- JSON `snake_case`. Encoding UTF-8.
- Kontrak resmi: `docs/api/openapi.yaml` (OpenAPI 3.0.3), ditulis **spec-first** oleh agent BE; kode server Go digenerate darinya (oapi-codegen) sehingga server selalu sesuai kontrak. FE generate tipe dari file ini (`openapi-typescript`), MB generate client Dart darinya.

## ID
- Semua entity memakai UUID (kolom `uuid` di Postgres). ID baru = **UUID v7** (berurutan menurut waktu).
- Saat create, client **boleh** mengirim `id` (UUID v7 buatan client, wajib untuk mobile offline). Kalau tidak dikirim, server membuatnya.
- Create dengan `id` yang sudah ada milik user yang sama → idempotent: kembalikan data yang ada dengan `200` (bukan error). Milik user lain → `409`.
- Kejadian recurring memakai UUID v5 (lihat `rules.md` §6).

## Waktu
- Timestamp: RFC 3339 UTC, contoh `2026-11-23T03:15:00Z`.
- Tanggal "hari": `YYYY-MM-DD` tanpa zona (tanggal lokal user).
- Server menentukan "hari ini" dari `users.timezone`.

## Respons
- Objek tunggal: dikembalikan langsung (tanpa pembungkus).
- List: `{ "data": [...], "next_cursor": "opaque-string|null" }`. Pagination berbasis cursor, parameter `?limit=` (default 50, max 200) & `?cursor=`.
- Error:
  ```json
  { "error": { "code": "TODO_NOT_FOUND", "message": "Todo tidak ditemukan", "fields": { "title": "wajib diisi" } } }
  ```
  - `code` stabil (UPPER_SNAKE), dipakai client untuk logika. `message` boleh berubah. `fields` hanya untuk error validasi.
- Kode HTTP: 200, 201, 204, 400 (payload rusak), 401, 403, 404, 409 (konflik), 422 (validasi/aturan bisnis), 429, 500.
- Resource milik user lain dikembalikan sebagai **404** (bukan 403) agar tidak membocorkan keberadaannya.

## Soft delete
- `DELETE` = isi `deleted_at`. Data terhapus tidak muncul di endpoint biasa.
- Workspace tidak bisa di-DELETE (pakai archive) — lihat plan Bagian 2.
- Endpoint sync (BE-12) mengembalikan baris terhapus sebagai tombstone.

## Auth
- Header `Authorization: Bearer <access_token>` (JWT, 15 menit).
- Header `X-Client: web | android` wajib di endpoint auth.
  - `web`: refresh token dikirim sebagai cookie `refresh_token` (`HttpOnly; Secure; SameSite=Lax; Path=/api/v1/auth`), tidak ada di body.
  - `android`: refresh token ada di body respons.
- Refresh token dirotasi setiap kali dipakai (yang lama langsung tidak berlaku).

## CORS
- Origin web diizinkan dari env `CORS_ALLOWED_ORIGINS` (dev: `http://localhost:5173`). `credentials: true` untuk cookie refresh.

## Rate limit
- Endpoint auth: 10 request/menit per IP. Lainnya: 120 request/menit per user. Melebihi → `429` + header `Retry-After`.
