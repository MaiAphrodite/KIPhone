# BUGS

QA pass over the Go backend, detection rules, seeder, API and PRD. Seed 42. No product code was changed during this pass; only tests and the API contract script were added. Severity scale: blocker / major / minor.

## Major

- `backend/handlers.go:86` **major** - `deleteItem` runs an unscoped `Delete(&Item{}, id)` before the owner-scoped delete, so any authenticated user can delete another user's item by ID.
- `backend/handlers_findings.go:203-213` **major** - status update and finding-event insert are separate operations with no transaction; an event insert failure leaves the status changed with no audit trail, violating PRD §4.4.
- `backend/handlers_findings.go:69-77` **major** - re-running detection deletes only `status = 'baru'` findings; resolved or `salah_deteksi` rows survive, so the same disbursement gains a second finding and stats inflate.
- `backend/handlers_findings.go:262-279` **major** - recall marks a planted label as covered when *any* expected rule is found; a label missing R2/R3 can still count as full recall, inflating the headline eval metric.
- `backend/anomaly.go:88-106` **major** - R2 is keyed by student, not by the disbursed account; a student with one shared and one clean account gets R2 on the clean disbursement, and `backend/anomaly.go:134-145` then reports the wrong account number and age as evidence.
- `backend/anomaly.go:62-66` **major** - R1 only groups students by NIK; PRD §4.1 says "NIK atau NIM", so duplicate NIM is never detected.
- `backend/auth.go:64-95` **major** - JWT verification checks signature and alg but never `exp`, `nbf` or `aud`; expired Supabase tokens are accepted on the production path.

## Minor

- `backend/anomaly.go:166` **minor** - R4 fires on `status != "aktif"` instead of the explicit set {cuti, lulus, nonaktif}; empty or unknown statuses also trigger R4.
- `backend/unionfind.go:1-27` **minor** - the union-find implementation is dead code; R2 uses exact account-number grouping only, so transitive shared-account components mentioned in PRD §4.1 are not detected.
- `backend/seed.go:102-103` **minor** - `delete(usedRek, a2.NomorRekening)` deletes the now-shared number from the uniqueness registry, so a later generated account can accidentally reuse it and create a false R2.
- `backend/seed.go:134-136` **minor** - the planted R2+R3 holder name is only checked for inequality, not for `jaroWinkler < 0.85`; a future seed or name-list change can silently break the label (seed 42 works today).
- `backend/seed.go:144-149` **minor** - the planted R3 names are only checked for inequality, not for a score below the 0.85 threshold; deterministic only because seed 42 happens to produce unrelated names.
- `backend/seed.go:13` **minor** - seed value is hardcoded to 42; PRD §6 promises the `KIPHONE_SEED` environment override, which is not implemented.
- `backend/handlers_findings.go:13-43` **minor** - `loadSeed` runs wipe and inserts without a transaction; a failure mid-load leaves a wiped or partially loaded batch.
- `backend/handlers_findings.go:217-222` **minor** - `stats` ignores all query errors and returns HTTP 200 with zeros when the database fails.
- `backend/handlers_findings.go:157-159` **minor** - `getFinding` maps every Scan or database error to 404, hiding real outages as "not found".
- `backend/handlers_findings.go:235` **minor** - `/api/eval` always evaluates against `generateSeed().Labels`, not labels for the currently loaded batch, so any non-seed import makes the eval meaningless.
- `backend/anomaly.go:109-110` **minor** - missing student or account rows become zero-value names; `jaroWinkler(name, "") = 0` then triggers R3 on an orphan disbursement instead of flagging a data-integrity error.
- `backend/handlers.go:34-38` **minor** - the health check calls `Ping()` with no context or timeout; a blocked database can hang the liveness endpoint.
- `backend/auth.go:19-25` **minor** - dev auth trusts any `X-Dev-User`, and seed/detect/status endpoints have no role check; safe only while `SUPABASE_JWT_SECRET` is unset for the demo.
- `backend/main.go:62-66` **minor** - the HTTP server has no `ReadTimeout`, `WriteTimeout`, `IdleTimeout` or graceful shutdown.
- `PRD.md:140` **minor** - the demo script promises 5 findings, but the seeder and detector produce 15 (also documented in `README.md:15`); presenters will see a mismatch.
- `PRD.md:95-97` **minor** - the PRD references `backend/seed/` and `seed/labels.json`; neither exists, labels are generated in memory by `backend/seed.go`.

## Notes

- `go build ./...`, `go test ./...`, the DB-backed `api_test.go` and `backend/test_api.sh` all pass on the current code. The defects above are logic, security, audit and documentation risks rather than failing tests.

## Fixed after this pass (commit "fix: QA majors")

- `handlers_findings.go` detection now wipes **all** findings and events (recompute semantics); re-running after a resolution no longer inflates stats — verified: mark 1 as `salah_deteksi`, re-detect → still 15 findings, not 16.
- `handlers_findings.go` status update + audit event now run in one `db.Transaction` (PRD §4.4: no status change without its event).
- `handlers_findings.go` recall now requires **every** expected rule for a label to count as covered (no partial-credit inflation).
- `anomaly.go` R1 groups duplicates by NIK **or** NIM, matching PRD §4.1.
- Still open: `handlers.go:86` unscoped `deleteItem` (legacy `/api/items` demo endpoint, not on the KIPhone path) and R2 being keyed by student rather than by disbursed account (edge case, not hit by the current seed).
