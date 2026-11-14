import { useEffect, useState } from 'preact/hooks'
import {
  getStats,
  getFindings,
  getEval,
  runDetect,
  loadSeed,
  formatNumber,
  formatPct,
} from '../lib/integra.js'
import { Findings } from './Findings.jsx'
import { EvidenceDrawer } from '../components/EvidenceDrawer.jsx'
import { ChainPanel } from '../components/ChainPanel.jsx'

const CARD_TONES = {
  'Total Record': 'neutral',
  Temuan: 'accent',
  'Prioritas Tinggi': 'danger',
  'Salah Deteksi': 'danger',
}

export function Dashboard() {
  const [stats, setStats] = useState(null)
  const [findings, setFindings] = useState(null)
  const [evalData, setEvalData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [detectBusy, setDetectBusy] = useState(false)
  const [detectNote, setDetectNote] = useState('')
  const [detectError, setDetectError] = useState('')
  const [seedBusy, setSeedBusy] = useState(false)
  const [seedNote, setSeedNote] = useState('')
  const [filter, setFilter] = useState('semua')
  const [selectedId, setSelectedId] = useState(null)

  async function loadAll() {
    setError('')
    const [s, f, e] = await Promise.allSettled([getStats(), getFindings(), getEval()])
    const problems = []

    if (s.status === 'fulfilled') setStats(s.value)
    else problems.push(`statistik: ${s.reason.message}`)

    if (f.status === 'fulfilled') setFindings(f.value.findings ?? [])
    else problems.push(`temuan: ${f.reason.message}`)

    if (e.status === 'fulfilled') setEvalData(e.value)
    else problems.push(`evaluasi: ${e.reason.message}`)

    setError(problems.join(' · '))
    setLoading(false)
  }

  useEffect(() => {
    loadAll()
  }, [])

  async function handleDetect() {
    setDetectBusy(true)
    setDetectError('')
    setDetectNote('')
    try {
      const result = await runDetect()
      setSelectedId(null)
      await loadAll()
      setDetectNote(`${formatNumber(result.findings_created)} temuan · ${result.runtime_ms} ms`)
    } catch (err) {
      setDetectError(err.message)
    } finally {
      setDetectBusy(false)
    }
  }

  async function handleSeed() {
    setSeedBusy(true)
    setDetectError('')
    setDetectNote('')
    setSeedNote('')
    try {
      await loadSeed()
      setSelectedId(null)
      setFilter('semua')
      await loadAll()
      setSeedNote('Data contoh dimuat ulang.')
    } catch (err) {
      setDetectError(err.message)
    } finally {
      setSeedBusy(false)
    }
  }

  const cards = stats
    ? [
        ['Total Record', stats.records],
        ['Temuan', stats.findings],
        ['Prioritas Tinggi', stats.prioritas_tinggi],
        ['Salah Deteksi', stats.salah_deteksi],
      ]
    : []

  return (
    <main class="shell">
      <header class="masthead dashboard-head">
        <div class="fade-up">
          <p class="eyebrow">Integra · KIP-K</p>
          <h1>Integritas penyaluran KIP-K</h1>
          <p class="muted">Temuan ber-bukti. Keputusan tetap di tangan manusia.</p>
        </div>
        <button class="ghost" onClick={handleSeed} disabled={seedBusy}>
          {seedBusy ? 'Memuat…' : 'Muat ulang data contoh'}
        </button>
      </header>

      {seedNote && <p class="caption seed-note">{seedNote}</p>}
      {error && (
        <div class="banner error-banner" role="alert">
          Sebagian data gagal dimuat: {error}
        </div>
      )}

      {loading ? (
        <p class="muted empty">Memuat data…</p>
      ) : (
        stats && (
          <section class="stats-grid" aria-label="Ringkasan">
            {cards.map(([label, value], i) => (
              <article
                class={`stat-card tone-${CARD_TONES[label]} fade-up`}
                key={label}
                style={{ animationDelay: (i * 60) + 'ms' }}
              >
                <span class="stat-label">{label}</span>
                <strong class="stat-value">{formatNumber(value)}</strong>
              </article>
            ))}
          </section>
        )
      )}

      <section class="action-row">
        <button class="btn" onClick={handleDetect} disabled={detectBusy || loading}>
          {detectBusy ? 'Menjalankan…' : 'Jalankan Deteksi'}
        </button>
        {detectNote && <span class="caption" aria-live="polite">{detectNote}</span>}
        {detectError && <span class="error" role="alert">{detectError}</span>}
      </section>

      {evalData && (
        <section class="eval-strip" aria-label="Evaluasi model">
          <div class="eval-item">
            <span class="eval-label">Precision@10</span>
            <strong class="eval-value">{formatPct(evalData.precision_at_10)}</strong>
          </div>
          <div class="eval-item">
            <span class="eval-label">Recall</span>
            <strong class="eval-value">{formatPct(evalData.recall)}</strong>
          </div>
          <div class="eval-item">
            <span class="eval-label">False Positive Rate</span>
            <strong class="eval-value">{formatPct(evalData.false_positive_rate)}</strong>
          </div>
          <p class="eval-method">Metodologi: data sintetis berlabel (seed 42)</p>
        </section>
      )}

      {!loading && findings && (
        <Findings
          findings={findings}
          filter={filter}
          onFilter={setFilter}
          onSelect={setSelectedId}
        />
      )}

      {selectedId != null && (
        <EvidenceDrawer
          findingId={selectedId}
          onClose={() => setSelectedId(null)}
          onChanged={loadAll}
        />
      )}
      <ChainPanel />
    </main>
  )
}
