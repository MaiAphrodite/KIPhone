import { useState } from 'preact/hooks'
import { Landing } from './screens/Landing.jsx'
import { Dashboard } from './screens/Dashboard.jsx'
import { PublicView } from './screens/PublicView.jsx'

export function App() {
  const [view, setView] = useState('landing')

  if (view === 'public') return <PublicView onBack={() => setView('landing')} />
  if (view === 'landing') {
    return <Landing onStart={() => setView('app')} onPublic={() => setView('public')} />
  }

  return <Dashboard />
}
