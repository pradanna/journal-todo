# Agent FE — Web (SvelteKit)

Brief lengkap untuk agent Frontend Web: konteks, aturan kerja, mode mock, urutan pengerjaan, dan detail tiap fitur.

## Peran & cara kerja
- Kamu mengerjakan **seluruh web app** di folder `web/`.
- **FE berjalan paralel dengan BE memakai data mock.** Kamu tidak menunggu endpoint BE selesai — cukup menunggu **kontrak** fitur tersebut tersedia di `docs/api/openapi.yaml` (**GATE-K** di `be.md`).
- Begitu BE menyelesaikan suatu fitur (**GATE penuh**), fitur itu **di-switch** dari mock ke API asli lewat konfigurasi, lalu diuji ulang.
- Keputusan produk ada di `plan-project.md` (root repo). Poin **[FIX]** tidak boleh diubah tanpa konfirmasi Pradana.
- Kamu **tidak boleh mengubah kontrak** (`docs/api/`). Kalau kontrak kurang/salah untuk kebutuhan UI → tulis permintaan di PR/issue untuk agent BE.

## Konteks produk (ringkas)
App jurnal kerja + todo pintar untuk siapa saja yang bekerja per-task/per-hari.
- **Workspace** = klien/project (diarsip, tidak dihapus).
- **Todo**: prioritas, deadline, tanggal rencana, sub-task. Todo bertanggal yang lewat & belum selesai → tab **Terlewat** (mengurangi % kepatuhan). Tanpa tanggal → **Inbox**.
- **Hari Ini**: todo hari ini, urutan dari skor prioritas (dihitung BE).
- **Rutin**: todo berulang. **% kepatuhan** 30 hari.
- **Timer**: mini-bar selalu tampil saat berjalan. **Catat cepat**: tombol `+` / shortcut.
- **Jurnal harian**: rekap otomatis + catatan + mood. **Laporan** per workspace. **Reminder** via Web Push.

## Wajib dibaca sebelum mulai
1. `plan-project.md` — bagian 1A (fitur & alur).
2. `docs/spec/api-conventions.md` — format respons/error, tanggal, auth web (cookie).
3. `docs/spec/rules.md` — untuk **memahami arti data**. FE **tidak menghitung ulang** skor prioritas, terlewat, kepatuhan, atau rekap di mode API asli — semua dari BE. (Pengecualian: logika tiruan sederhana di dalam lapisan mock, lihat "Mode mock".)
4. `docs/agents/be.md` — tabel status untuk mengetahui kontrak/fitur BE mana yang sudah siap.

## Stack

| Kebutuhan | Pilihan |
|---|---|
| Framework | SvelteKit + **Svelte 5** (runes), TypeScript strict |
| Build | `adapter-static` (SPA, `ssr = false`), hasil = file statis |
| Styling | Tailwind CSS |
| Tipe API | `openapi-typescript` dari `docs/api/openapi.yaml` |
| HTTP client | `openapi-fetch` (path & body bertipe dari kontrak) |
| Data fetching | TanStack Query untuk Svelte (cache, invalidasi, optimistic update) |
| Mock | **MSW** (Mock Service Worker) — handler bertipe dari kontrak + mock store di memori |
| Tanggal | `date-fns` + `date-fns-tz` (tampilkan dalam zona waktu user dari `/me`) |
| Grafik | LayerChart (atau Chart.js) |
| Test | Vitest (unit) + Playwright (e2e) |
| Package manager | pnpm (atau npm) — konsisten satu |

**Lingkungan development: Windows** (PC Pradana); CI Linux. Semua script `package.json` harus jalan di PowerShell & Linux: jangan pakai `rm -rf`, `cp`, `export VAR=…`, `&&` khusus bash — pakai paket lintas OS (`rimraf`, `cross-env`, `shx`) atau script Node kecil. Line ending LF (`.gitattributes` di root).

## Mode mock & switch ke API asli
Inti cara kerja FE. Kerjakan fondasinya di FE-00.

### Konfigurasi
```
PUBLIC_API_URL=http://localhost:8080/api/v1
PUBLIC_API_MODE=mock            # mock | real | hybrid
PUBLIC_MOCK_FEATURES=todos,routines,missed,journal   # dipakai saat hybrid: fitur yang MASIH mock
```
- `mock` — semua request dijawab MSW (BE tidak perlu menyala).
- `real` — tidak ada mock sama sekali.
- `hybrid` — fitur di `PUBLIC_MOCK_FEATURES` dijawab mock, sisanya diteruskan ke BE asli (`onUnhandledRequest: 'bypass'`). Dipakai selama transisi: tiap fitur BE yang lolos GATE penuh dihapus dari daftar.
- Banner kecil "MODE MOCK" / "HYBRID: todos, journal" tampil di pojok layar selama mode bukan `real`. Build produksi **wajib** `real` (cek di CI).

### Struktur
```
web/src/lib/api/
  schema.d.ts          HASIL GENERATE dari openapi.yaml — jangan diedit
  client.ts            openapi-fetch + interceptor auth/refresh/error
  queries/<fitur>.ts   fungsi TanStack Query per fitur (satu-satunya tempat komponen mengambil data)
web/src/mocks/
  browser.ts           setup MSW sesuai PUBLIC_API_MODE / PUBLIC_MOCK_FEATURES
  db.ts                mock store di memori (opsional disimpan ke localStorage agar data bertahan saat reload)
  seed.ts              data contoh realistis (beberapa workspace, todo hari ini, terlewat, rutin, jurnal)
  handlers/<fitur>.ts  handler MSW per fitur
```

### Aturan mock
- **Handler wajib bertipe dari `schema.d.ts`** (request & response). Kalau kontrak berubah, mock gagal compile → ketahuan sebelum switch.
- Mock harus **stateful**: create/update/complete benar-benar mengubah `db.ts`, supaya alur UI bisa dicoba end-to-end.
- Mock meniru perilaku penting secara sederhana: urutan Hari Ini, status terlewat, % kepatuhan kasar, rekap jurnal. Logika tiruan ini **hanya boleh ada di `src/mocks/`**, tidak pernah di komponen atau `lib/`. Angka mock tidak harus identik dengan BE.
- Mock juga mensimulasikan error sesuai kontrak (`422` validasi, `404`, `409`, `401` token kedaluwarsa) dan latensi kecil (200–500 ms) agar state loading/error teruji. Sediakan toggle dev untuk memaksa error.
- Komponen **tidak boleh tahu** sedang mock atau tidak — mereka hanya memanggil `lib/api/queries/*`.

### Prosedur switch per fitur (saat BE lolos GATE penuh)
1. Jalankan BE lokal (`task db:up`, `task run` di `backend/`) atau staging lokal (M4.5).
2. `pnpm api:types` — pastikan tipe terbaru.
3. Hapus fitur dari `PUBLIC_MOCK_FEATURES` (mode `hybrid`).
4. Jalankan Playwright fitur itu terhadap BE asli (`pnpm test:e2e:real -- --grep @<fitur>`).
5. Perbedaan perilaku → **jangan ditambal di FE**. Kalau BE tidak sesuai kontrak/`rules.md` → laporkan ke agent BE. Kalau mock yang keliru → perbaiki mock.
6. Perbarui kolom "Sumber data" di tabel status jadi **Asli**.
7. Setelah semua fitur asli → `PUBLIC_API_MODE=real`; mock tetap dipertahankan untuk test & development offline.

## Aturan kerja
- **Satu fitur/langkah = satu branch/PR**: `feat/FE-<id>-<slug>`. Commit: `<tipe>(<fitur>): <ringkasan>`.
- **Definition of Done** setiap fitur:
  - AC terpenuhi di mode mock
  - Unit test logika (Vitest) + Playwright alur utama yang **bisa dijalankan di mode mock maupun real** (tag `@<fitur>`)
  - `pnpm lint`, `pnpm check` (svelte-check/typecheck), `pnpm test`, `pnpm build` lolos
  - Tipe API up-to-date dengan `openapi.yaml`
- Teks UI bahasa Indonesia lewat satu modul pesan (siap i18n). Identifier kode bahasa Inggris.
- Setiap layar punya state **loading** (skeleton), **kosong** (empty state yang mengajak aksi), dan **error** (pesan + coba lagi).
- Optimistic update untuk aksi cepat (centang todo, start/stop timer, catatan), rollback kalau gagal.
- Waktu ditampilkan dalam **zona waktu user dari `/me`**, bukan zona browser.
- Responsif: desktop utama, tetap nyaman di browser HP. Aksesibilitas dasar: bisa dipakai dengan keyboard, label pada input, kontras cukup.
- Jangan menambah dependency besar tanpa alasan di PR.
- Kalau `fe.md`, `rules.md`, dan plan bertentangan → berhenti, tanya di PR.

## Pola baku per fitur
1. **Ambil kontrak** — pastikan GATE-K fitur ini lolos, `pnpm api:types`
2. **Mock** — `mocks/handlers/<fitur>.ts` + data di `seed.ts` + logika store di `db.ts`
3. **Query layer** — `lib/api/queries/<fitur>.ts` (query key, fetch, mutation, invalidasi)
4. **UI** — route + komponen
5. **Test** — Vitest + Playwright (`@<fitur>`) di mode mock
6. **Switch** (nanti, saat BE GATE penuh) — prosedur di atas

## Urutan & status

| # | ID | Fitur | Butuh kontrak (GATE-K BE) | Ukuran | Status | Sumber data |
|---|---|---|---|---|---|---|
| 1 | FE-00 | Fondasi + mode mock | G0 | L | ⬜ | — |
| 2 | FE-01 | Auth | M1 | M | ⬜ | Mock |
| 3 | FE-02 | Workspace | M2 | M | ⬜ | Mock |
| 4 | FE-03 | Todo + sub-task | M3 | L | ⬜ | Mock |
| 5 | FE-04 | Hari Ini | M4 | M | ⬜ | Mock |
| 6 | FE-07 | Timer + mini-bar | M5 | M | ⬜ | Mock |
| 7 | FE-08 | Catat cepat | M3, M8a | S | ⬜ | Mock |
| 8 | FE-05 | Rutin | M6 | L | ⬜ | Mock |
| 9 | FE-06 | Terlewat + kepatuhan | M7 | M | ⬜ | Mock |
| 10 | FE-09 | Jurnal harian | M8a, M8b | M | ⬜ | Mock |
| 11 | FE-10 | Laporan | M9 | M | ⬜ | Mock |
| 12 | FE-11 | Reminder (Web Push) | M10a, M10b | M | ⬜ | Mock* |
| 13 | FE-12 | Switch penuh ke API asli + rilis | semua GATE penuh | M | ⬜ | Asli |

Ukuran: **S** ≈ setengah hari, **M** ≈ 1–2 hari, **L** ≈ 3–4 hari (perkiraan kasar).
Status: ⬜ belum · 🟡 dikerjakan · ✅ selesai (mode mock) · 🔁 sudah di-switch ke asli. Tabel ini diperbarui Pradana.
\* Web Push tidak bisa disimulasikan penuh oleh mock; FE-11 hanya bisa diuji tuntas dengan BE asli.

Urutan ini mengikuti nilai bagi user, bukan urutan BE. Kalau kontrak suatu fitur belum tersedia, **ambil fitur berikutnya yang kontraknya sudah ada** — jangan mengarang bentuk data sendiri.

---

## FE-00 — Fondasi + mode mock
Butuh: GATE G0 (`openapi.yaml` awal + generator).

- [ ] **0.1 Proyek** — SvelteKit di `web/`, TypeScript strict, ESLint + Prettier, Tailwind, alias path, `adapter-static` + fallback `index.html`. Script lintas OS (Windows & Linux).
- [ ] **0.2 Tipe & client** — `pnpm api:types` → `src/lib/api/schema.d.ts`; `client.ts` dengan `openapi-fetch`, base URL dari env.
- [ ] **0.3 Lapisan auth di client** — header `Authorization`; satu proses refresh bersamaan saat `401` (request lain menunggu lalu diulang); `credentials: 'include'` untuk endpoint auth; header `X-Client: web`; pemetaan error standar → pesan UI.
- [ ] **0.4 Mode mock** — MSW + `PUBLIC_API_MODE` (`mock|real|hybrid`) + `PUBLIC_MOCK_FEATURES`, `db.ts` (+ opsi simpan ke localStorage), `seed.ts`, banner mode, toggle paksa error & latensi, cek CI bahwa build produksi `real`.
- [ ] **0.5 Layout** — sidebar (Hari Ini, Terlewat, Inbox, Semua Todo, Rutin, Jurnal, Laporan, Workspace, Pengaturan) + header (slot mini-bar timer & tombol catat cepat). Di layar kecil: drawer/bottom-nav.
- [ ] **0.6 Komponen dasar** — Button, Input, Textarea, Select, DatePicker, TimePicker, Modal, Drawer, Toast, Badge, Skeleton, EmptyState, ConfirmDialog, ColorPicker, IconPicker.
- [ ] **0.7 Util** — format tanggal Indonesia dalam zona waktu user, durasi (`1j 25m`), modul pesan (i18n), shortcut keyboard global.
- [ ] **0.8 Testing & CI** — Vitest, Playwright dengan dua profil (`mock` default, `real` terhadap BE lokal), GitHub Actions: lint, check, test, build.
- [ ] **0.9 Docker produksi** — image static (Caddy/nginx) dengan fallback SPA, siap dipasang di compose BE (M4.5/M11).
- AC: `pnpm dev` di Windows menampilkan layout berisi data seed dari mock tanpa BE menyala; `pnpm build` menghasilkan SPA statis.

## FE-01 — Auth
Butuh kontrak: M1.

- [ ] `/login` & `/register` (email + password), tombol **Masuk dengan Google** (Google Identity Services, client ID web dari env → `id_token` ke `POST /auth/google`). Di mode mock: Google langsung sukses dengan akun demo.
- [ ] Kirim `timezone` dari `Intl.DateTimeFormat().resolvedOptions().timeZone` saat register/login.
- [ ] Access token **hanya di memori**. Saat app dibuka → `POST /auth/refresh` (cookie) untuk memulihkan sesi.
- [ ] Route guard: halaman app butuh login; `/login` → `/today` kalau sudah login.
- [ ] Pengaturan: ubah nama, zona waktu (tawarkan `PATCH /me` kalau zona browser berbeda), Logout, Logout semua perangkat.
- AC (Playwright `@auth`): daftar → masuk Hari Ini; reload tetap login; logout → /login; `401` di tengah sesi memicu refresh sekali lalu request diulang.

## FE-02 — Workspace
Butuh kontrak: M2.

- [ ] `/workspaces`: daftar (warna, ikon, nama, kepatuhan bila sudah ada), buat/edit via modal, arsip/batal arsip, urutkan drag & drop.
- [ ] Komponen `WorkspacePicker` (dipakai form todo, timer, catatan): hanya workspace aktif + "buat baru" inline.
- [ ] Bagian workspace terarsip terpisah, bisa dibuka untuk melihat histori.
- AC (`@workspace`): workspace terarsip tidak muncul di picker; error `WORKSPACE_ARCHIVED` ditampilkan jelas.

## FE-03 — Todo + sub-task
Butuh kontrak: M3.

- [ ] Komponen `TodoItem`: checkbox, judul, badge prioritas, workspace, tanggal/deadline, progres sub-task (x/y), tombol ▶ (aktif setelah FE-07).
- [ ] Drawer detail todo: judul, catatan (markdown ringan), workspace, prioritas, tanggal rencana, deadline (tanggal + jam opsional), estimasi, checklist sub-task (tambah/centang/urut/hapus), aksi Mulai/Selesai/Lewati/Batal/Buka lagi/Hapus.
- [ ] Halaman `/inbox` (tanpa tanggal) dan `/todos` (semua, filter workspace/status/rentang).
- [ ] Input cepat tambah todo di atas daftar (judul + Enter).
- AC (`@todo`): menyelesaikan induk ikut mencentang sub-task sesuai respons; transisi status yang ditolak menampilkan pesan.

## FE-04 — Hari Ini
Butuh kontrak: M4.

- [ ] `/today` = halaman awal setelah login. Urutan **persis dari respons** (jangan diurutkan ulang).
- [ ] Penanda: "jatuh tempo hari ini", ikon rutin. Banner jumlah **Terlewat** (link `/missed`) dan **Inbox**.
- [ ] Header: tanggal hari ini + "x dari y selesai". Todo yang dicentang pindah ke bagian "Selesai hari ini" (bisa dilipat).
- [ ] Shortcut: `j/k` navigasi, `x` selesai, `Enter` buka detail.
- AC (`@today`): data di-refresh saat tab kembali fokus dan saat tanggal berganti (tengah malam zona user).

## FE-07 — Timer + mini-bar
Butuh kontrak: M5.

- [ ] Tombol ▶/■ di `TodoItem` & detail. Mulai timer tanpa todo: pilih workspace + catatan.
- [ ] **Mini-bar** di header, tampil di semua halaman selama timer jalan: nama todo/workspace, waktu berjalan (dihitung lokal dari `started_at`, disinkronkan ulang dari `/timer/current` saat fokus), tombol stop, klik → buka todo.
- [ ] `document.title` menampilkan waktu berjalan (`⏱ 00:25:13 · Nama todo`).
- [ ] Daftar TimeLog di detail todo + input manual (durasi) + edit/hapus; peringatan untuk TimeLog `suspicious` (>16 jam).
- AC (`@timer`): dua tab menampilkan timer yang sama setelah fokus; start di todo lain menghentikan timer lama di UI.

## FE-08 — Catat cepat
Butuh kontrak: M3 (todo) dan M8a (catatan jurnal).

- [ ] Tombol `+` di header + shortcut `N` dan `Ctrl/Cmd+K`.
- [ ] Modal satu baris: teks + jenis (Todo · Catatan · Blocker · Keputusan) + workspace opsional. `Enter` simpan, `Tab` ganti jenis, `Esc` tutup.
  - Todo → `POST /todos` (default tanggal = hari ini; opsi "Inbox")
  - Catatan/Blocker/Keputusan → `POST /journal/{hari ini}/notes` dengan label
- AC (`@capture`): keempat jenis bisa disimpan hanya dengan keyboard, < 3 detik dari tekan `N`.

## FE-05 — Rutin
Butuh kontrak: M6.

- [ ] Komponen `RecurrenceEditor` → RRULE subset (`rules.md` §6):
  - Preset: Setiap hari · Hari kerja (Sen–Jum) · Setiap minggu (pilih hari) · Setiap bulan tanggal X · Setiap bulan "Senin pertama/terakhir" · Kustom (setiap N hari/minggu/bulan)
  - Berakhir: tidak pernah / pada tanggal / setelah N kali
  - Info untuk tanggal 29–31: "Di bulan yang lebih pendek, jatuh di hari terakhir bulan itu"
  - Pratinjau tanggal berikutnya: dari respons (`next_dates`); sebelum disimpan boleh perkiraan lokal berlabel "perkiraan"
- [ ] `/routines`: daftar template (ringkasan pola, tanggal berikutnya, % kepatuhan), buat/edit (form todo + editor + checklist), hentikan rutin. Catatan saat edit: "Perubahan berlaku untuk jadwal berikutnya".
- AC (`@routine`): setiap preset menghasilkan RRULE yang diterima kontrak (unit test per preset).

## FE-06 — Terlewat + kepatuhan
Butuh kontrak: M7.

- [ ] `/missed`: daftar terlewat (todo biasa & rutin) dikelompokkan per tanggal komitmen. Aksi **Selesaikan (terlambat)**, **Lewati** (alasan opsional: sakit/libur/dibatalkan/lainnya), **Jadwalkan ulang** (pilih tanggal; beri tahu "terlewatnya tetap tercatat").
- [ ] `ComplianceBadge` (≥80 hijau, 50–79 kuning, <50 merah, `null` = "belum ada data") dipakai di Rutin, Workspace, Laporan, header/profil + penjelasan cara hitung.
- AC (`@missed`): setelah aksi di item terlewat, badge kepatuhan terkait ikut ter-update (invalidasi query).

## FE-09 — Jurnal harian
Butuh kontrak: M8a (catatan & mood — bisa dikerjakan lebih dulu) dan M8b (rekap).

- [ ] `/journal/[date]` (default hari ini) + navigasi hari & pemilih tanggal.
- [ ] **Rekap otomatis** (read-only): todo selesai (tandai terlambat), terlewat/dilewati/dijadwalkan ulang, jam per workspace (bar kecil), kepatuhan hari itu.
- [ ] **Catatan**: daftar dengan label (warna per label) & workspace, tambah/edit/hapus, aksi "Jadikan todo". Pemilih mood (5 emoji).
- [ ] `/journal` histori: kalender bulanan/daftar hari dengan indikator (mood, menit kerja, jumlah selesai) + filter workspace.
- AC (`@journal`): hari tanpa data → empty state yang mengajak menulis catatan, bukan error.

## FE-10 — Laporan
Butuh kontrak: M9.

- [ ] `/reports`: rentang (Minggu ini, Bulan ini, Bulan lalu, Kustom), tabel per workspace (jam, selesai tepat/terlambat, terlewat, kepatuhan).
- [ ] `/reports/[workspaceId]`: grafik jam per hari/minggu, daftar todo selesai, total jam besar.
- [ ] Tombol **salin ringkasan teks** (untuk dikirim ke klien). Export PDF/CSV menyusul di versi berikutnya.
- AC (`@report`): angka sama persis dengan respons (tidak dihitung ulang).

## FE-11 — Reminder (Web Push)
Butuh kontrak: M10a (reminder todo) dan M10b (reminder di template rutin).

- [ ] Editor reminder di detail todo: waktu absolut atau preset relatif ("saat jatuh tempo", "30 menit sebelum", "1 jam sebelum", "1 hari sebelum"); di template rutin hanya relatif.
- [ ] Service worker untuk menerima push & menampilkan notifikasi; klik → buka todo. (Service worker MSW dan service worker push harus bisa hidup berdampingan — pisahkan scope/file.)
- [ ] Alur izin: minta izin saat user pertama menambah reminder, dengan penjelasan dulu; kalau ditolak, tampilkan cara mengaktifkan.
- [ ] Kelola langganan push per browser di Pengaturan.
- AC (`@reminder`, mode real): reminder diterima saat tab tertutup (browser tetap jalan); unsubscribe menghapus langganan di server.

## FE-12 — Switch penuh ke API asli + rilis
Butuh: semua GATE penuh BE (M1–M10b) dan staging lokal (M4.5).

- [ ] Jalankan prosedur switch untuk fitur yang tersisa; `PUBLIC_API_MODE=real`.
- [ ] Seluruh Playwright dijalankan di profil `real` terhadap staging lokal → semua hijau.
- [ ] Uji manual alur harian lengkap (plan 1A) bersama Pradana.
- [ ] Image produksi web dipasang di compose produksi (BE M11).
- AC: build produksi tanpa MSW di bundle; tidak ada banner mode.
