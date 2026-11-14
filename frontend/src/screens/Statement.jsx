import { useEffect, useState } from 'preact/hooks'

// Bank-statement view over the dummy partner-bank feed: running balance,
// daily outflow chart, category breakdown, and the full mutation list.

const HEADERS = { 'Content-Type': 'application/json', 'X-Dev-User': 'admin' }
const api = async (path) => {
  const res = await fetch(`/api${path}`, { headers: HEADERS })
  if (!res.ok) {
    const b = await res.json().catch(() => ({}))
    throw new Error(b.error || `HTTP ${res.status}`)
  }
  return res.json()
}

const rupiah = (v) => 'Rp ' + Number(v || 0).toLocaleString('id-ID')
const short = (v) => {
  const n = Number(v || 0)
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1).replace('.0', '') + 'jt'
  if (n >= 1_000) return Math.round(n / 1_000) + 'rb'
  return String(n)
}
const CAT_LABEL = {
  makan: 'Makan', transport: 'Transport', pulsa: 'Pulsa', belanja: 'Belanja',
  hiburan: 'Hiburan', mewah: 'Barang mewah', taruhan: 'Taruhan',
  tunai: 'Tarik tunai', transfer: 'Transfer', masuk: 'Dana masuk', lainnya: 'Lainnya',
}
const CAT_TONE = {
  taruhan: 'danger', mewah: 'danger', tunai: 'warn', transfer: 'warn',
  masuk: 'ok', makan: '', transport: '', pulsa: '', belanja: '', hiburan: '',
}

export function Statement({ studentId, onBack }) {
  const [data, setData] = useState(null)
  const [err, setErr] = useState('')

  useEffect(() => {
    api(`/tx/students/${studentId}/statement`).then(setData).catch((e) => setErr(e.message))
  }, [studentId])

  if (err) {
    return (
      <main class="shell">
        <p class="banner error-banner">{err}</p>
        <button class="ghost btn-sm" onClick={onBack}>← Kembali</button>
      </main>
    )
  }
  if (!data) return <main class="shell"><p class="muted empty">Memuat mutasi…</p></main>

  const days = data.days || []
  const maxDay = Math.max(1, ...days.map((d) => d.keluar))
  const cats = data.by_category || []
  const maxCat = Math.max(1, ...cats.map((c) => c.total))

  return (
    <main class="shell">
      <header class="masthead dashboard-head">
        <div>
          <p class="eyebrow">Mutasi rekening · {data.source}</p>
          <h1>{data.student.nama}</h1>
          <p class="muted">
            {data.student.nim} · {data.student.prodi} · pencairan {data.disbursement.periode} {rupiah(data.disbursement.nominal)}
          </p>
        </div>
        <button class="ghost btn-sm" onClick={onBack}>← Kembali</button>
      </header>

      <section class="stats-grid">
        <article class="stat-card tone-neutral fade-up">
          <span class="stat-label">Transaksi</span>
          <strong class="stat-value">{data.rows.length - 1}</strong>
        </article>
        <article class="stat-card tone-accent fade-up" style="animation-delay:60ms">
          <span class="stat-label">Total keluar</span>
          <strong class="stat-value" style="font-size:26px">{rupiah(data.total_keluar)}</strong>
        </article>
        <article class="stat-card tone-danger fade-up" style="animation-delay:120ms">
          <span class="stat-label">Saldo akhir</span>
          <strong class="stat-value" style="font-size:26px">{rupiah(data.saldo_akhir)}</strong>
        </article>
        <article class="stat-card tone-accent fade-up" style="animation-delay:180ms">
          <span class="stat-label">Hari aktif</span>
          <strong class="stat-value">{days.length}</strong>
        </article>
      </section>

      <section class="drawer-section">
        <h3>Pengeluaran harian</h3>
        <div class="bars" role="img" aria-label="Grafik pengeluaran harian">
          {days.map((d) => (
            <div class="bar-col" key={d.tanggal} title={`${d.tanggal}: ${rupiah(d.keluar)}`}>
              <div class="bar" style={{ height: `${Math.max(2, Math.round((d.keluar / maxDay) * 100))}%` }} />
              <span class="bar-x">{d.tanggal.slice(8, 10)}</span>
            </div>
          ))}
        </div>
        <p class="caption">Tanggal 1–{days.length} setelah pencairan. Batang = pengeluaran hari itu.</p>
      </section>

      <section class="drawer-section">
        <h3>Ke mana uangnya pergi</h3>
        <div class="cat-list">
          {cats.map((c) => (
            <div class="cat-row" key={c.kategori}>
              <span class="cat-name">{CAT_LABEL[c.kategori] || c.kategori}</span>
              <div class="cat-track">
                <div class={`cat-fill ${CAT_TONE[c.kategori] || ''}`} style={{ width: `${Math.round((c.total / maxCat) * 100)}%` }} />
              </div>
              <span class="cat-val">{short(c.total)}</span>
            </div>
          ))}
        </div>
      </section>

      <section class="drawer-section">
        <h3>Mutasi lengkap</h3>
        <div class="table-wrap">
          <table class="findings statement">
            <thead>
              <tr><th>Tanggal</th><th>Keterangan</th><th>Kanal</th><th class="right">Nominal</th><th class="right">Saldo</th></tr>
            </thead>
            <tbody>
              {data.rows.map((r, i) => (
                <tr key={`${r.id}-${i}`} class={r.kategori === 'taruhan' || r.kategori === 'mewah' ? 'row-flag' : ''}>
                  <td class="mono">{r.tanggal}</td>
                  <td>
                    {r.merchant}
                    {r.kategori !== 'masuk' && (
                      <span class={`pill cat-${CAT_TONE[r.kategori] || 'plain'}`}>{CAT_LABEL[r.kategori] || r.kategori}</span>
                    )}
                  </td>
                  <td class="caption">{r.channel}</td>
                  <td class={`mono right ${r.arah === 'masuk' ? 'credit' : ''}`}>
                    {r.arah === 'masuk' ? '+' : '−'}{rupiah(r.nominal)}
                  </td>
                  <td class="mono right">{rupiah(r.saldo)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <p class="caption chain-foot">
          Mutasi ini berasal dari feed bank (dummy pada demo). Kampus tidak membuka
          rekening mahasiswa; data datang lewat kerja sama, dan hanya pola yang
          ditandai, bukan seluruh riwayat, yang masuk daftar periksa.
        </p>
      </section>
    </main>
  )
}
