const rawApiBase = import.meta.env.VITE_API_BASE
const trimmed =
  typeof rawApiBase === 'string' ? rawApiBase.trim() : ''
export const config = {
  apiBase: trimmed !== '' ? trimmed : '/api',
}

