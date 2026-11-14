import { useEffect, useState } from 'preact/hooks'
import { getStats, getBatch, getEval } from '../lib/integra.js'

// Public transparency view. Aggregates and the batch receipt only — this screen
// deliberately never renders a student, an account number, or an individual
// finding.
export function PublicView({ onBack }) {
  const [stats, setStats] = useState(null)
  const [batch, setBatch] = useState(null)
  const [evalData, setEvalData] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([getStats(), getBatch(), getEval()])
      .then(([s, b, e]) => { setStats(s); setBatch(b); setEvalData(e) })
      .catch((err) => setError(err.message))
  }, [])

  return (
    <main class="shell public">
      <header class="dashboard-head">
        <div>
          <p class="eyebrow">Halaman publik · transparansi penyaluran</p>
          <h1>Integritas penyaluran KIP-K, terbuka untuk diperiksa.</h1>
          <p class="muted chain-lead">
            Halaman ini mempublikasikan <strong>agregat</strong> dan
            <strong> sidik jari batch</strong> pencairan. Bukan daftar nama.
          </p>
        </div>
        <button class="ghost btn-sm" onClick={onBack}>Kembali</button>
      </header>

      {error && <p class="banner error-banner">{error}</p>}

      <section class="stats-grid" aria-label="Ringkasan publik">
        <article class="stat-card tone-neutral fade-up">
          <span class="stat-label">Record dipindai</span>
          <strong class="stat-value">{stats ? stats.records : '—'}</strong>
        </article>
        <article class="stat-card tone-accent fade-up" style="animation-delay:60ms">
          <span class="stat-label">Temuan</span>
          <strong class="stat-value">{stats ? stats.findings : '—'}</strong>
        </article>
        <article class="stat-card tone-danger fade-up" style="animation-delay:120ms">
          <span class="stat-label">Prioritas tinggi</span>
          <strong class="stat-value">{stats ? stats.prioritas_tinggi : '—'}</strong>
        </article>
        <article class="stat-card tone-danger fade-up" style="animation-delay:180ms">
          <span class="stat-label">Salah deteksi</span>
          <strong class="stat-value">{stats ? stats.salah_deteksi : '—'}</strong>
        </article>
      </section>

      <section class="drawer-section">
        <h3>Bukti batch</h3>
        <dl class="receipt">
          <div class="receipt-row">
            <dt>Hash sha256</dt>
            <dd class="mono hash">{batch ? batch.hash : '—'}</dd>
          </div>
          <div class="receipt-row">
            <dt>Jumlah record</dt>
            <dd>{batch ? batch.record_n : '—'}</dd>
          </div>
          {batch && batch.signature && (
            <div class="receipt-row">
              <dt>Transaksi Solana</dt>
              <dd>
                <a class="explorer" href={batch.explorer} target="_blank" rel="noreferrer">
                  {batch.signature.slice(0, 32)}… ↗
                </a>
              </dd>
            </div>
          )}
        </dl>
        <p class="caption chain-foot">
          Siapa pun dapat memverifikasi bahwa batch ini tidak diubah: hash dihitung
          ulang dari seluruh baris penyaluran, dan transaksi Memo di Solana devnet
          mencatat sidik jarinya secara permanen.
        </p>
      </section>

      <section class="drawer-section">
        <h3>Yang tidak dipublikasikan</h3>
        <ul class="privacy-list">
          <li>Tidak ada nama, NIM, NIK, atau nomor rekening mahasiswa.</li>
          <li>Tidak ada temuan per orang — daftar periksa hanya untuk verifikator internal.</li>
          <li>Tidak ada status "terindikasi" yang melekat pada individu, di mana pun.</li>
          <li>Chain hanya menyimpan hash batch, bukan data pribadi.</li>
        </ul>
        <p class="caption chain-foot">
          Model tidak memblokir apa pun. Keputusan tetap di tangan manusia, dan
          keputusan itu tercatat di sistem internal — bukan di halaman ini.
        </p>
      </section>

      {evalData && (
        <section class="eval-strip" aria-label="Metodologi">
          <div class="eval-item">
            <span class="eval-label">Precision@10</span>
            <strong class="eval-value">{evalData.precision_at_10.toFixed(2)}</strong>
          </div>
          <div class="eval-item">
            <span class="eval-label">Recall</span>
            <strong class="eval-value">{evalData.recall.toFixed(2)}</strong>
          </div>
          <div class="eval-item">
            <span class="eval-label">False positive</span>
            <strong class="eval-value">{Math.round(evalData.false_positive_rate * 100)}%</strong>
          </div>
          <span class="eval-method">
            Metodologi: data sintetis berlabel (seed 42). Bukan klaim akurasi data asli.
          </span>
        </section>
      )}
    </main>
  )
}
