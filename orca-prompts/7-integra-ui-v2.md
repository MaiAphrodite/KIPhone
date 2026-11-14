You are the FRONTEND DESIGN worker for Integra. Repo: this worktree (Vite + Preact SPA, no TypeScript, no router library).

GOAL: replace the current dark Linear-style UI with a minimalist warm editorial dashboard in the **Anthropic / Claude** aesthetic, with tasteful motion. It must remain a working SPA wired to the live API.

READ FIRST: PRD.md, frontend/src/app.jsx, screens/Dashboard.jsx, screens/Findings.jsx, components/EvidenceDrawer.jsx, components/ChainPanel.jsx, lib/integra.js, styles/tokens.css, styles/app.css, styles/integra.css.

API IS LIVE at http://127.0.0.1:8080 (seeded, 15 findings). All endpoints already exist in lib/integra.js — DO NOT change request shapes.

## Design system to implement

Create `frontend/src/styles/theme-claude.css`, imported AFTER tokens.css in main.jsx, that redefines the CSS variables:

```
--paper: #faf9f5;  --surface: #ffffff;  --surface-alt: #f5f4ef;
--ink: #141413;    --ink-2: #3d3b36;    --muted: #6b6862;
--hairline: #e8e6dc;  --hairline-strong: #d8d5c8;
--clay: #d97757;   --clay-soft: #f4e5df;  --danger: #b3352f; --ok: #2f6b3f;
--radius: 8px;  --ease: cubic-bezier(.22,.61,.36,1);
```

Rules:
- Warm paper background everywhere. NO dark surfaces, NO gradients, NO glass, NO emoji, NO shadows.
- Serif (Georgia, 'Times New Roman', serif) for display headline, screen titles, and the evidence quote. Sans (system-ui) for everything else.
- 1px hairlines only. Max radius 8px. Generous vertical rhythm (sections separated by 48–64px).
- Clay #d97757 is the ONLY accent: primary button, active filter, risk bar fill for high band. Danger/ok only for status text.

## Screens to build (all in the SPA, no page reloads)

1. **Landing** (`screens/Landing.jsx`) — keep it as the entry view but restyle: serif display headline 'Integritas penyaluran, terlihat.', muted subhead, clay CTA 'Buka dashboard', three short trust lines with hairline separators (not cards). Subtle fade-up on mount.

2. **Dashboard** (`screens/Dashboard.jsx`):
   - Sticky top bar: serif wordmark `Integra`, muted 'KIP-K disbursement integrity', right-aligned text button 'Muat ulang data contoh'.
   - Headline block with a one-line summary of what the page shows.
   - **Metrics row**: 4 numbers (Total Record, Temuan, Prioritas Tinggi, Salah Deteksi) separated by vertical hairlines — NOT cards. Small-caps label above, large numeral below, tabular-nums.
   - Action row: clay button 'Jalankan Deteksi', caption with runtime, error text if any.
   - Eval footnote strip: `precision@10 · recall · false positive · metodologi data sintetis`.
   - Findings table (keep existing logic and filters) restyled: mono NIM, sans nama, outlined R-chips, thin risk bar + numeric, muted status pill, 'Lihat bukti' text action. Row hover `--surface-alt`. Staggered row fade-in 20ms per row.
   - Chain receipt band at the bottom (reuse ChainPanel logic) restyled to this palette.

3. **Evidence drawer** (`components/EvidenceDrawer.jsx`) — keep logic; restyle to 480px paper panel sliding from the right, backdrop fade, serif italic explanation quote, two-column definition list for evidence fields, pill badges, lifecycle buttons (clay primary for Verifikasi, outlined danger for Tandai Salah Deteksi, plain for Selesai), event history list.

4. **View transitions**: switching landing → dashboard must fade/translate in over 180ms. No jQuery, no router lib — CSS classes + a small `useEffect` toggle is enough.

## Motion spec
```css
@keyframes fade-up { from { opacity:0; transform: translateY(8px) } to { opacity:1; transform:none } }
@media (prefers-reduced-motion: reduce) { *, *::before, *::after { animation: none !important; transition: none !important } }
```
- Headline + metrics: fade-up, staggered 60ms via inline `style={{ animationDelay }}`.
- Table rows: fade-up with `animationDelay: index * 20ms`, re-key the tbody on filter change so it replays.
- Drawer: `transform: translateX(100%)` → `0`, 240ms `--ease`; backdrop opacity 0 → 1 over 160ms.
- Buttons: background/color 120ms. No scale/bounce.

## Constraints
- Do not change `lib/integra.js` request shapes or any Go file.
- Keep `frontend/src/main.jsx` importing tokens.css, app.css, integra.css, theme-claude.css in that order.
- Delete or neutralize now-unused dark-theme rules in integra.css if they fight the new theme; prefer adding overrides in theme-claude.css.
- No new dependencies.
- Keep each file under ~220 lines.

## Finish by
1. `npm run build` — paste the output. It MUST pass.
2. `curl -s http://127.0.0.1:8080/api/findings -H 'X-Dev-User: admin' | head -c 200` to confirm data shape.
3. Write `frontend/DESIGN-NOTES.md`: what you changed, the palette, the motion timings, and anything unfinished.
Commit nothing.
