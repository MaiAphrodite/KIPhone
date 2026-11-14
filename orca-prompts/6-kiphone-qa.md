You are the QA worker for the KIPhone hackathon project. Repo: this worktree (Go backend + Preact frontend).

READ FIRST: PRD.md (sections 4, 7, 9), backend/anomaly.go, backend/jw.go, backend/seed.go, backend/handlers_findings.go.

TASK 1 — Go tests (backend/), table-driven, no database needed:
  a) jw_test.go: Jaro-Winkler sanity — identical names = 1.0, "Budi Santoso" vs "Budi Santosa" > 0.9, unrelated names < 0.6, empty string = 0. tokenDiff cases.
  b) anomaly_test.go: build small in-memory fixtures (no DB) and assert:
     - R1 fires when two students share a NIK
     - R2 fires when two accounts share a nomor_rekening, and the evidence contains rekening_umur_hari
     - R3 fires when the account holder name is unrelated (score < 0.85), does NOT fire on a 1-token surname difference
     - R4 fires for status cuti/lulus/nonaktif when tanggal_cair > last_active_at, does NOT fire for aktif
     - composite risk: a record hitting R1+R2+R3 = 0.85 and band prioritas_tinggi; a single R2 = 0.30 and band rendah
     - one Finding per disbursement with rules joined as "R1,R2,R3"
  c) seed_test.go: generateSeed() returns exactly 215 students, 215 accounts, 215 disbursements, 15 labels, and every label NIM exists in the students slice; running detect() over the seed yields >= 15 findings whose NIMs cover all 15 labels.

  Run `go build ./... && go test ./...` and paste the result. Use the existing module; add no dependencies beyond the stdlib.

TASK 2 — API contract test (backend/api_test.go):
  Use net/http/httptest against the handlers with a SQLite-in-memory or skip-if-no-DB guard. If wiring the DB is too heavy, instead write a shell script `backend/test_api.sh` that curls a running server on 127.0.0.1:8080 and asserts: /health 200, /api/stats has records>0 after seed, /api/detect returns findings_created>0, /api/findings returns 15 rows, a status transition baru->salah_deteksi returns 200 and records from_status "baru", an invalid transition returns 409. Print PASS/FAIL per check and exit non-zero on failure. Run it and paste the output.

TASK 3 — Write BUGS.md at the repo root:
  Every real defect or risk you find while reading/writing tests, one line each, with file:line and severity (blocker/major/minor). Be specific and honest — the team will fix these before the demo. If you fix something, note the fix and mark it fixed.

FINISH BY pasting: the go test output, the API script output, and a 3-line summary of findings. Do not refactor product code beyond what tests require. Commit nothing.
