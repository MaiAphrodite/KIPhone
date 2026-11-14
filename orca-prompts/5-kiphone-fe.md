You are the FRONTEND worker for the KIPhone hackathon project. Repo: this worktree (Vite + Preact, no TypeScript).

READ FIRST: PRD.md (sections 4, 8, 10), frontend/src/app.jsx, frontend/src/lib/api.js, frontend/src/styles/tokens.css, frontend/src/styles/app.css.
STYLE RULE: use ONLY CSS variables from tokens.css. Never hardcode a hex color. Dark, Linear-like, precise, low-chrome.

The Go API is RUNNING at http://127.0.0.1:8080 (already seeded, 15 findings). Endpoints:
- GET  /api/stats              -> {records, findings, prioritas_tinggi, perlu_verifikasi, rendah, baru, salah_deteksi}
- GET  /api/findings           -> {findings:[{id,nim,nama,prodi,rule,risk_score,band,explanation,evidence,status,sk_nomor,periode,nominal,tanggal_cair,nomor_rekening}]}
- GET  /api/findings/{id}      -> {finding, events:[{from_status,to_status,actor,note,at}]}
- POST /api/findings/{id}/status  body {to_status, note}  (valid: baru->diverifikasi, baru->salah_deteksi, diverifikasi->selesai, verifikasi->salah_deteksi, salah_deteksi->selesai)
- POST /api/detect             -> {findings_created, by_rule, runtime_ms}
- POST /api/seed/load          -> reloads the batch
- GET  /api/eval               -> {precision_at_10, recall, false_positive_rate, detection_runtime_ms, planted, covered}
All calls need header: X-Dev-User: admin (lib/api.js already sends it in dev mode).

BUILD THESE SCREENS (PRD §8):

1. Dashboard (/):
   - Four stat cards: Total Record, Temuan, Prioritas Tinggi, Salah Deteksi.
   - One primary button "Jalankan Deteksi" -> POST /api/detect, then refresh; show runtime_ms as a caption ("15 temuan · 27 ms").
   - Below: an "Eval" strip with precision@10, recall, FP rate from /api/eval, labelled "Metodologi: data sintetis berlabel (seed 42)".
   - Small muted link "Muat ulang data contoh" -> POST /api/seed/load.

2. Findings list (same page, below stats):
   - Table rows: NIM, Nama, Rule badges (R1..R4 as small pills), risk bar (width = risk_score), band label (RED "prioritas tinggi", AMBER "perlu verifikasi"), status pill, and a "Lihat bukti" button.
   - Sort by risk_score desc. Filter chips: Semua / Prioritas / Perlu verifikasi / Salah deteksi.

3. Evidence drawer (right side panel or modal, your call):
   - Header: nama + NIM + periode + nominal.
   - Rule badges + risk score + band.
   - Evidence blocks: parse finding.evidence (JSON string) and render each rule's fields as a definition list (R1: NIK, claimed_by; R2: nomor_rekening, dipakai_oleh, rekening_umur_hari; R3: nama_sk, nama_rekening, nama_match_score, token_berbeda; R4: status_akademik, last_active_at, tanggal_cair).
   - The explanation sentence in a highlighted quote block.
   - Lifecycle buttons: "Verifikasi", "Tandai Salah Deteksi", "Selesaikan" (show only allowed transitions). Each asks for an optional note, POSTs, then refreshes.
   - Event history list at the bottom: "baru -> salah_deteksi · verifikator1 · note · time".

HARD REQUIREMENTS:
- Keep the existing app.jsx shell working (it has a Landing screen at startup with onStart -> setView('app')). Replace the old items/list demo with the KIPhone dashboard; the Landing screen can stay as the hero shown first with its CTA leading into the dashboard.
- Split into files: src/screens/Dashboard.jsx, src/screens/Findings.jsx (or components/), src/lib/integra.js (API calls). Keep each file under ~200 lines.
- Handle loading, empty, and error states. If the API returns an error, show the message, never a blank screen.
- No new dependencies. No router library — a simple view state is enough.
- Indonesian UI copy (the demo audience is Indonesian). Numbers formatted with thousand separators.

FINISH BY:
- Running `npm run build` and pasting the output.
- Curling the API yourself to confirm the shapes you coded against: `curl -s http://127.0.0.1:8080/api/findings -H 'X-Dev-User: admin' | head -c 400`.
- Writing FRONTEND-NOTES.md with: files created, what you verified, anything you could not finish.
Do not touch Go files, PRD.md, or design/. Commit nothing.
