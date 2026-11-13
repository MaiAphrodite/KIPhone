import { accessToken } from './supabase.js'

async function request(path, options = {}) {
  const token = await accessToken()
  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) }
  if (token) headers.Authorization = `Bearer ${token}`
  else headers['X-Dev-User'] = 'dev-user'

  const res = await fetch(`/api${path}`, { ...options, headers })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.status === 204 ? null : res.json()
}

export const listItems = () => request('/items')
export const createItem = (item) => request('/items', { method: 'POST', body: JSON.stringify(item) })
export const deleteItem = (id) => request(`/items/${id}`, { method: 'DELETE' })
