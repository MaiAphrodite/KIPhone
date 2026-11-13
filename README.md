# Hackstack — Go + GORM + Supabase + Vite/Preact

A working vertical slice you can rename into whatever the theme turns out to be.
Nothing here is theme-specific except the `Item` model.

```
Preact (Vite, :5173)  ──/api proxy──▶  Go API (:8080)  ──GORM──▶  Postgres (Supabase)
        │                                                           
        └──supabase-js (auth only: sign-in → JWT)                   
```

The frontend never talks to Postgres. Supabase is used for auth (and any
storage you add); all data access goes through Go so business rules live in one
place.

## Run it

```sh
# 1. database (local, or skip and point DATABASE_URL at Supabase)
cd hackstack && docker compose up -d

# 2. api
cd backend
cp .env.example .env
DATABASE_URL="postgresql://postgres:postgres@127.0.0.1:5432/hackstack" go run .

# 3. web
cd frontend
npm install
npm run dev          # http://localhost:5173
```

With `SUPABASE_JWT_SECRET` unset the API runs in dev mode and trusts the
`X-Dev-User` header, so the UI works before auth is wired.

## Three gotchas that bite this exact stack

1. **RLS does not protect you through GORM.** Go connects with a privileged
   role, so Postgres row-level security is bypassed for every query the backend
   makes. Authorization must be enforced in Go — see `handlers.go`, where
   `owner_id` is always filtered from the verified subject. Keep RLS enabled
   anyway as defence for anything that talks to Supabase directly (anon key).
2. **Supabase's transaction pooler (port 6543) does not support prepared
   statements.** The GORM/pgx config here sets `PreferSimpleProtocol: true` so it
   works on the pooler. Use the direct connection (5432) for migrations.
3. **Don't add CORS if you use the Vite proxy.** `/api` is proxied to Go, so the
   browser stays same-origin. CORS headers exist only for the case where you
   call the API directly from another origin.

## The design contract

`DESIGN.md` and `tokens.css` at the repo root come from OpenDesign's
`linear-app` design system. `frontend/src/styles/tokens.css` is a copy; the app
CSS consumes variables and never hardcodes a color. Keep the deck, the landing
page and this app pointed at the same file and they will match on stage.

## What to change for the theme

| Piece | Where |
|---|---|
| Domain model | `backend/main.go` → `Item` struct |
| Endpoints | `backend/handlers.go` |
| Screens | `frontend/src/app.jsx` |
| Auth UI | `frontend/src/lib/supabase.js` (client is ready, no screen yet) |
