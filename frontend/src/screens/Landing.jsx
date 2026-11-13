// Imported from the OpenDesign artifact (design/landing.html).
// Only CSS variables from src/styles/tokens.css are used, so this screen keeps
// the same brand contract as the deck and the prototype.
export function Landing({ onStart }) {
  return (
    <main class="shell">
      <header class="masthead">
        <h1>Meetings that end with the notes already written.</h1>
        <p class="muted">
          Notewise joins your calls, captures what was said, and turns every meeting
          into structured notes with action items and owners.
        </p>
        <button class="btn" onClick={onStart}>Start free</button>
      </header>

      <section class="features">
        {[
          ['Live transcription', 'Every speaker labelled as it happens, in 40 languages.'],
          ['Structured summaries', 'Decisions, risks and owners — not a wall of text.'],
          ['Action items', 'Assigned, dated and pushed to the tools your team already uses.'],
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
