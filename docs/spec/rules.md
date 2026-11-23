# Aturan Bisnis Bersama (rules.md)

Satu-satunya sumber formula & aturan perilaku. BE (Go) dan MB (Dart, untuk mode offline) mengimplementasikan aturan yang sama dan wajib lolos `test-vectors/*.json` yang sama. FE web tidak menghitung ulang — FE menampilkan hasil dari API.

Semua aturan di dokumen ini sudah dikonfirmasi Pradana (23 Nov 2026).

---

## 1. Hari & zona waktu
- `users.timezone` (IANA, contoh `Asia/Jakarta`). Default saat daftar: zona waktu yang dikirim client; fallback `Asia/Jakarta`.
- **"Hari ini"** = tanggal lokal `now()` di zona waktu user. Semua kolom "hari" (`scheduled_date`, `entry_date`) bertipe `DATE` tanpa zona.
- `due_at`, `completed_at`, `started_at`, `ended_at`, `remind_at` bertipe timestamp UTC.
- "Tanggal lokal" sebuah timestamp = tanggal timestamp itu dikonversi ke `users.timezone`.
- Client memperbarui timezone lewat `PATCH /me` kalau zona waktu perangkat berubah.

## 2. Status todo
- **Disimpan**: `pending`, `in_progress`, `done`, `cancelled`, `skipped`
- **Diturunkan (tidak disimpan)**: `missed`. Todo dianggap *terlewat* bila statusnya `pending`/`in_progress`, bukan template, bukan sub-task, **dan** salah satu:
  - `scheduled_date < hari ini` (berlaku untuk todo biasa **dan** kejadian recurring), atau
  - `due_at` tidak null dan `due_at < now`
  > Alasan: tidak perlu job yang menulis status, tidak ada konflik antar device.
- **Tanggal komitmen** sebuah todo = `scheduled_date` kalau ada; kalau tidak ada, tanggal lokal `due_at`. Todo tanpa keduanya bukan komitmen (masuk Inbox).
- **Selesai terlambat** (diturunkan): status `done` dan
  - tanggal lokal `completed_at` > `scheduled_date` (kalau ada), atau
  - `completed_at > due_at` (kalau ada)
- Transisi yang valid:
  - `pending ⇄ in_progress`
  - `pending|in_progress → done | cancelled | skipped`
  - `done|cancelled|skipped → pending` (buka lagi; `completed_at` di-null-kan)
- `completed_at` diisi saat masuk `done`, di-null-kan saat keluar dari `done`.
- Menyelesaikan todo induk → semua sub-task yang masih terbuka ikut `done`.
- Template recurring (`recurrence_rule` tidak null) **bukan** todo yang bisa dikerjakan: tidak punya status aktif, tidak muncul di Hari Ini/Terlewat.

## 3. Daftar yang ditampilkan
- **Hari Ini** = todo terbuka (pending/in_progress), bukan template, bukan sub-task, tidak terlewat, dan salah satu:
  - `scheduled_date = hari ini`, atau
  - tanggal lokal `due_at` = hari ini
- **Terlewat** = semua todo yang memenuhi definisi *terlewat* di §2
- **Inbox** = todo terbuka tanpa `scheduled_date` dan tanpa `due_at` (belum direncanakan). Tidak muncul di Hari Ini.
- **Rutin** = daftar template recurring milik user

## 4. Carry-over — DIHAPUS
- **[FIX]** Tidak ada carry-over. Todo yang tanggalnya lewat dan belum selesai = terlewat (§2), muncul di tab Terlewat, dan mengurangi kepatuhan (§7).
- Mengubah `scheduled_date` sebelum harinya lewat = merencanakan ulang, tidak ada penalti.
- Menjadwalkan ulang todo yang **sudah terlewat** tetap mencatat terlewatnya — lihat §7 "Riwayat terlewat".

## 5. Skor auto-prioritas
Dihitung saat ditampilkan, tidak disimpan. Semakin besar = semakin atas.

```
score = P + D + R

P (prioritas manual):  1 (urgent)=40, 2 (tinggi)=25, 3 (normal, default)=10, 4 (rendah)=0
D (deadline):          due hari ini=40, besok=25, ≤3 hari=15, ≤7 hari=5, lainnya/tanpa due=0
R (rutin):             todo hasil recurring = 10, lainnya = 0
```
Urutan seri (tie-break): `due_at` paling awal (null paling akhir) → `scheduled_date` paling awal → `created_at` paling awal.
"besok", "≤3 hari" dihitung dari tanggal lokal `due_at` dibanding hari ini.

## 6. Todo berulang (recurring)
### Aturan pola (subset RRULE RFC 5545)
- Disimpan di template: `recurrence_rule` (string RRULE) + `scheduled_date` sebagai DTSTART (tanggal mulai).
- Yang didukung:
  - `FREQ=DAILY;INTERVAL=n`
  - `FREQ=WEEKLY;INTERVAL=n;BYDAY=MO,TH` (≥1 hari)
  - `FREQ=MONTHLY;INTERVAL=n;BYMONTHDAY=d` (d = 1..31)
  - `FREQ=MONTHLY;INTERVAL=n;BYDAY=1MO` (urutan 1..4 atau -1 = terakhir)
  - Penutup opsional: `UNTIL=YYYYMMDD` **atau** `COUNT=n`
- Selain itu ditolak (`422 RECURRENCE_UNSUPPORTED`).
- **Penyimpangan dari RFC**: `BYMONTHDAY=29..31` di bulan yang lebih pendek → jatuh ke **hari terakhir bulan itu** (RFC melewati bulan tersebut). Generator wajib menangani ini sendiri.

### Kejadian (occurrence)
- Kejadian = baris todo biasa dengan `recurrence_parent_id` = id template dan `scheduled_date` = tanggal kejadian.
- **ID kejadian** = `UUIDv5(NS_OCCURRENCE, "<template_id>:<YYYY-MM-DD>")`
  - `NS_OCCURRENCE = 13deda98-7bb3-46b1-acba-0b9a218d5dd6`
- **ID salinan sub-task / reminder** = `UUIDv5(NS_CHILD, "<occurrence_id>:<id_sumber_di_template>")`
  - `NS_CHILD = cd86087f-3314-4ffb-9126-3d90711aa607`
- Karena deterministik, membuat kejadian yang sama dua kali (dua device, retry) = idempotent (`INSERT ... ON CONFLICT (id) DO NOTHING`).

### Pembuatan kejadian
- Dibuat **saat dibaca** (endpoint Hari Ini/Terlewat/Rutin, atau saat sync di mobile), sampai **hari ini** (inklusif).
- Template menyimpan `generated_until` (DATE). Generator membuat semua kejadian dari `max(generated_until + 1, DTSTART, hari ini − 30)` sampai hari ini, lalu memperbarui `generated_until`.
  - Backfill maksimal 30 hari ke belakang supaya kejadian yang terlewat saat app tidak dibuka tetap ada (dan terhitung di kepatuhan).
- Saat kejadian dibuat: judul, catatan, workspace, prioritas, estimasi disalin dari template; sub-task template disalin; reminder offset template diubah jadi reminder absolut (`remind_at = waktu acuan − offset`).
- Tampilan "akan datang" (mis. 7 hari ke depan) dihitung virtual dari RRULE, **tidak** dibuat barisnya.

### Edit & berhenti
- Edit template → berlaku untuk kejadian yang **belum dibuat**. Kejadian yang sudah ada tidak berubah.
- Edit satu kejadian → hanya kejadian itu.
- Berhenti: set `UNTIL` ke hari ini (atau hapus/soft delete template). Kejadian lama tetap ada untuk histori.

## 7. % Kepatuhan
- Jendela: 30 hari lokal terakhir, `[hari ini − 29, hari ini]`.
- **Komitmen** yang dihitung = todo (bukan template, bukan sub-task) yang **tanggal komitmennya** (§2) ada di dalam jendela — berlaku untuk todo biasa maupun kejadian recurring
- Nilai tiap komitmen:
  | Kondisi | Poin | Masuk pembagi? |
  |---|---|---|
  | `done` tepat waktu | 1.0 | ya |
  | `done` terlambat | 0.5 | ya |
  | terlewat (§2) | 0 | ya |
  | `skipped` / `cancelled` | — | **tidak** |
  | masih terbuka dan belum terlewat (mis. jatuh tempo nanti sore) | — | **tidak** |
- `kepatuhan = round(Σ poin / jumlah komitmen × 100)`; kalau pembagi 0 → `null` ("belum ada data").
- Lingkup: per template recurring (kejadian-kejadiannya), per workspace, per user.
- Sub-task tidak dihitung.

### Riwayat terlewat (`todo_misses`)
- Setiap kali `scheduled_date` atau `due_at` sebuah todo diubah **saat todo itu sedang terlewat** (lewat aksi "Jadwalkan ulang" maupun edit biasa), simpan baris `todo_misses`: `todo_id`, `commitment_date` (tanggal komitmen lama), `user_id`, `workspace_id`.
- ID = `UUIDv5(NS_CHILD, "<todo_id>:miss:<YYYY-MM-DD>")` → idempotent antar device.
- Setiap baris `todo_misses` dengan `commitment_date` di dalam jendela dihitung sebagai **terlewat (0 poin, masuk pembagi)**, terpisah dari komitmen todo saat ini.
- Mengubah tanggal todo yang **belum** terlewat tidak membuat baris apa pun.
- `todo_misses` ikut ditampilkan di rekap jurnal hari `commitment_date` (sebagai "terlewat, dijadwalkan ulang ke ...").

## 8. Timer & TimeLog
- Maksimal **satu** timer berjalan per user (`ended_at IS NULL`). Memulai timer baru otomatis menghentikan yang berjalan (`ended_at = now`).
- Timer bisa terhubung ke todo (opsional). Kalau terhubung, `workspace_id` diambil dari todo; todo `pending` berubah jadi `in_progress`.
- Durasi = `ended_at − started_at`, atau `duration_minutes` untuk entri manual (manual: `started_at` = waktu acuan, `ended_at` null, `duration_minutes` wajib).
- Atribusi hari: tanggal lokal `started_at` (TimeLog yang melewati tengah malam tidak dipecah — cukup untuk MVP).
- TimeLog dengan durasi > 16 jam ditandai `suspicious` di respons (kemungkinan lupa stop) — tidak diblokir.

## 9. Jurnal harian
- Satu `journal_entry` per user per tanggal (dibuat otomatis saat pertama kali ditulis: mood atau catatan).
- `mood`: integer 1–5 (opsional).
- **Rekap hari X** (dihitung, tidak disimpan):
  - todo `done` dengan tanggal lokal `completed_at` = X (tandai yang terlambat)
  - todo yang terlewat dengan tanggal komitmen X, dan `skipped` dengan tanggal komitmen X
  - total menit TimeLog per workspace dengan tanggal lokal `started_at` = X
  - kepatuhan khusus hari X (rumus §7 dengan jendela 1 hari)
- Filter per workspace: catatan + todo + TimeLog dengan `workspace_id` tersebut.
- Catatan → todo: membuat todo baru (`source = journal`, `source_note_id` = id catatan), catatan tetap ada.

## 10. Reminder
- Reminder absolut: `remind_at` (UTC). Reminder template: `offset_minutes` relatif terhadap `due_at` kejadian, atau terhadap jam 09:00 lokal di `scheduled_date` kalau kejadian tidak punya `due_at`.
- Reminder terkirim sekali (`sent_at`). Reminder untuk todo yang sudah `done/cancelled/skipped` tidak dikirim.
