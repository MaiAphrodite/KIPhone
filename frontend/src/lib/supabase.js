import { createClient } from '@supabase/supabase-js'

const url = import.meta.env.VITE_SUPABASE_URL
const key = import.meta.env.VITE_SUPABASE_ANON_KEY

// null when Supabase is not configured yet, so the UI still runs against the
// Go backend in dev mode (X-Dev-User).
export const supabase = url && key ? createClient(url, key) : null

export async function accessToken() {
  if (!supabase) return null
  const { data } = await supabase.auth.getSession()
  return data.session?.access_token ?? null
}
