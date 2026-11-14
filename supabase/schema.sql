-- KIPhone schema for Supabase (mirrors backend/models_domain.go + batch.go).
-- GORM's AutoMigrate produces the same table names, so the Go code needs no change.

create table if not exists students (
  id                bigserial primary key,
  nim               text,
  nik               text,
  nama              text,
  prodi             text,
  status_akademik   text,
  last_active_at    timestamptz
);
create index if not exists idx_students_nim on students (nim);
create index if not exists idx_students_nik on students (nik);

create table if not exists accounts (
  id              bigserial primary key,
  student_id      bigint,
  nomor_rekening  text,
  nama_pemilik    text,
  created_at      timestamptz
);
create index if not exists idx_accounts_student_id on accounts (student_id);
create index if not exists idx_accounts_nomor_rekening on accounts (nomor_rekening);

create table if not exists disbursements (
  id            bigserial primary key,
  student_id    bigint,
  account_id    bigint,
  sk_nomor      text,
  periode       text,
  nominal       bigint,
  tanggal_cair  timestamptz
);
create index if not exists idx_disbursements_student_id on disbursements (student_id);

create table if not exists findings (
  id               bigserial primary key,
  disbursement_id  bigint,
  rule             text,
  risk_score       double precision,
  band             text,
  explanation      text,
  evidence         jsonb,
  status           text,
  created_at       timestamptz
);
create index if not exists idx_findings_disbursement_id on findings (disbursement_id);

create table if not exists finding_events (
  id           bigserial primary key,
  finding_id   bigint,
  from_status  text,
  to_status    text,
  actor        text,
  note         text,
  at           timestamptz
);
create index if not exists idx_finding_events_finding_id on finding_events (finding_id);

create table if not exists batches (
  id           bigserial primary key,
  periode      text,
  hash         text,
  record_n     integer,
  signature    text,
  cluster      text,
  anchored_at  timestamptz
);
create index if not exists idx_batches_periode on batches (periode);

-- Row level security: the Go backend connects with a privileged role and
-- enforces ownership itself (see PRD §4.4), but keep RLS on so the anon key
-- cannot read raw disbursement data from the browser.
alter table students        enable row level security;
alter table accounts        enable row level security;
alter table disbursements   enable row level security;
alter table findings        enable row level security;
alter table finding_events  enable row level security;
alter table batches         enable row level security;
