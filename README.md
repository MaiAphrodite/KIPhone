# KIPhone: Deteksi Anomali Penyaluran dan Penggunaan Dana KIP-K

> Dashboard integritas yang menandai penyimpangan dana KIP-K dengan bukti yang bisa diaudit, bukan tuduhan

[![License](https://img.shields.io/badge/license-MIT-2ea44f?style=flat-square)](LICENSE)
[![Stack](https://img.shields.io/badge/stack-Go%20%2B%20Postgres%20%2B%20Preact-000?style=flat-square)](#tech-stack)

Dashboard integritas yang mendeteksi penyimpangan **penyaluran** dan pola
**penggunaan dana** KIP-K sebelum merugikan mahasiswa yang berhak, dengan
bukti, bukan tuduhan. Jalur penyaluran memakai empat aturan deterministik
(R1-R4), jalur penggunaan memakai lima aturan (T1-T5) di atas feed bank dummy.
Tidak ada model yang memblokir apa pun: setiap temuan membawa evidence, dan
status temuannya bisa dikoreksi manusia. Build spec: [PRD.md](PRD.md).

```
Preact (Vite, :5173)  --/api proxy-->  Go API (:8080)  --GORM-->  Postgres (local Docker or Supabase)
        |
        +--supabase-js (auth only: sign-in -> JWT)
```

The frontend never talks to Postgres. Supabase is used for auth only; all data
access goes through Go so business rules live in one place.

## Getting started

Prerequisites: Go 1.26+ (see `backend/go.mod`), Node with npm, and Docker, or a
Supabase Postgres you can point at.

```sh
# 1. database (local Postgres on 127.0.0.1:5432, from ./docker-compose.yml)
docker compose up -d

# 2. api
cd backend
cp .env.example .env
DATABASE_URL="postgresql://postgres:postgres@127.0.0.1:5432/hackstack" go run .

# 3. web
cd ../frontend
npm install
npm run dev          # http://localhost:5173
```

With `SUPABASE_JWT_SECRET` unset the API runs in dev mode and trusts the
`X-Dev-User` header, so the UI works before auth is wired.

## Demo sequence

Every call below is a real route in `backend/main.go`. Run it after the API is
up; `POST .../api/tx/*` needs the disbursement batch loaded first.

```sh
curl -s -X POST localhost:8080/api/seed/load -H 'X-Dev-User: admin'   # 215 records, 15 planted anomalies
curl -s -X POST localhost:8080/api/detect    -H 'X-Dev-User: admin'   # 15 findings, rules R1-R4
curl -s -X POST localhost:8080/api/tx/seed   -H 'X-Dev-User: admin'   # dummy partner-bank feed, 215 recipients
curl -s -X POST localhost:8080/api/tx/detect -H 'X-Dev-User: admin'   # spending findings, rules T1-T5
```

`GET /api/eval` and `GET /api/tx/eval` score the current batch against the
labelled synthetic seed (seed 42): precision@10, recall, false-positive rate and
the measured detection runtime. The labels are synthetic, not real KIP-K data.

`sh run-local.sh` does the same thing in one shot: Postgres, API on :8080, a
production build of the SPA served by `vite preview` on :8088. It expects an
already built `backend/backend` binary, so run `cd backend && go build -o backend .`
once first.

## What it flags

| Path | Rule | Signal |
|---|---|---|
| Disbursement | R1 | one NIK or NIM claimed by more than one student |
| Disbursement | R2 | more than one student on one account number |
| Disbursement | R3 | account holder name does not match the SK name (Jaro-Winkler below 0.85) |
| Disbursement | R4 | academic status is not `aktif` when the funds land |
| Spending | T1 | any transaction at a gambling merchant |
| Spending | T2 | 60% or more of the disbursement leaves within the first 7 days |
| Spending | T3 | 45% or more of the disbursement withdrawn as cash |
| Spending | T4 | a luxury purchase of Rp 1.500.000 or more |
| Spending | T5 | 3 or more transfers to the same counterparty (calo pattern) |

Findings carry a risk score, a band, and per-rule evidence in JSON. Status moves
from `baru` through an explicit lifecycle (`salah_deteksi` included), and every
transition writes an event row so a false positive is recorded, not hidden.

## Tech stack

| Layer | Choice |
|---|---|
| Frontend | Vite 6 + Preact 10, CSS variables from the design tokens |
| Backend | Go 1.26, `net/http` mux, GORM |
| Data | Postgres 16 (local Docker or Supabase pooler) |
| Auth | Supabase JWT (`SUPABASE_JWT_SECRET`), dev fallback on `X-Dev-User` |
| Chain | Solana batch anchor (`SOLANA_KEYPAIR`, optional) |

## Notes and gotchas

1. **RLS does not protect you through GORM.** Go connects with a privileged
   role, so Postgres row-level security is bypassed for every query the backend
   makes. Authorization must be enforced in Go: see `handlers.go`, where
   `owner_id` is always filtered from the verified subject. Keep RLS enabled
   anyway as defence for anything that talks to Supabase directly (anon key).
2. **Supabase's transaction pooler (port 6543) does not support prepared
   statements.** The GORM/pgx config here sets `PreferSimpleProtocol: true` so it
   works on the pooler. Use the direct connection (5432) for migrations.
3. **Don't add CORS if you use the Vite proxy.** `/api` is proxied to Go, so the
   browser stays same-origin. CORS headers exist only for the case where you
   call the API directly from another origin.

## Design and docs

`DESIGN.md` and `tokens.css` at the repo root come from OpenDesign's
`linear-app` design system. `frontend/src/styles/tokens.css` is a copy; the app
CSS consumes those variables, with a few literal colors still left in
`kiphone.css`. Keep the deck, the landing page and this app pointed at the same
file and they will match on stage.

- [PRD.md](PRD.md): problem, non-goals, rules and API contract
- [DESIGN.md](DESIGN.md): tokens, layout and component rules
- [BUGS.md](BUGS.md): QA pass over the backend, rules, seeder and API
- [DEMO.md](DEMO.md): 60 second stage script, with fallbacks
- [`design/kiphone-pitch.html`](design/kiphone-pitch.html): 10 slide deck with
  keyboard navigation, exported as
  [`design/kiphone-pitch.pdf`](design/kiphone-pitch.pdf)
- [`design/PITCH-SCRIPT.md`](design/PITCH-SCRIPT.md): speaker script and Q&A

## Where to change things

| Piece | Where |
|---|---|
| Legacy `/api/items` demo model | `backend/main.go` -> `Item` struct |
| Endpoints | `backend/main.go`, `backend/handlers*.go` |
| Detection rules | `backend/anomaly.go`, `backend/txn_engine.go` |
| Seed data | `backend/seed.go`, `backend/txn_seed.go` |
| Screens | `frontend/src/app.jsx`, `frontend/src/components/` |
| Auth UI | `frontend/src/lib/supabase.js` (client is ready, no screen yet) |

## License

MIT. See [LICENSE](LICENSE).
