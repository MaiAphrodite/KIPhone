import { accessToken } from './supabase.js'

const DEV_USER = 'admin'

// Static demo build: no backend. Reads a snapshot in /demo and keeps lifecycle
// changes in memory so the drawer still behaves like the live app.
const DEMO_MODE = import.meta.env.VITE_DEMO === '1'
const demoStatus = new Map()
const demoEvents = new Map()

function demoFile(path) {
  const clean = path.replace(/^\//, '').replace(/\//g, '-')
  return `/demo/${clean}.json`
}

async function demoRequest(path, options) {
  const method = (options.method || 'GET').toUpperCase()
  if (method === 'POST' && path === '/detect') return { findings_created: 15, by_rule: { R1: 4, R2: 6, R3: 6, R4: 3 }, runtime_ms: 13 }
  if (method === 'POST' && path === '/seed/load') return { students: 215, accounts: 215, disbursements: 215, planted: 15 }
  if (method === 'POST' && path === '/batch/anchor') throw new Error('Demo statis: anchor butuh backend + wallet devnet')
  const statusMatch = path.match(/^\/findings\/(\d+)\/status$/)
  if (method === 'POST' && statusMatch) {
    const id = Number(statusMatch[1])
    const body = JSON.parse(options.body || '{}')
    demoStatus.set(id, body.to_status)
    const list = demoEvents.get(id) || []
    list.push({ from_status: 'baru', to_status: body.to_status, actor: 'demo', note: body.note || '', at: new Date().toISOString() })
    demoEvents.set(id, list)
    return { status: body.to_status }
  }
  const detailMatch = path.match(/^\/findings\/(\d+)$/)
  const res = await fetch(demoFile(path))
  if (!res.ok) throw new Error('Data demo tidak tersedia')
  const data = await res.json()
  if (detailMatch) {
    const id = Number(detailMatch[1])
    if (demoStatus.has(id)) data.finding.status = demoStatus.get(id)
    data.events = [...(data.events || []), ...(demoEvents.get(id) || [])]
    return data
  }
  if (path === '/findings' && Array.isArray(data.findings)) {
    data.findings = data.findings.map((f) => demoStatus.has(f.id) ? { ...f, status: demoStatus.get(f.id) } : f)
  }
  return data
}

async function request(path, options = {}) {
  if (DEMO_MODE) return demoRequest(path, options)
  const token = await accessToken()
  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) }
  if (token) headers.Authorization = `Bearer ${token}`
  else headers['X-Dev-User'] = DEV_USER

  const res = await fetch(`/api${path}`, { ...options, headers })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.status === 204 ? null : res.json()
}

export const getStats = () => request('/stats')
export const getFindings = () => request('/findings')
export const getFinding = (id) => request(`/findings/${id}`)
export const getEval = () => request('/eval')
export const runDetect = () => request('/detect', { method: 'POST' })
export const loadSeed = () => request('/seed/load', { method: 'POST' })

export const setFindingStatus = (id, toStatus, note) =>
  request(`/findings/${id}/status`, {
    method: 'POST',
    body: JSON.stringify({ to_status: toStatus, note }),
  })

export const BAND = {
  prioritas_tinggi: { label: 'prioritas tinggi', tone: 'danger' },
  perlu_verifikasi: { label: 'perlu verifikasi', tone: 'warn' },
  rendah: { label: 'risiko rendah', tone: 'ok' },
}

export const STATUS = {
  baru: { label: 'Baru', tone: 'muted' },
  diverifikasi: { label: 'Diverifikasi', tone: 'accent' },
  salah_deteksi: { label: 'Salah Deteksi', tone: 'danger' },
  selesai: { label: 'Selesai', tone: 'ok' },
}

export const TRANSITIONS = {
  baru: [
    { to: 'diverifikasi', label: 'Verifikasi', tone: 'primary' },
    { to: 'salah_deteksi', label: 'Tandai Salah Deteksi', tone: 'danger' },
  ],
  diverifikasi: [
    { to: 'selesai', label: 'Selesaikan', tone: 'primary' },
    { to: 'salah_deteksi', label: 'Tandai Salah Deteksi', tone: 'danger' },
  ],
  salah_deteksi: [{ to: 'selesai', label: 'Selesaikan', tone: 'primary' }],
  selesai: [],
}

export const RULE_NAMES = {
  R1: 'Penerima ganda',
  R2: 'Rekening bersama',
  R3: 'Nama tidak cocok',
  R4: 'Nonaktif menerima',
}

export const EVIDENCE_FIELDS = {
  R1: [
    ['nik', 'NIK'],
    ['claimed_by', 'Diklaim oleh'],
  ],
  R2: [
    ['nomor_rekening', 'Nomor rekening'],
    ['dipakai_oleh', 'Dipakai oleh'],
    ['rekening_umur_hari', 'Umur rekening (hari)'],
  ],
  R3: [
    ['nama_sk', 'Nama di SK'],
    ['nama_rekening', 'Nama rekening'],
    ['nama_match_score', 'Skor kemiripan'],
    ['token_berbeda', 'Token berbeda'],
  ],
  R4: [
    ['status_akademik', 'Status akademik'],
    ['last_active_at', 'Terakhir aktif'],
    ['tanggal_cair', 'Tanggal cair'],
  ],
}

export function rulesOf(rule) {
  return (rule || '').split(',').map((r) => r.trim()).filter(Boolean)
}

export function parseEvidence(evidence) {
  if (!evidence) return {}
  if (typeof evidence === 'object') return evidence
  try {
    return JSON.parse(evidence)
  } catch {
    return {}
  }
}

const numberFmt = new Intl.NumberFormat('id-ID')
const dateFmt = new Intl.DateTimeFormat('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
const dateTimeFmt = new Intl.DateTimeFormat('id-ID', {
  day: '2-digit',
  month: 'short',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

function toDate(value) {
  if (!value) return null
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? null : d
}

export function formatNumber(value) {
  return numberFmt.format(Number(value) || 0)
}

export function formatRupiah(value) {
  return `Rp ${numberFmt.format(Number(value) || 0)}`
}

export function formatDate(value) {
  const d = toDate(value)
  return d ? dateFmt.format(d) : String(value ?? '-')
}

export function formatDateTime(value) {
  const d = toDate(value)
  return d ? dateTimeFmt.format(d) : String(value ?? '-')
}

export function formatPct(value) {
  if (value == null) return '-'
  return `${Math.round(Number(value) * 100)}%`
}

// ── batch receipt / chain ──────────────────────────────────────────────
export const getBatch = () => request('/batch')
export const anchorBatch = () => request('/batch/anchor', { method: 'POST' })
