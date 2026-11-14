import { useEffect, useState } from 'preact/hooks'

// Spending view over the partner-bank transaction feed (dummy data in this
// build). Same contract as the disbursement view: evidence, no automatic
// action, human decides.
const BASE_HEADERS = { 'Content-Type': 'application/json', 'X-Dev-User': 'admin' }

async function api(path, options = {}) {
  const res = await fetch(`/api${path}`, { ...options, headers: { ...BASE_HEADERS, ...(options.headers || {}) } })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

const BAND = {
  prioritas_tinggi: { label: 'prioritas tinggi', tone: 'danger' },
  perlu_verifikasi: { label: 'perlu verifikasi', tone: 'warn' },
  rendah: { label: 'risiko rendah', tone: 'ok' },
}
const RULES = {
  T1: 'Transaksi ke merchant taruhan',
  T2: 'Dana habis dalam 7 hari',
  T3: 'Tarik tunai terkonsentrasi',
  T4: 'Pembelian barang mewah',
  T5: 'Transfer berulang ke penerima sama (calo/joki)',
}

function rupiah(v) {
  return 'Rp ' + Number(v || 0).toLocaleString('id-ID')
}

function Evidence({ finding, onClose, onChanged }) {
  const [detail, setDetail] = useState(null)
  const [note, setNote] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')

  useEffect(() => {
    api(`/tx/findings/${finding.id}`).then(setDetail).catch((e) => setErr(e.message))
  }, [finding.id])

  async function move(to) {
    setBusy(true)
    try {
      await api(`/tx/findings/${finding.id}/status`, { method: 'POST', body: JSON.stringify({ to_status: to, note }) })
      onChanged()
    } catch (e) {
      setErr(e.message)
    } finally {
      setBusy(false)
    }
  }

  let evidence = {}
  try { evidence = JSON.parse(finding.evidence || '{}') } catch { /* ignore */ }

  return (
    <div class="drawer-backdrop" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <aside class="drawer">
        <div class="drawer-head">
          <div>
            <p class="eyebrow">Penggunaan dana</p>
            <h2 class="drawer-title">{finding.nama}</h2>
            <p class="drawer-meta">{finding.nim} · {RULES[finding.rules.split(',')[0]] || finding.rules}</p>
          </div>
          <button class="ghost btn-sm" onClick={onClose}>Tutup</button>
        </div>

        <div class="drawer-badges">
          {finding.rules.split(',').map((r) => <span class="pill rule" key={r}>{r}</span>)}
          <span class={`band ${(BAND[finding.band] || {}).tone || 'ok'}`}>{(BAND[finding.band] || {}).label || finding.band}</span>
          <span class="risk-score">{finding.risk_score.toFixed(2)}</span>
        </div>

        {err && <p class="banner error-banner">{err}</p>}

        <section class="drawer-section">
          <h3>Mengapa ditandai</h3>
          <p class="explanation">{finding.explanation}</p>
        </section>

        <section class="drawer-section">
          <h3>Bukti per pola</h3>
          {Object.entries(evidence).map(([rule, ev]) => (
            <div class="evidence-block" key={rule}>
              <span class="pill rule">{rule}</span> <span class="caption">{RULES[rule]}</span>
              <dl class="evidence-dl">
                {Object.entries(ev).map(([k, v]) => (
                  <div class="evidence-row" key={k}>
                    <span class="field-label">{k}</span>
                    <span>{Array.isArray(v) ? v.join(', ') : String(v)}</span>
                  </div>
                ))}
              </dl>
            </div>
          ))}
        </section>

        <section class="drawer-section">
          <h3>Transaksi terkait</h3>
          {!detail ? (
            <p class="muted empty">Memuat…</p>
          ) : (
            <div class="table-wrap">
              <table class="findings">
                <thead><tr><th>Tanggal</th><th>Merchant</th><th>Kategori</th><th>Nominal</th></tr></thead>
                <tbody>
                  {detail.transactions.filter((t) => t.kategori === 'taruhan' || t.kategori === 'mewah' || t.kategori === 'tunai').slice(0, 12).map((t) => (
                    <tr key={t.id}>
                      <td class="mono">{String(t.tanggal).slice(0, 10)}</td>
                      <td>{t.merchant}</td>
                      <td class="caption">{t.kategori}</td>
                      <td class="mono">{rupiah(t.nominal)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>

        <section class="drawer-section">
          <h3>Keputusan</h3>
          <div class="note-box">
            <textarea class="note-input" rows="2" placeholder="Catatan verifikator (opsional)"
              value={note} onInput={(e) => setNote(e.currentTarget.value)} />
          </div>
          <div class="lifecycle-actions">
            <button class="btn" disabled={busy} onClick={() => move('diverifikasi')}>Verifikasi</button>
            <button class="btn btn-danger" disabled={busy} onClick={() => move('salah_deteksi')}>Tandai Salah Deteksi</button>
            <button class="ghost" disabled={busy} onClick={() => move('selesai')}>Selesaikan</button>
          </div>
          <p class="caption">
            Tidak ada tindakan otomatis. Dana mahasiswa tidak pernah diblokir oleh model.
          </p>
        </section>
      </aside>
    </div>
  )
}

export function Spending({ onBack }) {
  const [stats, setStats] = useState(null)
  const [findings, setFindings] = useState(null)
  const [evalData, setEvalData] = useState(null)
  const [selected, setSelected] = useState(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')

  async function load() {
    try {
      setErr('')
      const [s, f, e] = await Promise.all([api('/tx/stats'), api('/tx/findings'), api('/tx/eval')])
      setStats(s); setFindings(f.findings); setEvalData(e)
    } catch (e) {
      setErr(e.message)
    }
  }
  useEffect(() => { load() }, [])

  async function runDetect() {
    setBusy(true)
    try {
      await api('/tx/detect', { method: 'POST' })
      await load()
    } catch (e) {
      setErr(e.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <main class="shell">
      <header class="masthead dashboard-head">
        <div>
          <p class="eyebrow">Penggunaan dana · feed bank (dummy)</p>
          <h1>Apakah dana dipakai sebagaimana mestinya?</h1>
          <p class="muted">Pola dari transaksi penerima, dari kerja sama bank — bukan kampus mengintip rekening.</p>
        </div>
        <button class="ghost btn-sm" onClick={onBack}>← Kembali ke penyaluran</button>
      </header>

      {err && <div class="banner error-banner">{err}</div>}

      <section class="stats-grid">
        <article class="stat-card tone-neutral fade-up"><span class="stat-label">Transaksi diproses</span>
          <strong class="stat-value">{stats ? stats.transactions.toLocaleString('id-ID') : '—'}</strong></article>
        <article class="stat-card tone-accent fade-up" style="animation-delay:60ms"><span class="stat-label">Pola ditandai</span>
          <strong class="stat-value">{stats ? stats.findings : '—'}</strong></article>
        <article class="stat-card tone-danger fade-up" style="animation-delay:120ms"><span class="stat-label">Prioritas tinggi</span>
          <strong class="stat-value">{stats ? stats.prioritas_tinggi : '—'}</strong></article>
        <article class="stat-card tone-danger fade-up" style="animation-delay:180ms"><span class="stat-label">Salah deteksi</span>
          <strong class="stat-value">{stats ? stats.salah_deteksi : '—'}</strong></article>
      </section>

      <section class="action-row">
        <button class="btn" onClick={runDetect} disabled={busy}>{busy ? 'Menjalankan…' : 'Jalankan Deteksi Pola'}</button>
      </section>

      {evalData && (
        <section class="eval-strip">
          <div class="eval-item"><span class="eval-label">Precision@10</span><strong class="eval-value">{evalData.precision_at_10.toFixed(2)}</strong></div>
          <div class="eval-item"><span class="eval-label">Recall</span><strong class="eval-value">{evalData.recall.toFixed(2)}</strong></div>
          <div class="eval-item"><span class="eval-label">False positive</span><strong class="eval-value">{Math.round(evalData.false_positive_rate * 100)}%</strong></div>
          <span class="eval-method">{evalData.methodology}</span>
        </section>
      )}

      <section class="findings-section">
        <div class="section-head"><h2>Temuan</h2><span class="caption">Lima pola: T1 taruhan · T2 dana habis cepat · T3 tunai terkonsentrasi · T4 barang mewah · T5 transfer berulang</span></div>
        {!findings ? <p class="muted empty">Memuat…</p> : findings.length === 0 ? (
          <p class="muted empty">Belum ada pola. Jalankan deteksi.</p>
        ) : (
          <div class="table-wrap">
            <table class="findings">
              <thead>
                <tr><th>NIM</th><th>Nama</th><th>Pola</th><th>Risiko</th><th>Status</th><th /></tr>
              </thead>
              <tbody>
                {findings.map((f, i) => {
                  const band = BAND[f.band] || { label: f.band, tone: 'ok' }
                  return (
                    <tr key={f.id} class="fade-up" style={{ animationDelay: `${i * 20}ms` }}>
                      <td class="mono">{f.nim}</td>
                      <td class="nama-cell">{f.nama}</td>
                      <td class="rule-cell">{f.rules.split(',').map((r) => <span class="pill rule" key={r}>{r}</span>)}</td>
                      <td class="risk-cell">
                        <div class="risk-track"><div class={`risk-fill ${band.tone}`} style={{ width: `${Math.round(f.risk_score * 100)}%` }} /></div>
                        <span class="risk-score">{f.risk_score.toFixed(2)}</span>
                      </td>
                      <td><span class={`pill status-${f.status === 'baru' ? 'muted' : f.status === 'salah_deteksi' ? 'danger' : f.status === 'selesai' ? 'ok' : 'accent'}`}>{f.status}</span></td>
                      <td><button class="ghost btn-sm" onClick={() => setSelected(f)}>Lihat bukti</button></td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {selected && <Evidence finding={selected} onClose={() => setSelected(null)} onChanged={load} />}
    </main>
  )
}
