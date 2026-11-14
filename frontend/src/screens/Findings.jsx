import { BAND, STATUS, rulesOf } from '../lib/integra.js'

const FILTERS = [
  ['semua', 'Semua'],
  ['prioritas', 'Prioritas'],
  ['verifikasi', 'Perlu verifikasi'],
  ['salah_deteksi', 'Salah deteksi'],
]

function matches(finding, filter) {
  if (filter === 'prioritas') return finding.band === 'prioritas_tinggi'
  if (filter === 'verifikasi') return finding.band === 'perlu_verifikasi'
  if (filter === 'salah_deteksi') return finding.status === 'salah_deteksi'
  return true
}

const EMPTY_COPY = {
  semua: 'Belum ada temuan. Jalankan deteksi untuk memindai batch.',
  prioritas: 'Tidak ada temuan prioritas tinggi.',
  verifikasi: 'Tidak ada temuan yang perlu verifikasi.',
  salah_deteksi: 'Belum ada temuan yang ditandai salah deteksi.',
}

function RuleBadges({ rule }) {
  return (
    <div class="rule-cell">
      {rulesOf(rule).map((r) => (
        <span class="pill rule" key={r}>{r}</span>
      ))}
    </div>
  )
}

function FindingRow({ finding, onSelect, index }) {
  const band = BAND[finding.band] ?? { label: finding.band, tone: 'ok' }
  const status = STATUS[finding.status] ?? { label: finding.status, tone: 'muted' }
  const score = Number(finding.risk_score) || 0

  return (
    <tr class="fade-up" style={{ animationDelay: (index * 20) + 'ms' }}>
      <td class="mono">{finding.nim}</td>
      <td class="nama-cell">{finding.nama}</td>
      <td><RuleBadges rule={finding.rule} /></td>
      <td>
        <div class="risk-cell">
          <span class="risk-track" aria-hidden="true">
            <span class={`risk-fill ${band.tone}`} style={{ width: `${Math.round(score * 100)}%` }} />
          </span>
          <span class="risk-score">{score.toFixed(2)}</span>
        </div>
      </td>
      <td><span class={`band ${band.tone}`}>{band.label}</span></td>
      <td><span class={`pill status-${status.tone}`}>{status.label}</span></td>
      <td>
        <button class="ghost btn-sm" onClick={() => onSelect(finding.id)}>Lihat bukti</button>
      </td>
    </tr>
  )
}

export function Findings({ findings, filter, onFilter, onSelect }) {
  const rows = findings
    .filter((finding) => matches(finding, filter))
    .sort((a, b) => b.risk_score - a.risk_score)

  return (
    <section class="findings-section" aria-label="Daftar temuan">
      <div class="section-head">
        <div>
          <h2>Daftar temuan</h2>
          <p class="muted small">Diurutkan dari skor risiko tertinggi.</p>
        </div>
        <div class="chips" role="tablist" aria-label="Filter temuan">
          {FILTERS.map(([value, label]) => (
            <button
              key={value}
              class={`chip${filter === value ? ' active' : ''}`}
              role="tab"
              aria-selected={filter === value}
              onClick={() => onFilter(value)}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      {rows.length === 0 ? (
        <p class="muted empty">{EMPTY_COPY[filter] ?? EMPTY_COPY.semua}</p>
      ) : (
        <div class="table-wrap">
          <table class="findings">
            <thead>
              <tr>
                <th>NIM</th>
                <th>Nama</th>
                <th>Rule</th>
                <th>Risiko</th>
                <th>Band</th>
                <th>Status</th>
                <th><span class="sr-only">Aksi</span></th>
              </tr>
            </thead>
            <tbody>
              {rows.map((finding, i) => (
                <FindingRow key={finding.id} finding={finding} onSelect={onSelect} index={i} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
