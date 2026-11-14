// First screen of the demo. Visual language comes from the OpenDesign artifact
// (design/landing.html); the copy is Integra's.
export function Landing({ onStart }) {
  return (
    <main class="shell">
      <header class="masthead">
        <p class="eyebrow">Integritas penyaluran beasiswa</p>
        <h1>Dana KIP-K cair ke orang yang salah. Sekarang terlihat.</h1>
        <p class="muted">
          Integra memindai batch penyaluran dan menandai anomali yang bisa
          dibuktikan — NIK ganda, rekening bersama, nama tidak cocok, penerima
          nonaktif. Model tidak memblokir apa pun. Manusia yang memutuskan.
        </p>
        <div class="action-row" style="margin-top:24px">
          <button class="btn" onClick={onStart}>Buka dashboard</button>
          <span class="caption">215 record contoh · 15 temuan · ~30 ms</span>
        </div>
      </header>

      <section class="features">
        {[
          ['Bukti, bukan tuduhan', 'Setiap temuan membawa field yang berbeda, tanggal, dan sumbernya.'],
          ['Transparan soal akurasi', 'Temuan bisa ditandai salah deteksi, dan itu tercatat.'],
          ['Offline, bisa diaudit', 'Empat aturan deterministik di Go. Tidak ada kotak hitam.'],
        ].map(([title, body]) => (
          <article class="card feature" key={title}>
            <h2>{title}</h2>
            <p class="muted">{body}</p>
          </article>
        ))}
      </section>
    </main>
  )
}
