# Pitch script - KIPhone

10 slide, target 6 menit. Versi 3 menit ada di bagian akhir.

## Sebelum naik panggung

```sh
sh run-local.sh
# buka http://127.0.0.1:8088
```

Kalau API sudah jalan dan data ingin diulang:

```sh
curl -s -X POST localhost:8080/api/seed/load -H 'X-Dev-User: admin'
curl -s -X POST localhost:8080/api/detect    -H 'X-Dev-User: admin'
curl -s -X POST localhost:8080/api/tx/seed  -H 'X-Dev-User: admin'
curl -s -X POST localhost:8080/api/tx/detect -H 'X-Dev-User: admin'
```

Cek cepat sebelum naik:

```sh
curl -s localhost:8080/api/stats    -H 'X-Dev-User: admin'
curl -s localhost:8080/api/tx/stats -H 'X-Dev-User: admin'
```

Angka yang harus konsisten:

| Bagian | Angka |
|---|---|
| Penerima KIP-K 2026 | 1.047.221 penerima, Rp15,32 triliun (DIPA 2026) |
| Penyaluran | 215 record, 15 temuan, R1-R4, ~21 ms, P@10 1.00, recall 1.00, FP 0% |
| Penggunaan dana | 5.411 transaksi, 20 temuan, 5 prioritas tinggi, T1-T5, ~67 ms |
| Evaluasi penggunaan | P@10 1.00, recall 0.88, FP 1.5%, 17 pola ditanam, 3 false positive |
| Bukti batch | hash SHA-256, transaksi Memo Solana devnet |

---

## 01. Trust - 0:00-0:35

Layar: slide 1, dua screenshot dan headline.

Katakan:

> Kami tidak akan mencoba membuktikan screenshot ini benar. Kami akan menunjukkan kenapa screenshot ini bisa menjadi satu-satunya "bukti" yang publik punya.
>
> Dua postingan anonim, 6 dan 12 November. Isinya: mahasiswa KIP-K dengan iPhone 17 Pro Max. Entah postingannya benar atau tidak, percakapannya sudah berubah. Dari "apakah dia berhak?" menjadi "semua penerima KIP-K begini".
>
> Satu screenshot. Sejuta penerima dihakimi. Masalahnya bukan iPhone. Masalahnya tidak ada yang memeriksa polanya.

Transisi:

> Kalau kita tidak bisa menunjukkan ke mana uang pergi, publik akan mengisi kekosongan itu dengan screenshot.

---

## 02. Pola - 0:35-1:05

Layar: slide 2, Ilham dan tiga langkah.

Katakan:

> Di balik screenshot ada pola. Ilham, ilustrasi dari data sintetis. Rp2,4 juta masuk. Dalam sepuluh hari, Rp1,8 juta keluar: taruhan kecil berulang, transfer ke satu agen, dan pembelian di Apple Store.
>
> Tidak ada yang melihat, karena kampus tidak pernah menerima mutasi. Postingan viral melihat satu iPhone. Tidak ada yang melihat 5.411 transaksi.

Transisi:

> Jadi masalahnya bukan satu postingan. Masalahnya tiga lubang sistemik.

---

## 03. Masalah - 1:05-1:35

Layar: slide 3, tiga kolom.

Katakan:

> Lubang pertama: tidak ada trace end-to-end. SK, rekening, dan transaksi hidup di sistem berbeda. Tidak ada satu baris dari dana cair ke merchant.
>
> Lubang kedua: pola penyalahgunaan lolos. Taruhan, dana habis tujuh hari, tarik tunai terkonsentrasi, barang mewah, transfer berulang ke calo. Semua terlihat dari mutasi, tapi kampus tidak melihat mutasi.
>
> Lubang ketiga: audit tanpa bukti keaslian. Saat diperiksa, tidak ada cara membuktikan data yang dilihat auditor sama dengan data saat pencairan.
>
> Setiap rupiah yang bocor adalah kuota yang hilang untuk mahasiswa miskin yang berhak.

---

## 04. Solusi - 1:35-2:05

Layar: slide 4, tiga pilar.

Katakan:

> Kami membangun tiga hal.
>
> Trace: setiap temuan tertaut ke mahasiswa, rekening, dan transaksi. Ada merchant, nominal, tanggal, dan evidence JSON.
>
> Detect: R1-R4 untuk siapa yang menerima. T1-T5 untuk ke mana uangnya pergi. Deterministik, dengan alasan per aturan.
>
> Prove: SHA-256 dari batch penyaluran, di-anchor sebagai memo di Solana devnet. Halaman publik menampilkan agregat dan hash, tanpa nama.
>
> Scope-nya jelas: data mutasi datang dari feed bank mitra, dummy di demo ini. Bukan kampus membuka rekening mahasiswa.

---

## 05. Penggunaan dana - 2:05-2:45

Layar: slide 5, lima pola dan statistik.

Katakan:

> Ini lima pola yang kami tandai.
>
> T1 taruhan, bobot 0,50. T2 dana habis dalam tujuh hari, 0,30. T3 tarik tunai terkonsentrasi, 0,15. T4 barang mewah di atas Rp1,5 juta, 0,20. T5 transfer berulang ke satu penerima atau calo, 0,25.
>
> Dari 5.411 transaksi, muncul 20 temuan, 5 di antaranya prioritas tinggi. Klik satu temuan: ada bukti per pola, lalu mutasi rekening dengan running balance, grafik harian, dan kategori pengeluaran.
>
> Evaluasi pada 17 pola sintetis: precision@10 1,00, recall 0,88, false positive 1,5 persen. Feed bank masih dummy. Ini rujukan verifikasi, bukan tuduhan.

---

## 06. Demo - 2:45-4:15

Layar: dashboard, lalu penggunaan dana, lalu mutasi.

Urutan klik:

1. Dashboard penyaluran.
   > "215 record, 15 temuan, empat aturan R1-R4."

2. Klik **Penggunaan dana →**.
   > "5.411 transaksi, 20 temuan, 5 prioritas tinggi."

3. Klik **Jalankan Deteksi Pola**.
   > "Lima pola, 67 milidetik. Tidak ada kotak hitam."

4. Klik **Lihat bukti** pada temuan prioritas tinggi.
   > "T1 plus T2: empat transaksi taruhan, mayoritas dana keluar dalam tujuh hari. Di data ini, salah satu contohnya 86 persen."

5. Klik **Lihat mutasi rekening**.
   > "Ini mutasinya. Running balance, grafik harian, kategori, dan daftar transaksi lengkap."

6. Klik **Tandai Salah Deteksi**.
   > "Kalau kami salah, kami bilang salah. Statusnya berubah. Tidak ada dana yang diblokir otomatis."

7. Kembali ke landing, klik **Halaman publik ↗**.
   > "Publik hanya melihat agregat dan hash batch. Tidak ada nama, NIM, rekening, atau temuan per orang."

Kalimat penutup demo:

> Model tidak memblokir apa pun. Manusia yang memutuskan.

Fallback: video 90 detik dan screenshot mutasi. Jangan bergantung pada WiFi.

---

## 07. Kenapa bukan blockchain penuh - 4:15-4:50

Layar: slide 7.

Katakan:

> Kami memakai chain untuk membuktikan data, bukan menyimpan data.
>
> NIK, NIM, rekening, nominal, dan merchant adalah data pribadi sensitif. UU PDP 27/2022 melarang itu masuk ledger publik. Right to erasure juga bertabrakan dengan immutability.
>
> Jadi yang kami anchor hanya satu hash batch penyaluran. Nol PII on-chain. Temuan tidak pernah on-chain. Halaman publik memakai hash itu untuk membuktikan batch tidak diubah.
>
> Hari ini satu hash per batch. Merkle proof per record ada di roadmap, bukan klaim demo.

---

## 08. Arsitektur - 4:50-5:20

Layar: slide 8.

Katakan:

> Arsitekturnya sengaja kecil. Preact di depan, satu binary Go di tengah, Postgres 16 di belakang.
>
> R1-R4 dan T1-T5 jalan di proses yang sama. Feed bank mitra masih dummy. Anchor Solana opsional dan manual, tidak ada di jalur kritis.
>
> Tidak ada ML, tidak ada Redis, tidak ada panggilan jaringan di jalur demo. Setiap komponen tambahan adalah titik gagal di panggung.
>
> Produksi: CSV atau REST dari kampus, plus feed bank mitra dengan perjanjian berbagi data.

---

## 09. Privasi dan etika - 5:20-5:50

Layar: slide 9.

Katakan:

> Kami tidak membuka rekening langsung. Data mutasi hanya dari feed bank mitra. Demo memakai data dummy. Produksi butuh perjanjian berbagi data dan dasar hukum.
>
> Tidak ada blokir otomatis. Tidak ada fitur demografis. Salah deteksi adalah status kelas satu, bukan catatan tersembunyi.
>
> Halaman publik hanya agregat dan hash. Tanpa nama, NIM, rekening, atau temuan per orang.
>
> Consent eksplisit dan portal transparansi ada di roadmap. Kami tidak mengklaim sudah dibangun.

---

## 10. Dampak dan penutup - 5:50-6:20

Layar: slide 10.

Katakan:

> Dampaknya untuk empat pihak. Kemendikbud dapat bukti per batch dan sinyal dini. Kampus dapat antrean verifikasi ber-bukti. Mahasiswa yang sah tidak dihakimi screenshot. Bank dan regulator dapat feed pola lewat kerja sama, bukan akses langsung.
>
> Kami minta pilot satu kampus, satu batch, 30 hari.
>
> Kami tidak menghakimi screenshot. Kami memeriksa pola. Manusia yang memutuskan.

---

## Versi 3 menit

| Waktu | Bagian | Inti |
|---|---|---|
| 0:00-0:30 | Trust | Dua screenshot, "sejuta penerima dihakimi", masalahnya bukan iPhone |
| 0:30-1:00 | Masalah | Ilham, tiga lubang: trace, pola, audit |
| 1:00-1:40 | Solusi | Trace, detect, prove; R1-R4 + T1-T5 |
| 1:40-2:40 | Demo | Penggunaan dana → prioritas tinggi → mutasi → salah deteksi |
| 2:40-3:00 | Privasi + tutup | Feed bank mitra, no auto-block, halaman publik, pilot 30 hari |

---

## Q&A

**"Bisa mendeteksi iPhone atau barang mewah?"**
> T4 menandai pembelian di atas Rp1,5 juta di merchant mewah dari feed bank mitra. Di data demo ada transaksi APPLE STORE. Tapi feed-nya dummy dan hasilnya sinyal untuk verifikasi, bukan bukti. Kami tidak menghakimi dari screenshot.

**"Ini kan mengawasi mahasiswa?"**
> Kami tidak membuka rekening langsung. Data datang dari feed bank mitra, hanya pola yang ditandai yang masuk daftar periksa, dan tidak ada tindakan otomatis. Halaman publik hanya agregat dan hash. Produksi butuh perjanjian berbagi data dan dasar hukum.

**"Datanya dari mana?"**
> Penyaluran dari sistem kampus: SK, NIM/NIK, rekening, status akademik. Transaksi dari feed bank mitra, dummy di demo ini. Integrasi produksi lewat CSV/REST kampus dan kerja sama bank.

**"Kenapa tidak pakai ML?"**
> R1-R4 dan T1-T5 deterministik supaya bisa diaudit baris per baris. Isolation Forest ada di roadmap ketika data asli dan labelnya tersedia. LLM tidak memutuskan siapa curang.

**"False positive-nya bagaimana?"**
> Pada data sintetis: 20 temuan dari 17 pola ditanam, 3 false positive, FP 1,5 persen. Threshold sengaja konservatif. Setiap temuan punya tombol salah deteksi, dan tidak ada dana yang diblokir otomatis.

**"Kenapa blockchain?"**
> Hanya untuk membuktikan batch penyaluran tidak diubah. Hash saja, tanpa PII, tanpa temuan. Halaman publik bisa memverifikasi hash yang sama.

**"Consent-nya mana?"**
> Belum dibangun. Itu roadmap. Demo memakai data sintetis. Produksi butuh perjanjian berbagi data dan dasar hukum yang jelas.

---

## Jangan klaim

- Jangan bilang screenshot adalah bukti fraud.
- Jangan bilang data demo adalah data nyata.
- Jangan bilang detektor menangkap semua judi; merchant bisa disamarkan.
- Jangan bilang memakai ML.
- Jangan bilang Merkle proof per record sudah ada.
- Jangan bilang portal consent sudah dibangun.
