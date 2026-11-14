# DEMO — 60 detik

Panduan lisan + urutan klik. Cetak atau buka di layar kedua.

## Sebelum naik panggung (2 menit)

```sh
docker compose up -d                      # postgres
cd backend && DATABASE_URL="postgresql://postgres:postgres@127.0.0.1:5432/hackstack" PORT=8080 ./backend &

curl -s -X POST localhost:8080/api/seed/load -H 'X-Dev-User: admin'   # 215 record, 15 ditanam
curl -s -X POST localhost:8080/api/detect    -H 'X-Dev-User: admin'   # 15 temuan

cd ../frontend && npm run dev              # http://localhost:5173
```

Buka browser di halaman landing. **Jangan** klik apa pun sampai demo mulai.

Cek cadangan: kalau `/api/eval` kosong, jalankan `detect` sekali lagi sebelum naik.

## Alur 60 detik

| Detik | Yang dilakukan | Yang diucapkan |
|---|---|---|
| 0–10 | Tampilkan landing page | "Dana KIP-K cair ke rekening pribadi. Kampus tidak bisa melihat ke mana perginya — dan kami tidak mencoba mengintip. Kami memeriksa **penyalurannya**." |
| 10–20 | Klik **Buka dashboard** | "215 record batch. Empat aturan deterministik: NIK ganda, rekening bersama, nama tidak cocok, penerima nonaktif." |
| 20–30 | Sorot 4 kartu statistik, lalu klik **Jalankan Deteksi** | "15 temuan dalam ~30 milidetik. Tidak ada model yang memutuskan — hanya aturan yang bisa diaudit." |
| 30–45 | Klik **Lihat bukti** pada baris teratas (prioritas tinggi, R1+R2+R3) | "Dua mahasiswa memakai satu rekening. Rekening itu baru berumur 4 hari saat pencairan. Nama pemiliknya tidak cocok dengan SK. Ini tiga bukti terpisah, bukan satu skor." |
| 45–55 | Klik **Tandai Salah Deteksi**, isi catatan "rekening keluarga", kirim | "Kalau kami salah, kami bilang salah. Statusnya tercatat, dengan nama pemeriksa. Tidak ada tuduhan permanen." |
| 55–60 | Tutup drawer, lihat strip evaluasi | "Precision@10 satu koma nol, recall satu koma nol — **pada data sintetis berlabel**. Kami tidak mengklaim akurasi data asli. Lalu: model tidak memblokir apa pun. Manusia yang memutuskan, dan keputusan itu tercatat." |

## Kalimat penutup (hafalkan)

> "Model tidak memblokir apa pun. Manusia yang memutuskan, dan keputusan itu tercatat."

## Kalau ada yang gagal di panggung

| Gejala | Tindakan |
|---|---|
| Tabel kosong | `curl -X POST localhost:8080/api/detect -H 'X-Dev-User: admin'` lalu refresh |
| Dashboard error merah | Backend mati → jalankan ulang `./backend`, refresh |
| Angka statistik 0 semua | `POST /api/seed/load` lalu `POST /api/detect` |
| Postgres mati | `docker compose up -d`, tunggu 5 detik, ulangi seed+detect |
| Semua gagal | Buka `design/landing.html` (artifact OD statis) dan lanjut narasi lisan |

## Jawaban untuk pertanyaan juri

**"Datanya dari mana?"**
> Batch penyaluran dari sistem kampus: SK penerima, NIM/NIK, nomor rekening, status akademik. Untuk demo kami memakai 215 record sintetis berlabel. Integrasi produksi = impor CSV atau REST dari sistem yang sudah ada.

**"Kenapa tidak mendeteksi judi online?"**
> Karena datanya tidak ada. Mutasi rekening dilindungi UU Perbankan Pasal 40, dan kampus tidak melihat merchant. Membangun fitur itu berarti membangun produk yang tidak bisa dioperasikan. Kami memilih titik di mana data benar-benar tersedia dan dampaknya lebih besar.

**"Bagaimana kalau false positive?"**
> Konsekuensinya dana mahasiswa miskin tertahan. Karena itu threshold kami konservatif: lebih memilih "tidak yakin" daripada menuduh. Setiap temuan punya tombol salah deteksi, dan statusnya terbuka.

**"Ini kan mengawasi mahasiswa."**
> Sebaliknya — penerima fiktif menggeser kuota orang yang berhak. Kami melindungi penerima yang benar. Dan tidak ada satu pun tindakan otomatis terhadap mahasiswa.

**"Kenapa tidak pakai LLM untuk memutuskan?"**
> Karena keputusan harus bisa dipertanggungjawabkan. Empat aturan deterministik bisa diaudit baris per baris. LLM dipakai hanya untuk menulis penjelasan, bukan menilai orang.
