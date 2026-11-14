import { useEffect, useState } from 'preact/hooks'
import { getBatch, anchorBatch } from '../lib/integra.js'

// Batch receipt + Solana anchor. The hash is computed offline by the backend,
// so this screen is fully useful without a network or a funded wallet.
export function ChainPanel() {
  const [batch, setBatch] = useState(null)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const [note, setNote] = useState('')

  async function load() {
    try {
      setError('')
      setBatch(await getBatch())
    } catch (e) {
      setError(e.message)
    }
  }

  useEffect(() => { load() }, [])

  async function anchor() {
    setBusy(true)
    setNote('')
    try {
      const res = await anchorBatch()
      setNote('Ter-anchor: ' + res.signature.slice(0, 20) + '…')
      await load()
    } catch (e) {
      setNote(e.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <section class="chain-section">
      <div class="section-head">
        <div>
          <p class="eyebrow">Bukti batch</p>
          <h2 class="chain-title">Satu hash untuk seluruh batch pencairan.</h2>
          <p class="muted chain-lead">
            Hash dihitung dari setiap baris penyaluran, terurut dan deterministik.
            Tidak ada data pribadi yang masuk ke chain — hanya sidik jari batch.
          </p>
        </div>
      </div>

      {error && <p class="banner error-banner">{error}</p>}

      {!batch ? (
        <p class="muted empty">Memuat bukti batch…</p>
      ) : (
        <>
          <dl class="receipt">
            <div class="receipt-row">
              <dt>Hash batch (sha256)</dt>
              <dd class="mono hash">{batch.hash}</dd>
            </div>
            <div class="receipt-row">
              <dt>Jumlah record</dt>
              <dd>{batch.record_n}</dd>
            </div>
            <div class="receipt-row">
              <dt>Jaringan</dt>
              <dd>Solana {batch.cluster}</dd>
            </div>
            <div class="receipt-row">
              <dt>Status anchor</dt>
              <dd>
                {batch.signature ? (
                  <span class="pill status-ok">ter-anchor</span>
                ) : batch.anchor_ready ? (
                  <span class="pill status-muted">siap di-anchor</span>
                ) : (
                  <span class="pill status-muted">siap di-anchor (wallet belum diset)</span>
                )}
              </dd>
            </div>
            {batch.signature && (
              <div class="receipt-row">
                <dt>Transaksi</dt>
                <dd>
                  <a class="explorer" href={batch.explorer} target="_blank" rel="noreferrer">
                    {batch.signature.slice(0, 24)}… ↗
                  </a>
                </dd>
              </div>
            )}
          </dl>

          <div class="action-row chain-actions">
            <button class="btn" onClick={anchor} disabled={busy || !!batch.signature}>
              {busy ? 'Mengirim…' : batch.signature ? 'Sudah ter-anchor' : 'Anchor ke Solana devnet'}
            </button>
            <button class="ghost btn-sm" onClick={load} disabled={busy}>Muat ulang</button>
            {note && <span class="caption">{note}</span>}
          </div>

          <p class="caption chain-foot">
            Anchor adalah satu transaksi Memo — tanpa program kustom, tanpa biaya
            per record. Sengaja manual: jejak audit tidak boleh memblokir pencairan.
          </p>
        </>
      )}
    </section>
  )
}
