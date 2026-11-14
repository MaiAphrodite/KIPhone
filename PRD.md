# PRD — KIPhone: Deteksi Anomali Penyaluran KIP-K

**Version** 0.1 · **Date** 2026-11-14 · **Status** build-ready
**Team** 2 people + agent fleet · **Event** 36-hour hackathon

> One-liner: dashboard integritas yang mendeteksi penyimpangan **penyaluran**
> dana KIP-K sebelum merugikan mahasiswa yang berhak — dengan bukti, bukan tuduhan.

---

## 1. Problem

Dana KIP Kuliah cair ke ratusan ribu rekening. Verifikasi manual tidak mungkin
skala-nya, jadi fraud klasik lolos: penerima ganda, rekening penampung/joki,
mahasiswa nonaktif yang masih menerima, dan rekening yang tidak cocok dengan
nama penerima. Setiap rupiah yang bocor = kuota yang hilang untuk mahasiswa
miskin yang berhak.

## 2. Non-goals (locked — do not reopen)

| Rejected | Why |
|---|---|
| Deteksi belanja/judi mahasiswa | Kampus tidak punya data mutasi (UU Perbankan Pasal 40); tidak mungkin dioperasikan |
| Flag kecurangan di blockchain | Tuduhan permanen tanpa due process = fitnah yang tidak bisa dihapus |
| LLM memutuskan siapa curang | Model menjelaskan temuan; keputusan selalu manusia |
| Pemblokiran otomatis | Nol tindakan otomatis terhadap pencairan |
| Pengawasan hanya ke penerima beasiswa | Framing anti-mahasiswa yang kalah di ruang juri |

## 3. Users

1. **Staf kemahasiswaan/keuangan** — menerima daftar periksa ber-bukti,
   memverifikasi, menandai temuan.
2. **Juri** — menilai 60 detik demo + slide.

## 4. Product

Single dashboard: import batch penyaluran → jalankan deteksi → daftar temuan
dengan skor risiko → panel bukti per temuan → status lifecycle yang tercatat.

### 4.1 Detection rules (4, deterministic)

| # | Rule | Logic | Evidence fields | Seeder plants |
|---|---|---|---|---|
| R1 | Penerima ganda | NIK atau NIM muncul di >1 SK, lintas periode/kampus | both NIK/NIM, SK ids, periods | 4 |
| R2 | Rekening bersama | >1 mahasiswa pada nomor rekening sama (union-find component ≥ 2) | account no., member list, ages | 4 |
| R3 | Nama tidak cocok | Jaro-Winkler(nama_SK, nama_rekening) < 0.85 | both names, score, token diff | 4 |
| R4 | Nonaktif menerima | status akademik ∈ {cuti, nonaktif, lulus} saat pencairan | status, last active date, SK date | 3 |

### 4.2 Scoring

```
rule_hit = 1 per rule triggered
risk = (0.30·R1 + 0.30·R2 + 0.25·R3 + 0.15·R4)  → 0..1
band:  risk ≥ 0.55  →  "perlu verifikasi"   (amber)
       risk ≥ 0.80  →  "prioritas tinggi"   (red)
       else         →  bersih               (green)
```

Thresholds deliberately conservative: the cost of a false positive (dana
mahasiswa miskin tertahan) exceeds the cost of a missed case (ditemukan di
siklus berikutnya). This sentence goes on the slide, verbatim.

### 4.3 Explanation

Deterministic templates per rule, e.g. R3:
`"Nama rekening ('{a}') berbeda {n}/4 token dari nama SK ('{b}'); skor kemiripan {s}."`
No network call in the demo path. (Optional stretch: LLM-polished narrative,
pre-computed, same field.)

### 4.4 Lifecycle (human-in-the-loop, mandatory)

```
baru → diverifikasi → selesai
  └──────→ salah_deteksi ────┘
```

Every transition recorded (actor, timestamp, note). `salah_deteksi` is a
first-class, visible state — it is the honesty signal for judges.

## 5. Data model

| Table | Key fields |
|---|---|
| `students` | id, nim, nik, nama, prodi, status_akademik, last_active_at |
| `accounts` | id, student_id, nomor_rekening, nama_pemilik, created_at |
| `disbursements` | id, student_id, account_id, sk_nomor, periode, nominal, tanggal_cair |
| `findings` | id, disbursement_id, rule, risk_score, band, explanation, evidence(jsonb), status, created_at |
| `finding_events` | id, finding_id, from_status, to_status, actor, note, at |

Chain (stretch only): store `sha256(batch canonical json)` per periode in
`batches.chain_hash`, displayed as "bukti batch". Never store findings on it.

## 6. Synthetic data

`backend/seed/` generates **215 records**: 200 clean + 15 planted (R1×4, R2×4,
R3×4, R4×3), labels in `seed/labels.json` for evaluation. Deterministic seed
(`INTEGRA_SEED`, default 42) so the demo is reproducible on any machine.

## 7. API

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/seed/load` | wipe + load synthetic batch |
| POST | `/api/detect` | run 4 rules over current batch, write findings |
| GET | `/api/findings?status=&band=` | list findings w/ joined student+account |
| GET | `/api/findings/{id}` | one finding + full evidence |
| POST | `/api/findings/{id}/status` | lifecycle transition + event record |
| GET | `/api/stats` | counts for dashboard cards |
| GET | `/api/eval` | precision@10, recall vs labels.json, FP rate |

Auth: hackathon demo runs dev mode (`X-Dev-User`), same as existing scaffold.
Supabase JWT path stays as-is for the "production" slide.

## 8. Screens

1. **Dashboard** — cards: total record, temuan, prioritas tinggi, salah deteksi;
   one button "Jalankan Deteksi".
2. **Findings list** — rows: NIM, nama, rule badge, risk bar, band, status.
3. **Evidence drawer** — per-rule evidence block, explanation, lifecycle buttons
   (`Verifikasi`, `Tandai Salah Deteksi`, `Selesaikan`) + event history.
4. **Eval strip** (optional bottom of dashboard) — the 3 numbers from §9.

Style: `tokens.css` only (linear-app dark). No hardcoded colors.

## 9. Evaluation (the slide that wins)

| Metric | Target | Source |
|---|---|---|
| Precision@10 | ≥ 0.8 | `/api/eval` over labels.json |
| Recall (15 planted) | ≥ 0.7 | same |
| False positive rate | < 5% | same |
| Detection runtime (215 records) | < 50 ms | measured, printed in response |

Presented as "methodology on synthetic data", explicitly — not claimed accuracy
on real data.

## 10. Demo script (60 s, frozen)

1. Dashboard: 215 record, semua hijau.
2. Klik **Jalankan Deteksi** → muncul 5 temuan.
3. Buka temuan R2: dua mahasiswa, satu rekening, rekening berumur 4 hari.
4. Panel bukti → klik **Tandai Salah Deteksi** → status berubah, event tercatat.
5. Closing line: *"Model tidak memblokir apa pun. Manusia yang memutuskan,
   dan keputusan itu tercatat."*

## 11. Roadmap slide (one line each, do not build)

- Kerja sama bank penyalur/QRIS untuk data transaksi (butuh kemitraan resmi).
- Modul kesejahteraan finansial opt-in (perlindungan diri, bukan investigasi).
- Isolation Forest/graph embedding ketika data asli tersedia.
- On-chain batch receipts untuk audit publik (hash saja, tanpa temuan).

## 12. Risks

| Risk | Mitigation |
|---|---|
| "Datanya asli dari mana?" | Slide: data sintetis berlabel; integrasi = CSV/REST dari sistem kampus |
| Demo network failure | Full offline: local Postgres, no LLM call, no CDN |
| Scope creep ke 8 rules | Locked at 4; the rest are roadmap bullets |
| Judge asks re: surveillance | Framing: melindungi mahasiswa yang benar dari penerima fiktif |

## 13. Build split (accepted plan)

| Owner | Deliverable |
|---|---|
| This session | `backend/anomaly/` (4 rules, scorer, evidence), `backend/seed/`, findings API |
| Orca worker FE | Dashboard + findings list + evidence drawer (Preact, tokens only) |
| Orca worker QA | Go tests for rules + scorer, eval endpoint test, BUGS.md |
| OD | Pitch deck (1 pass, after demo freeze) |
