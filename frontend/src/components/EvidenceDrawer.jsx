import { useEffect, useState } from 'preact/hooks'
import {
  getFinding,
  setFindingStatus,
  parseEvidence,
  rulesOf,
  formatRupiah,
  formatDate,
  formatDateTime,
  BAND,
  STATUS,
  TRANSITIONS,
  RULE_NAMES,
  EVIDENCE_FIELDS,
} from '../lib/integra.js'

function fieldValue(key, value) {
  if (Array.isArray(value)) return value.join(', ')
  if (key === 'tanggal_cair' || key === 'last_active_at') return formatDate(value)
  if (value && typeof value === 'object') return JSON.stringify(value)
  return String(value ?? '-')
}

function EvidenceBlocks({ rule, blocks }) {
  const pairs = rulesOf(rule).map((r) => [r, blocks[r]]).filter(([, ev]) => ev)
  return pairs.map(([r, ev]) => {
    const fields = EVIDENCE_FIELDS[r] ?? Object.keys(ev).map((key) => [key, key.replace(/_/g, ' ')])
    return (
      <div class="evidence-block" key={r}>
        <h5>{r} · {RULE_NAMES[r] ?? 'Bukti'}</h5>
        <dl class="evidence-dl">
          {fields.map(([key, label]) => (
            <div class="evidence-row" key={key}>
              <dt>{label}</dt>
              <dd>{fieldValue(key, ev[key])}</dd>
            </div>
          ))}
        </dl>
      </div>
    )
  })
}

export function EvidenceDrawer({ findingId, onClose, onChanged }) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [pending, setPending] = useState(null)
  const [note, setNote] = useState('')
  const [busy, setBusy] = useState(false)
  const [actionError, setActionError] = useState('')

  async function load() {
    setError('')
    try {
      setData(await getFinding(findingId))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    setLoading(true)
    load()
  }, [findingId])

  useEffect(() => {
    const onKey = (e) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  async function submit(toStatus) {
    setBusy(true)
    setActionError('')
    try {
      await setFindingStatus(findingId, toStatus, note.trim())
      setPending(null)
      setNote('')
      await load()
      onChanged?.()
    } catch (err) {
      setActionError(err.message)
    } finally {
      setBusy(false)
    }
  }

  const finding = data?.finding
  const events = data?.events ?? []
  const band = finding ? BAND[finding.band] ?? { label: finding.band, tone: 'ok' } : null
  const status = finding ? STATUS[finding.status] ?? { label: finding.status, tone: 'muted' } : null
  const transitions = finding ? TRANSITIONS[finding.status] ?? [] : []
  const pendingAction = transitions.find((t) => t.to === pending)

  return (
    <div class="drawer-backdrop" onClick={onClose} role="presentation">
      <aside
        class="drawer"
        role="dialog"
        aria-modal="true"
        aria-label="Panel bukti temuan"
        onClick={(e) => e.stopPropagation()}
      >
        <div class="drawer-head">
          <div>
            <h3 class="drawer-title">Bukti temuan</h3>
            <p class="drawer-meta">Panel bukti dan riwayat keputusan.</p>
          </div>
          <button class="ghost btn-sm" onClick={onClose} aria-label="Tutup panel bukti">Tutup</button>
        </div>

        {loading ? (
          <p class="muted empty">Memuat bukti…</p>
        ) : error ? (
          <div class="banner error-banner" role="alert">Gagal memuat bukti: {error}</div>
        ) : finding ? (
          <>
            <section class="drawer-section">
              <h2 class="drawer-title">{finding.nama}</h2>
              <p class="drawer-meta">
                NIM {finding.nim} · {finding.prodi} · Periode {finding.periode} · {formatRupiah(finding.nominal)}
              </p>
              <div class="drawer-badges">
                {rulesOf(finding.rule).map((r) => (
                  <span class="pill rule" key={r}>{r}</span>
                ))}
                <span class={`band ${band.tone}`}>{band.label}</span>
                <span class={`pill status-${status.tone}`}>{status.label}</span>
                <span class="risk-score">risiko {Number(finding.risk_score).toFixed(2)}</span>
              </div>
            </section>

            <section class="drawer-section">
              <h4>Bukti per aturan</h4>
              <EvidenceBlocks rule={finding.rule} blocks={parseEvidence(finding.evidence)} />
            </section>

            <section class="drawer-section">
              <h4>Penjelasan</h4>
              <blockquote class="explanation">{finding.explanation}</blockquote>
            </section>

            {transitions.length > 0 && (
              <section class="drawer-section">
                <h4>Tindakan</h4>
                {pending ? (
                  <div class="note-box">
                    <label class="field-label" for="audit-note">
                      Catatan opsional untuk "{pendingAction?.label ?? pending}"
                    </label>
                    <textarea
                      id="audit-note"
                      class="note-input"
                      rows={2}
                      value={note}
                      placeholder="Catatan untuk jejak audit…"
                      onInput={(e) => setNote(e.currentTarget.value)}
                    />
                    <div class="lifecycle-actions">
                      <button
                        class={pending === 'salah_deteksi' ? 'btn btn-danger' : 'btn'}
                        disabled={busy}
                        onClick={() => submit(pending)}
                      >
                        {busy ? 'Menyimpan…' : 'Konfirmasi'}
                      </button>
                      <button class="ghost" disabled={busy} onClick={() => setPending(null)}>
                        Batal
                      </button>
                    </div>
                  </div>
                ) : (
                  <div class="lifecycle-actions">
                    {transitions.map((t) => (
                      <button
                        class={t.tone === 'danger' ? 'btn btn-danger' : 'btn'}
                        key={t.to}
                        onClick={() => {
                          setPending(t.to)
                          setNote('')
                          setActionError('')
                        }}
                      >
                        {t.label}
                      </button>
                    ))}
                  </div>
                )}
                {actionError && <p class="error">{actionError}</p>}
              </section>
            )}

            <section class="drawer-section">
              <h4>Riwayat status</h4>
              {events.length === 0 ? (
                <p class="muted small">Belum ada riwayat.</p>
              ) : (
                <ul class="event-list">
                  {events.map((ev) => (
                    <li class="event-row" key={ev.id}>
                      {ev.from_status || 'awal'} → {ev.to_status} · {ev.actor}
                      {ev.note ? ` · ${ev.note}` : ''} · {formatDateTime(ev.at)}
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </>
        ) : null}
      </aside>
    </div>
  )
}
