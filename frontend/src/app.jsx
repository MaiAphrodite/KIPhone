import { useEffect, useState } from 'preact/hooks'
import { listItems, createItem, deleteItem } from './lib/api.js'
import { supabase } from './lib/supabase.js'

export function App() {
  const [items, setItems] = useState([])
  const [title, setTitle] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [user, setUser] = useState(null)

  async function refresh() {
    try {
      setError('')
      const data = await listItems()
      setItems(data.items ?? [])
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
    if (!supabase) return
    supabase.auth.getUser().then(({ data }) => setUser(data.user ?? null))
    const { data: sub } = supabase.auth.onAuthStateChange((_event, session) => {
      setUser(session?.user ?? null)
      refresh()
    })
    return () => sub.subscription.unsubscribe()
  }, [])

  async function add(e) {
    e.preventDefault()
    const value = title.trim()
    if (!value) return
    setTitle('')
    try {
      await createItem({ title: value })
      await refresh()
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <main class="shell">
      <header class="masthead">
        <h1>Hackstack</h1>
        <p class="muted">Go · GORM · Supabase · Vite · Preact</p>
      </header>

      <section class="card">
        <form onSubmit={add} class="row">
          <input
            class="field"
            value={title}
            placeholder="Add an item…"
            onInput={(e) => setTitle(e.currentTarget.value)}
          />
          <button class="btn" type="submit">Add</button>
        </form>

        {error && <p class="error">{error}</p>}
        {loading ? (
          <p class="muted">Loading…</p>
        ) : items.length === 0 ? (
          <p class="muted">Nothing yet. Add the first item.</p>
        ) : (
          <ul class="list">
            {items.map((item) => (
              <li key={item.id} class="list-row">
                <span>{item.title}</span>
                <button class="ghost" onClick={async () => { await deleteItem(item.id); refresh() }}>
                  Remove
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>

      <footer class="muted small">
        {user ? `signed in as ${user.email}` : supabase ? 'not signed in' : 'dev mode (no Supabase configured)'}
      </footer>
    </main>
  )
}
