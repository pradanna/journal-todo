# Pembagian Kerja Agent — Jurnal + Todo Pintar

Dokumen ini memecah pekerjaan MVP menjadi task per fitur untuk tiga agent:

| Agent | File task | Stack | Fase |
|---|---|---|---|
| **BE** (Backend) — 2 agent: BE-A & BE-B | [`be.md`](be.md) + aturan tetap `backend/AGENTS.md` | Go + chi + sqlc/pgx + PostgreSQL, spec-first OpenAPI | 0 → 2 |
| **FE** (Web) | [`fe.md`](fe.md) | SvelteKit (Svelte 5, TypeScript) | 0 → 1 |
| **MB** (Mobile) | [`mobile.md`](mobile.md) | Flutter (Android dulu) | 3 |

## Sumber kebenaran (wajib dibaca sebelum mengerjakan task apa pun)
1. `../../plan-project.md` — keputusan produk. Poin **[FIX]** tidak boleh diubah tanpa konfirmasi Pradana.
2. `../spec/rules.md` — aturan bisnis bersama (hari ini, terlewat, skor prioritas, kepatuhan, recurring, timer). **Satu-satunya** tempat formula didefinisikan.
3. `../spec/api-conventions.md` — konvensi API (format respons, error, ID, tanggal, auth).
4. `../api/openapi.yaml` — kontrak endpoint. Ditulis & dijaga oleh agent BE. FE dan MB hanya boleh memakai endpoint yang ada di sini.
5. `../spec/test-vectors/*.json` — kasus uji bersama untuk formula di `rules.md`. BE dan MB wajib lolos vektor yang sama.

## Strategi: web-first
- **Fase 0 — Fondasi**: BE-00 lalu FE-00 (FE-00 boleh mulai setelah GATE G0 di `be.md`)
- **Fase 1 — Web MVP**: per fitur, BE menulis **kontrak** di `openapi.yaml` lebih dulu (GATE-K) → FE boleh mulai dengan mock → BE menyelesaikan endpoint (GATE penuh). BE boleh lanjut ke fitur berikutnya sementara FE mengejar.
- **Fase 2 — Siap mobile**: BE-12 (Sync API) + BE-13 (deploy produksi)
- **Fase 3 — Android**: MB-00 … MB-12

## Peta fitur & dependensi

| ID | Fitur | BE | FE Web | Mobile | Butuh fitur |
|---|---|---|---|---|---|
| F00 | Fondasi proyek | BE-00 | FE-00 | MB-00 | — |
| F01 | Auth dasar | BE-01 | FE-01 | MB-01 | F00 |
| F02 | Workspace | BE-02 | FE-02 | MB-03 | F01 |
| F03 | Todo dasar + sub-task | BE-03 | FE-03 | MB-04 | F02 |
| F04 | Layar Hari Ini + auto-prioritas | BE-04 | FE-04 | MB-04 | F03 |
| F05 | Todo berulang (Rutin) | BE-05 | FE-05 | MB-05 | F04 |
| F06 | Terlewat + % kepatuhan | BE-06 | FE-06 | MB-06 | F05 |
| F07 | Timer + TimeLog | BE-07 | FE-07 | MB-07 | F03 |
| F08 | Catat cepat | (pakai F03 & F09) | FE-08 | MB-08 | F03, F09 |
| F09 | Jurnal harian | BE-09a (catatan), BE-09b (rekap) | FE-09 | MB-09 | F02 (catatan); F06, F07 (rekap) |
| F10 | Laporan per workspace | BE-10 | FE-10 | MB-10 | F06, F07 |
| F11 | Reminder | BE-11a, BE-11b (offset rutin) | FE-11 | MB-11 | F03; F05 (offset rutin) |
| F12 | Sync API (offline-first) | BE-12 | — | MB-02 | semua entity |
| F13 | Staging lokal & deploy produksi | BE-13a (staging), BE-13 | FE-00.8 | — | F01 |
| F14 | Rilis Play Store | — | — | MB-12 | semua |

Urutan kerja BE (dua jalur paralel BE-A & BE-B, langkah detail & GATE): lihat [`be.md`](be.md) bagian "Pembagian jalur".

## Aturan kerja untuk semua agent
- **Satu task = satu branch/PR**: `feat/<task-id>-<slug>` (contoh `feat/BE-03-todo-crud`). Commit kecil dan jelas.
- **Definition of Done** (berlaku di semua task kecuali disebut lain):
  - Semua acceptance criteria di task terpenuhi
  - Ada test (BE: unit test application layer + integration test handler/repository ke Postgres asli; FE: unit test logika + Playwright untuk alur utama; MB: unit + widget test)
  - Lint & format lolos, build lolos di CI
  - BE: `openapi.yaml` diperbarui untuk setiap endpoint baru/berubah
  - Tidak ada secret di repo
- **Kalau menemukan kontradiksi** antara task, plan, dan spec: berhenti, tulis pertanyaannya di PR, jangan menebak.
- **Jangan mengubah formula** di kode tanpa mengubah `rules.md` + test vector-nya dalam PR yang sama (dan minta konfirmasi).
- Identifier kode dalam bahasa Inggris; teks UI dalam bahasa Indonesia (siapkan i18n agar bisa ditambah bahasa Inggris nanti).
- Semua tanggal "hari" mengikuti zona waktu user (lihat `rules.md` §1). Jangan pernah memakai `time.Now()` / `new Date()` langsung di logika bisnis — pakai abstraksi Clock agar bisa dites.
