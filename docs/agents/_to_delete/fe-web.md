# Task Agent FE (Web — SvelteKit)

Stack: SvelteKit + Svelte 5 (runes), TypeScript strict, `adapter-static` (SPA, `ssr = false`), Tailwind CSS, `openapi-typescript` + `openapi-fetch` (tipe dari `docs/api/openapi.yaml`), TanStack Query untuk Svelte (cache & invalidasi), Vitest + Playwright.
Baca dulu: `README.md` di folder ini, `../spec/api-conventions.md`, dan `../spec/rules.md` (untuk memahami arti data — **FE tidak menghitung ulang** skor/kepatuhan/terlewat, semuanya dari API).

Prinsip UI:
- Bahasa Indonesia, siap i18n (semua teks lewat satu modul pesan).
- Responsif: dipakai di desktop, tapi tetap nyaman di browser HP.
- Setiap layar punya state: loading (skeleton), kosong (empty state yang mengajak aksi), error (pesan + coba lagi).
- Optimistic update untuk aksi cepat (centang todo, start/stop timer), rollback kalau gagal.
- Shortcut keyboard untuk aksi utama.

---

## FE-00 — Fondasi
Butuh: GATE G0 di `be.md` (openapi.yaml awal + generator)

- **FE-00.1** Proyek SvelteKit di `web/`, TypeScript strict, ESLint + Prettier, Tailwind, alias path.
- **FE-00.2** Script `npm run api:types` → generate tipe dari `../docs/api/openapi.yaml`. Siapkan mock server dari kontrak (MSW + handler per endpoint, atau Prism) supaya fitur bisa dikerjakan setelah GATE-K, sebelum endpoint BE selesai. Client `openapi-fetch` dengan base URL dari env (`PUBLIC_API_URL`).
- **FE-00.3** Lapisan API: interceptor `Authorization`, auto-refresh saat 401 (satu refresh bersamaan, request lain menunggu), pemetaan error standar → pesan UI, `credentials: 'include'` untuk endpoint auth.
- **FE-00.4** Layout aplikasi: sidebar (Hari Ini, Terlewat, Inbox, Semua Todo, Rutin, Jurnal, Laporan, Workspace, Pengaturan) + header (tempat mini-bar timer & tombol catat cepat). Di layar kecil sidebar jadi bottom-nav/drawer.
- **FE-00.5** Komponen dasar: Button, Input, Select, DatePicker, Modal, Drawer, Toast, Badge, Skeleton, EmptyState, ConfirmDialog, ColorPicker, IconPicker.
- **FE-00.6** Util tanggal: format tanggal lokal Indonesia, tampilkan timestamp dalam zona waktu user (dari `/me`), bukan zona browser.
- **FE-00.7** Testing: Vitest siap, Playwright dengan mock API (MSW) + 1 test e2e contoh. CI: lint, typecheck, test, build.
- **FE-00.8** Build produksi: Dockerfile static (Caddy/nginx) dengan fallback SPA ke `index.html`, siap dipasang di compose BE-13 (`app.<domain>`).
- AC: `npm run build` menghasilkan SPA statis; halaman kosong ber-layout tampil dengan sidebar.

---

## FE-01 — Auth
Butuh: BE-01, FE-00

- Halaman `/login` dan `/register` (email + password), tombol **Masuk dengan Google** (Google Identity Services, client ID web dari env → kirim `id_token` ke `POST /auth/google`).
- Semua request auth mengirim header `X-Client: web` dan `timezone` dari `Intl.DateTimeFormat().resolvedOptions().timeZone`.
- Access token disimpan **di memori saja** (bukan localStorage). Saat app dibuka → coba `POST /auth/refresh` (cookie) untuk memulihkan sesi.
- Route guard: halaman app butuh login; `/login` mengarah ke `/today` kalau sudah login.
- Pengaturan: ubah nama, lihat zona waktu (otomatis `PATCH /me` kalau zona browser berbeda — dengan konfirmasi), Logout, Logout semua perangkat.
- AC (Playwright): daftar → masuk ke Hari Ini; refresh browser tetap login; logout → kembali ke /login; 401 di tengah sesi memicu refresh sekali lalu request diulang.

---

## FE-02 — Workspace
Butuh: BE-02, FE-01

- Halaman `/workspaces`: daftar (warna + ikon + nama + kepatuhan setelah BE-06), buat/edit lewat modal, arsip/batal arsip, urutkan dengan drag & drop.
- Komponen `WorkspacePicker` (dipakai di form todo, timer, catatan) — hanya workspace aktif, ada opsi "buat baru" inline.
- Tampilan workspace terarsip: dipisah, bisa dibuka untuk melihat histori.
- AC: workspace terarsip tidak muncul di picker; error `WORKSPACE_ARCHIVED` ditampilkan jelas.

---

## FE-03 — Todo dasar + sub-task
Butuh: BE-03, FE-02

- Komponen `TodoItem` (checkbox, judul, badge prioritas, workspace, due, ikon sub-task x/y) — dipakai di semua daftar.
- Drawer detail todo (`/todos/[id]` atau panel samping): judul, catatan (markdown ringan), workspace, prioritas, `scheduled_date`, `due_at` (tanggal + jam opsional), estimasi, checklist sub-task (tambah/centang/urut/hapus), aksi Mulai/Selesai/Lewati/Batal/Buka lagi/Hapus.
- Halaman `/inbox` (belum direncanakan) dan `/todos` (semua, dengan filter workspace/status/rentang tanggal).
- Form cepat tambah todo di atas daftar (judul + Enter).
- AC: menyelesaikan induk ikut mencentang sub-task di UI (sesuai respons API); transisi status yang ditolak API menampilkan pesan.

---

## FE-04 — Layar Hari Ini
Butuh: BE-04, FE-03

- Halaman `/today` (halaman awal setelah login): urutan **persis dari API** (jangan diurutkan ulang di client).
- Penanda: "jatuh tempo hari ini", ikon rutin.
- Banner kecil berisi jumlah Terlewat (link ke /missed) dan Inbox.
- Header: tanggal hari ini + ringkasan (x dari y selesai).
- Todo yang dicentang pindah ke bagian "Selesai hari ini" (collapsible) di bawah.
- Shortcut: `j/k` navigasi, `x` selesai, `Enter` buka detail.
- AC: data di-refresh saat tab kembali fokus dan saat tanggal berganti (lewat tengah malam zona user).

---

## FE-05 — Rutin (todo berulang)
Butuh: BE-05, FE-03

- Komponen `RecurrenceEditor` yang menghasilkan RRULE subset (`rules.md` §6):
  - Preset: Setiap hari · Hari kerja (Sen–Jum) · Setiap minggu (pilih hari) · Setiap bulan tanggal X · Setiap bulan "Senin pertama/terakhir" · Kustom (setiap N hari/minggu/bulan)
  - Berakhir: tidak pernah / pada tanggal / setelah N kali
  - Pratinjau 5 tanggal berikutnya (dari respons API setelah simpan, atau hitung lokal hanya untuk pratinjau sebelum simpan — tandai "perkiraan")
  - Info khusus saat memilih tanggal 29–31: "Di bulan yang lebih pendek, jatuh di hari terakhir bulan itu"
- Halaman `/routines`: daftar template (ringkasan pola, tanggal berikutnya, % kepatuhan), buat/edit template (form todo + RecurrenceEditor + checklist), hentikan rutin.
- Saat mengedit template tampilkan catatan: "Perubahan berlaku untuk jadwal berikutnya".
- AC: setiap preset menghasilkan RRULE yang diterima API (test unit per preset).

---

## FE-06 — Terlewat + kepatuhan
Butuh: BE-06, FE-05

- Halaman `/missed`: daftar terlewat (todo biasa & rutin) dikelompokkan per tanggal komitmen, aksi **Selesaikan (terlambat)**, **Lewati**, dan **Jadwalkan ulang** (pilih tanggal baru; beri tahu user bahwa terlewatnya tetap tercatat) (dengan alasan opsional: sakit/libur/dibatalkan/lainnya — disimpan di notes kejadian).
- Komponen `ComplianceBadge` (persen + warna: ≥80 hijau, 50–79 kuning, <50 merah, null = "belum ada data") dipakai di Rutin, Workspace, Laporan, dan header/profil.
- Penjelasan singkat cara hitung (tooltip/modal) sesuai `rules.md` §7.
- AC: setelah menyelesaikan item terlewat, badge kepatuhan terkait ikut ter-update (invalidasi query).

---

## FE-07 — Timer + TimeLog
Butuh: BE-07, FE-03

- Tombol ▶/■ di `TodoItem` dan drawer detail.
- **Mini-bar timer** di header, tampil di semua halaman selama ada timer berjalan: nama todo/workspace, waktu berjalan (dihitung lokal dari `started_at`, sinkron ulang dari `/timer/current` saat fokus), tombol stop, klik → buka todo.
- `document.title` menampilkan waktu berjalan (mis. `⏱ 00:25:13 · Nama todo`).
- Mulai timer tanpa todo: pilih workspace + catatan.
- Daftar TimeLog di detail todo + form input manual (durasi) + edit/hapus.
- Peringatan untuk TimeLog `suspicious` (>16 jam): "Sepertinya timer lupa dihentikan — perbaiki?"
- AC: dua tab terbuka → keduanya menampilkan timer yang sama setelah fokus; start di todo lain menghentikan timer lama di UI.

---

## FE-08 — Catat cepat
Butuh: FE-03, GATE G8a (catatan jurnal)

- Tombol `+` global di header + shortcut `N` (dan `Ctrl/Cmd+K`).
- Modal satu baris: teks + pilihan jenis (Todo · Catatan · Blocker · Keputusan) + workspace opsional. `Enter` simpan, `Tab` pindah jenis, `Esc` tutup.
  - Todo → `POST /todos` (default `scheduled_date` = hari ini; opsi "Inbox")
  - Catatan/Blocker/Keputusan → `POST /journal/{hari ini}/notes` dengan label
- Target: dari tekan `N` sampai tersimpan < 3 detik, tanpa mouse.
- AC (Playwright): alur keyboard-only untuk keempat jenis.

---

## FE-09 — Jurnal harian
Butuh: GATE G8a untuk bagian catatan & mood (bisa dikerjakan lebih dulu); GATE G8b untuk rekap otomatis; FE-07

- Halaman `/journal/[date]` (default hari ini) dengan navigasi hari sebelumnya/berikutnya + pemilih tanggal.
- **Bagian A — Rekap otomatis** (read-only): todo selesai (tandai terlambat), terlewat/dilewati, jam kerja per workspace (bar kecil), kepatuhan hari itu.
- **Bagian B — Catatan**: daftar catatan dengan label (warna per label) & workspace, tambah/edit/hapus, aksi "Jadikan todo".
- Pemilih mood (5 pilihan emoji) di atas halaman.
- Halaman `/journal` (histori): kalender bulanan/daftar hari dengan indikator (mood, menit kerja, jumlah selesai) + filter workspace.
- AC: hari tanpa data menampilkan empty state yang mengajak menulis catatan, bukan error.

---

## FE-10 — Laporan per workspace
Butuh: BE-10, FE-06

- Halaman `/reports`: pilihan rentang (Minggu ini, Bulan ini, Bulan lalu, Kustom), tabel per workspace (jam, selesai tepat/terlambat, terlewat, kepatuhan).
- Detail `/reports/[workspaceId]`: grafik jam per hari/minggu (LayerChart atau Chart.js), daftar todo selesai, total jam besar di atas.
- Tombol salin ringkasan teks (untuk dikirim ke klien) — export PDF/CSV menyusul di versi berikutnya.
- AC: angka total sama dengan respons API (tidak dihitung ulang).

---

## FE-11 — Reminder (Web Push)
Butuh: GATE G10a (reminder todo biasa); GATE G10b untuk reminder di template rutin; FE-05

- Editor reminder di detail todo: waktu absolut atau preset relatif ("saat jatuh tempo", "30 menit sebelum", "1 jam sebelum", "1 hari sebelum"); di template rutin hanya relatif (offset).
- Service worker (`src/service-worker.ts`) untuk menerima push & menampilkan notifikasi; klik notifikasi membuka todo.
- Alur izin: jangan minta izin saat app dibuka — minta saat user pertama kali menambah reminder, dengan penjelasan dulu. Kalau ditolak, tampilkan cara mengaktifkan.
- Kelola langganan push per browser di Pengaturan.
- AC: reminder diterima saat tab tertutup (browser tetap berjalan); unsubscribe menghapus langganan di server.
