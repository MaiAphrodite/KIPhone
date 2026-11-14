import { useState } from 'preact/hooks'
import { Landing } from './screens/Landing.jsx'
import { Dashboard } from './screens/Dashboard.jsx'

export function App() {
  const [view, setView] = useState('landing')

  if (view === 'landing') return <Landing onStart={() => setView('app')} />

  return <Dashboard />
}
