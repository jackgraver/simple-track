export const config = {
  apiBase:
    import.meta.env.VITE_API_BASE ??
    (import.meta.env.DEV ? '/api' : 'http://localhost:8080'),
}

