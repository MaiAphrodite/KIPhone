You are the FRONTEND worker. Repo: this worktree (Vite + Preact SPA, no TypeScript). Work FAST and in small steps. Do not explore beyond the files listed.

READ ONLY THESE FIRST: frontend/src/main.jsx, src/app.jsx, src/styles/tokens.css, src/styles/app.css, src/styles/kiphone.css, src/lib/kiphone.js.

TASK: add a warm editorial "Claude-style" theme and apply it. Do NOT rewrite component logic. CSS + minimal JSX class tweaks only.

STEP 1: create `frontend/src/styles/theme-claude.css` and import it in `main.jsx` AFTER kiphone.css:
```css
:root {
  --paper:#faf9f5; --surface:#ffffff; --surface-alt:#f5f4ef;
  --ink:#141413; --ink-2:#3d3b36; --muted:#6b6862;
  --hairline:#e8e6dc; --hairline-strong:#d8d5c8;
  --clay:#d97757; --clay-soft:#f4e5df; --danger:#b3352f; --ok:#2f6b3f;
  --radius:8px; --ease:cubic-bezier(.22,.61,.36,1);

  /* remap the existing token variables the components already use */
  --color-bg:var(--paper); --color-bg-elevated:var(--surface);
  --color-text-primary:var(--ink); --color-text-secondary:var(--ink-2);
  --color-text-tertiary:var(--muted); --color-border:var(--hairline);
  --color-accent:var(--clay);
}
body { background:var(--paper); color:var(--ink); font-family:system-ui,-apple-system,"Segoe UI",sans-serif; }
h1,h2,h3,.chain-title { font-family:Georgia,"Times New Roman",serif; font-weight:500; letter-spacing:-0.01em; color:var(--ink); }
```
Then add overrides so these components look right on paper: `.shell`, `.masthead`, `.stats-grid`/`.stat-card` (remove card backgrounds: keep hairline separators instead), `.btn` (clay background, white text), `.ghost` (transparent with hairline border), `.field`, `.findings th/td` (hairline rows, hover `--surface-alt`), `.pill` tones (clay-soft for accents, muted grey for neutral), `.drawer`/`.drawer-backdrop` (paper panel), `.explanation` (serif italic quote, border-left clay), `.receipt`, `.eval-strip`, `.chip`, `.banner`.

STEP 2 — motion, in the same file:
```css
@keyframes fade-up { from {opacity:0; transform:translateY(8px)} to {opacity:1; transform:none} }
.fade-up { animation: fade-up .32s var(--ease) both; }
.drawer { animation: drawer-in .24s var(--ease) both; }
@keyframes drawer-in { from { transform:translateX(100%) } to { transform:none } }
.drawer-backdrop { animation: fade-in .16s linear both; }
@keyframes fade-in { from {opacity:0} to {opacity:1} }
@media (prefers-reduced-motion: reduce) { *,*::before,*::after { animation:none!important; transition:none!important } }
```
In `screens/Dashboard.jsx`: add `class="fade-up"` to the headline block and to each `.stat-card` with inline `style={{ animationDelay: (i*60)+'ms' }}`. In `screens/Findings.jsx`: add `class="fade-up"` and inline `style={{ animationDelay: (i*20)+'ms' }}` to each row.

STEP 3 — verify:
- `npm run build` in frontend/ and paste the output; it MUST pass.
- Fix only compile errors you introduced.

RULES: no new dependencies; no Go changes; do not touch lib/kiphone.js request shapes; keep files under 250 lines; do not restructure components beyond classes/imports. When done, print a 5-line summary of files touched and the build result. Commit nothing.
