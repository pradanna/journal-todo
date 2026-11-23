# Plan Project — App Jurnal + Todo Pintar (nama belum final)

Status per poin: **[FIX]** sudah diputuskan, **[OPEN]** masih perlu didiskusikan/diputuskan.

## 1. Overview & Tujuan
- [FIX] Target user: siapa saja yang kerja per-task/per-hari (bukan cuma freelance developer) — scope diperluas dari rencana awal
- [FIX] Dua pilar fitur utama:
  1. **Jurnal kerja** — rekap otomatis todo & jam kerja per hari + catatan (blocker, keputusan, dll), bisa difilter per Workspace (klien/project)
  2. **Todo list pintar**:
     - Auto-prioritas (urutan otomatis berdasar deadline/urgency)
     - Reminder & recurring (todo berulang harian/mingguan/bulanan) + % kepatuhan
     - Nyambung otomatis ke jurnal (todo bisa dibuat langsung dari catatan jurnal)
     - Dibantu AI (opsional — breakdown sub-task otomatis / saran dari histori jurnal)
- [FIX] Value proposition / diferensiator dari notes app biasa (Notion, Keep, dll): tracking waktu & histori kerja per klien/project yang rapi, bukan flat notes
- [FIX] General-purpose, rencana rilis ke Google Play
- [FIX] Model bisnis: gratis dulu, belum ada billing/monetisasi
- [OPEN] Tagline/value statement singkat buat listing Play Store — belum dirumuskan

## 1A. Fitur & Alur Pemakaian
### Alur harian — [FIX]
1. Buka app → layar **Hari Ini**: todo terjadwal/jatuh tempo hari ini + kejadian todo berulang, diurut auto-prioritas. Yang belum selesai dari hari sebelumnya ada di tab **Terlewat**
2. Mulai kerja → tap todo → ▶ timer (TimeLog ke workspace todo tsb)
3. Kapan saja → **catat cepat** (+): satu kalimat, pilih jenis (todo / catatan / blocker / keputusan)
4. Centang todo selesai → otomatis tercatat di jurnal hari ini
5. Akhir hari → jurnal hari ini sudah terisi sendiri dari langkah 2–4, tinggal tambah catatan/mood
6. Mingguan/bulanan → laporan per workspace (jam kerja, todo selesai, % kepatuhan)
- [FIX] Jurnal terisi otomatis dari aktivitas (rekap todo + timer) — lihat "Jurnal harian"

### Pembagian versi — [FIX]
- MVP: workspace, todo (prioritas, deadline, sub-task), **todo berulang + % kepatuhan**, catat cepat, timer, jurnal harian otomatis, laporan per workspace, reminder
- Versi berikutnya: tag, AI, export PDF/CSV, pencarian lanjutan

### Jurnal harian — [FIX]
- Halaman jurnal per hari = dua bagian:
  - **A. Rekap otomatis** (tidak diisi, dihitung dari data, tidak disimpan di DB): todo selesai (termasuk terlambat), todo terlewat/dilewati, total jam per workspace (dari TimeLog), skor kepatuhan hari itu
  - **B. Catatan** (diisi user): catatan bebas dengan label opsional (catatan / blocker / keputusan) + workspace opsional, dan mood hari itu
- Catatan bisa diubah jadi todo (todo menyimpan asal catatannya)

### Timer berjalan — [FIX]
- Di dalam app: bar mini timer yang selalu tampil di atas layar selama timer jalan (nama todo + waktu berjalan + tombol stop)
- Di luar app (Android): notifikasi ongoing dengan chronometer — waktu dihitung sistem dari `started_at`, tidak perlu foreground service
- Android 16+: notifikasi yang sama diajukan sebagai **Live Update** → tampil sebagai chip timer di status bar & lock screen (perangkat lama otomatis jatuh ke notifikasi biasa)
- iOS Dynamic Island (Live Activities): hanya kalau nanti rilis iOS — di luar scope sekarang
- Mode Fokus (kunci app / blokir app lain): **DIBATALKAN** — ribet, berisiko ditolak Play Store, dan tidak terlalu penting

### Todo berulang — [FIX] masuk MVP
- Pola: harian / setiap N hari; mingguan di hari tertentu (mis. Senin & Kamis, Senin–Jumat); bulanan di tanggal tertentu; opsional "Senin pertama tiap bulan". Tanggal 31 di bulan pendek → hari terakhir bulan itu
- Berhenti: tidak pernah / sampai tanggal tertentu / setelah N kali
- Edit di template = berlaku untuk kejadian berikutnya; edit di satu kejadian = hanya kejadian itu
- Sub-task (checklist) di template ikut disalin ke setiap kejadian
- Kejadian dibuat saat app dibuka (bisa offline); ID = UUID v5 dari (template + tanggal) supaya tidak dobel antar device

### Kepatuhan & daftar Terlewat — [FIX]
- Semua todo yang terlewat masuk daftar **Terlewat** (layar Hari Ini tetap bersih, tidak ada yang hilang diam-diam)
- **% kepatuhan** = selesai ÷ dijadwalkan, rentang **30 hari terakhir bergulir** (bisa pulih kalau rajin lagi). Ditampilkan per todo berulang, per workspace, dan per user
- Todo di Terlewat masih bisa diselesaikan → dihitung **selesai terlambat**, poin balik sebagian (mis. 50%)
- Tombol **Lewati** (sakit/libur/klien batal) → keluar dari perhitungan, tetap tercatat di histori
- Berlaku untuk **semua todo yang punya tanggal** (tanggal rencana atau deadline): todo biasa maupun berulang. **Tidak ada carry-over** — yang lewat masuk Terlewat
- Todo tanpa tanggal & tanpa deadline masuk **Inbox**, tidak memengaruhi skor
- Mengubah tanggal sebelum harinya lewat = merencanakan ulang, tanpa penalti
- Tab Terlewat punya aksi **Selesaikan (terlambat)**, **Lewati**, dan **Jadwalkan ulang** — menjadwalkan ulang tetap mencatat terlewatnya (tabel `todo_misses`), jadi tidak ada celah untuk menghindari penalti
- Skor dihitung saat ditampilkan, tidak disimpan di DB

## 1B. Strategi Pengerjaan — [FIX]
- **Web-first**: Backend Go → Web (SvelteKit) → baru Android (Flutter)
- Pekerjaan dipecah per fitur per agent (BE / FE Web / Mobile) di `docs/agents/` — mulai dari `docs/agents/README.md`
- Aturan bisnis bersama (formula prioritas, kepatuhan, recurring, dll) di `docs/spec/rules.md`; konvensi API di `docs/spec/api-conventions.md`; kontrak endpoint di `docs/api/openapi.yaml`
- Struktur repo: `backend/` (Go), `web/` (SvelteKit), `mobile/` (Flutter), `docs/`
- [FIX] Penyempurnaan yang sudah dikonfirmasi: status `missed` diturunkan (tidak disimpan), kejadian recurring di-backfill maks 30 hari, reminder Android dijadwalkan lokal (tanpa FCM di MVP)

## 2. Domain Model
### Konvensi umum
- [FIX] ID pakai UUID v7 yang boleh di-generate di client (offline-first), bukan auto-increment
- [FIX] Semua entity yang di-sync punya `created_at`, `updated_at`, `deleted_at` (soft delete)
- [FIX] `User` dan `RefreshToken` server-only, tidak ikut sync
- [FIX] Unique index pakai kondisi `WHERE deleted_at IS NULL` (email, nama tag per user, dll)
- [FIX] Soft delete tidak memicu `ON DELETE CASCADE` → aturan cascade ditangani di application layer
- [FIX] `User` menyimpan `timezone`; semua tanggal "hari" (`entry_date`, `scheduled_date`) disimpan sebagai tanggal lokal user

### Entity
- [FIX] `User` — email, password (nullable), google_id (nullable), name, timezone
- [FIX] `Workspace` — satu level saja (nama, warna, ikon), representasi klien/project. **Tidak bisa dihapus, hanya di-archive** (`archived_at`) supaya histori waktu kerja tetap utuh
- [FIX] `JournalEntry` — 1 entry = 1 hari per user (halaman harian): `entry_date`, `mood`. **Tanpa** `workspace_id` dan `hours_spent`
- [FIX] `JournalNote` (pengganti EntryItem) — catatan user di jurnal: isi, label opsional (note/blocker/decision), `workspace_id` opsional. Tipe "task" dihapus karena sudah diwakili Todo
  - Rekap otomatis (todo selesai/terlewat, jam kerja) tidak disimpan, dihitung dari Todo + TimeLog
  - Filter jurnal per workspace = catatan + todo + TimeLog dengan workspace tsb, dikelompokkan per tanggal
- [FIX] `TimeLog` — entity terpisah: workspace_id, todo_id (opsional), started_at/ended_at atau durasi manual, catatan. Satu-satunya sumber data durasi kerja
- [FIX] `Todo` — judul, catatan, workspace_id, status disimpan (pending/in_progress/done/cancelled/skipped) — `missed` diturunkan, tidak disimpan, `completed_at` (selesai terlambat = completed_at melewati jadwal/deadline, diturunkan, tidak disimpan), prioritas manual, `due_at`, `scheduled_date`, estimasi waktu, `parent_id` (sub-task **maks 1 level**), `source` (manual/journal/ai), `source_note_id`, `recurrence_rule` (RRULE, hanya di todo induk), `recurrence_parent_id` (todo hasil recurring)
- [FIX] `Reminder` — entity sendiri (satu todo bisa banyak reminder). Di todo induk recurring disimpan sebagai offset (mis. "30 menit sebelum due") lalu disalin ke tiap todo hasil recurring
- [FIX — versi berikutnya] `Tag` — per user, dipasang di catatan jurnal dan di todo: pivot `journal_note_tags` & `todo_tags`
- [FIX] `TodoMiss` — riwayat terlewat saat todo terlewat dijadwalkan ulang (todo_id, commitment_date)
- [FIX] AI tidak punya tabel khusus dulu, cukup `source = ai` di Todo

### Aturan perilaku
- [FIX] **Tanpa carry-over**: semua todo bertanggal (biasa, berulang, ber-deadline) yang lewat dan belum selesai → terlewat, tampil di tab Terlewat (lihat Bagian 1A)
- [FIX] **Auto-prioritas**: skor dihitung saat ditampilkan (prioritas manual, deadline, rutin), tidak disimpan di DB
- [FIX] **% kepatuhan**: dihitung saat ditampilkan dari histori status todo, tidak disimpan di DB
- [FIX] **Jurnal ↔ Todo**: todo selesai/terlewat otomatis muncul di rekap jurnal (dihitung, bukan disalin); todo yang dibuat dari catatan menyimpan `source_note_id`

## 3. Auth — DITUNDA (fokus ke fitur dulu)
- [FIX] Dual mode: Google OAuth + email/password sendiri
- [FIX] JWT access token (15 menit) + refresh token per device
- [DITUNDA] Detail auth dibahas lagi setelah core fitur jelas. Catatan yang sudah ketemu, jangan sampai lupa:
  - Verifikasi email untuk akun password: wajib sebelum akun boleh digabung dengan login Google (celah pengambilalihan akun)
  - Lupa password: OTP via email → butuh layanan kirim email
  - Hapus akun: wajib dari Google Play (di dalam app + halaman web)
  - Logout semua device, rate limit login/OTP, rotasi refresh token
  - Timezone user diambil dari HP saat login (dibutuhkan Bagian 2)

## 4. Backend Stack
- [FIX] Go + **chi** (router standar net/http) + **sqlc + pgx** (query SQL eksplisit, kode type-safe digenerate) + PostgreSQL + golang-migrate
- [FIX] Kontrak API **spec-first**: `openapi.yaml` ditulis dulu → **oapi-codegen** membuat kerangka server; kode yang tidak sesuai kontrak gagal compile. FE bisa mulai dengan mock dari kontrak
- [FIX] ID baru pakai **UUID v7** (berurutan menurut waktu); UUID v5 untuk kejadian rutin
- [FIX] Struktur folder DDD-style (domain / application / infrastructure / interfaces)
- [FIX] Development di **Windows**, server/CI Linux → task runner **Task** (`Taskfile.yml`, lintas OS) menggantikan Makefile; alat bantu (migrate, seed) ditulis sebagai program Go; line ending LF lewat `.gitattributes`
- [FIX] Scaffold awal di `backend/` (Fiber + GORM) **ditulis ulang** di BE M0 mengikuti stack & domain model baru
- [FIX] Daftar endpoint per fitur sudah dirinci di `docs/agents/be.md` (tags menyusul di versi berikutnya)
- [FIX] Scheduler reminder = worker goroutine di proses API (ticker 30 detik + `FOR UPDATE SKIP LOCKED`). Recurring tidak butuh scheduler (kejadian dibuat saat dibaca)
- [OPEN — versi berikutnya] Kalau fitur AI dipakai: integrasi ke LLM API mana, dan di sisi mana (backend proxy ke API, supaya API key gak kebawa ke client)
- [FIX] Validasi input (validator OpenAPI dari spec) & rate limit (auth 10/menit per IP, lainnya 120/menit per user) — BE M0

## 5. Infra & Hosting
- [FIX] VPS murah (bukan shared hosting — karena limitasi queue/cron)
- [FIX] Sementara **staging di lokal (Docker Desktop)**; VPS baru dipilih menjelang rilis web (BE M11)
- [FIX] VPS dipakai juga buat aplikasi lain → Docker Compose + Traefik (reverse proxy + SSL otomatis per subdomain)
- [OPEN] Provider VPS belum dipilih (DigitalOcean/Vultr/lokal seperti Biznet Gio/IDCloudHost)
- [OPEN] Spek VPS (RAM/CPU) belum ditentukan — perlu dipertimbangkan ulang karena ada job scheduler + kemungkinan panggilan ke AI API
- [OPEN] Subdomain final belum fix (placeholder saat ini: `journal.genossys.id`)

## 5A. Web (SvelteKit) — [FIX]
- SvelteKit + Svelte 5 + TypeScript, SPA statis (`adapter-static`), Tailwind CSS
- Tipe API di-generate dari `openapi.yaml`; access token di memori, refresh token via cookie HttpOnly
- Reminder di web lewat Web Push (service worker)
- Detail task: `docs/agents/fe.md`

## 6. Mobile (Flutter) — fase setelah web
- [FIX] Dikerjakan setelah web MVP jalan & Sync API (BE-12) siap. Detail task: `docs/agents/mobile.md`
- [FIX] Flutter dengan backend API (bukan backend-less)
- [FIX] Perlu local SQLite cache untuk offline quick-capture + sync ke server
- [FIX] Reminder todo jadi fitur inti — di Android dijadwalkan **lokal** di HP (jalan saat offline), **tanpa FCM** di MVP. Web pakai Web Push
- [FIX] Notifikasi timer berjalan (chronometer) + Live Update di Android 16+ — lihat Bagian 1A "Timer berjalan"
- [OPEN] State management belum dipilih (Riverpod / Bloc / lainnya)
- [FIX] Struktur folder Flutter per fitur (`features/<nama>/data|domain|presentation`) — MB-00
- [FIX] Sync: outbox + pull/push, last-write-wins berdasarkan `client_updated_at` (dijepit ke waktu server) — BE-12 & MB-02
- [OPEN — versi berikutnya] Override urutan auto-prioritas oleh user (MVP: urutan murni dari skor)

## 7. Rilis ke Play Store
- [OPEN] Nama app
- [OPEN] App icon & branding
- [OPEN] Privacy policy page (wajib buat submission Play Store, makin penting karena ada fitur AI opsional yang mungkin kirim data ke 3rd party)
- [OPEN] Play Console developer account & setup

## 8. Belum Dibahas Sama Sekali
- Search/filter lanjutan (misal cari lintas tag/workspace)
- Export data (PDF/CSV)
- Batas pakai gratis (kalau nanti mau freemium) — sekarang masih full gratis tanpa limit

---
*Dokumen ini working draft, diupdate seiring progress breakdown per poin.*
