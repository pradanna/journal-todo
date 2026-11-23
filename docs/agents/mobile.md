# Task Agent MB (Mobile — Flutter, Android dulu)

**Fase 3** — mulai setelah web MVP jalan dan BE-12 (Sync API) selesai.
Stack (usulan, keputusan final state management masih OPEN di plan Bagian 6): Flutter stable, Riverpod, drift (SQLite), dio + client Dart hasil generate dari `docs/api/openapi.yaml`, go_router, flutter_local_notifications, google_sign_in, workmanager.
Baca dulu: `README.md` di folder ini, `../spec/rules.md`, `../spec/api-conventions.md`, `../spec/sync.md`.

Perbedaan utama dengan web: **offline-first**. Semua baca/tulis ke database lokal dulu, lalu disinkronkan. Karena itu aturan di `rules.md` (Hari Ini, terlewat, skor prioritas, kepatuhan, kejadian recurring) **diimplementasikan ulang di Dart** dan wajib lolos `test-vectors/*.json` yang sama dengan BE.

---

## MB-00 — Fondasi
- Proyek di `mobile/`, flavor dev/prod, package name final menunggu nama app.
- Lint (`very_good_analysis` atau `flutter_lints` ketat), CI build + test.
- Generate client API dari OpenAPI; struktur folder per fitur (`features/<nama>/data|domain|presentation`).
- Tema (terang/gelap), komponen dasar senada dengan web, i18n Indonesia.
- Clock abstraction (sama seperti BE) untuk logika hari ini.

## MB-01 — Auth
Butuh: MB-00, BE-01
- Login email/password & Google (`google_sign_in`, client ID Android → `id_token` ke `/auth/google`), header `X-Client: android`.
- Refresh token di `flutter_secure_storage`; access token di memori. Auto-refresh saat 401.
- Kirim zona waktu perangkat saat login & saat berubah (`PATCH /me`).

## MB-02 — Database lokal + mesin sync
Butuh: MB-01, BE-12
- Skema drift yang mencerminkan entity server (UUID v7 dibuat di HP, updated_at, deleted_at).
- Outbox: setiap perubahan lokal dicatat sebagai antrian push.
- Sync: push outbox → pull sejak cursor → terapkan (LWW sesuai `sync.md`). Dipicu saat app dibuka, setelah perubahan (debounce), saat koneksi kembali, dan periodik via workmanager.
- Implementasi Dart dari `rules.md`: generator kejadian recurring (UUIDv5 + namespace yang sama), status terlewat, skor prioritas, kepatuhan. Test memakai `test-vectors/*.json`.
- Indikator status sync di UI (tersinkron / menunggu / gagal).
- AC: buat & edit todo dalam mode pesawat → muncul di web setelah online; kejadian recurring yang dibuat HP dan server tidak dobel.

## MB-03 — Workspace
Butuh: MB-02 · Daftar, buat/edit, arsip, urutkan; picker workspace.

## MB-04 — Todo + layar Hari Ini
Butuh: MB-03
- Layar Hari Ini (layar awal), Inbox, Semua Todo, detail todo dengan sub-task.
- Gestur: geser kanan = selesai, geser kiri = lewati.
- Urutan dari skor lokal (rules §5).

## MB-05 — Rutin
Butuh: MB-04 · Editor pola berulang (preset sama dengan web FE-05), daftar rutin + kepatuhan.

## MB-06 — Terlewat + kepatuhan
Butuh: MB-05 · Daftar terlewat dengan aksi selesaikan terlambat / lewati / jadwalkan ulang (mencatat `todo_misses` secara lokal, rules §7); badge kepatuhan.

## MB-07 — Timer + notifikasi berjalan
Butuh: MB-04
- Start/stop timer, mini-bar di atas layar selama timer jalan.
- Notifikasi ongoing dengan chronometer (`usesChronometer`, `when = started_at`) + tombol Stop di notifikasi — tanpa foreground service.
- Android 16+: ajukan sebagai Live Update (platform channel Kotlin: `NotificationCompat.Builder#requestPromotedOngoing`, izin manifest `POST_PROMOTED_NOTIFICATIONS`) → chip timer di status bar. Perangkat lama jatuh ke notifikasi biasa.
- AC: timer tetap benar setelah app di-kill & dibuka lagi (dihitung dari `started_at`).

## MB-08 — Catat cepat
Butuh: MB-04, MB-09
- FAB `+` → bottom sheet satu baris + pilihan jenis (Todo/Catatan/Blocker/Keputusan).
- Bonus: shortcut dari ikon app (long-press, App Shortcuts) dan Quick Settings tile untuk "Catat cepat".

## MB-09 — Jurnal harian
Butuh: MB-06, MB-07 · Halaman per hari: rekap otomatis (dihitung lokal) + catatan + mood; histori kalender; filter workspace; catatan → todo.

## MB-10 — Laporan
Butuh: MB-09 · Laporan per workspace. Boleh dari API (online saja) untuk MVP mobile — tampilkan pesan bila offline.

## MB-11 — Reminder
Butuh: MB-05
- Reminder dijadwalkan **lokal** (`flutter_local_notifications`, zonedSchedule) dari data reminder di DB lokal → tetap jalan offline. Dijadwalkan ulang setiap sync.
- Izin notifikasi Android 13+ diminta saat pertama menambah reminder.
- Alarm tepat waktu: cek kebijakan Play terbaru untuk `SCHEDULE_EXACT_ALARM` / `USE_EXACT_ALARM`; default pakai alarm tidak-tepat (inexact) kecuali user mengaktifkan izin alarm tepat.
- Tidak memakai FCM di MVP.

## MB-12 — Persiapan rilis Play Store
Butuh: semua
- Nama app, ikon adaptif, splash, screenshot listing.
- Halaman kebijakan privasi (di web) + formulir Data Safety.
- **Hapus akun dari dalam app + halaman web** (wajib Play) — butuh task BE tambahan dari bagian Auth yang ditunda.
- Target API level terbaru yang diwajibkan Play, signing (Play App Signing), track internal testing → closed testing.
