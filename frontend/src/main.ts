import { createApp } from 'vue'
import { VueQueryPlugin, QueryClient } from '@tanstack/vue-query'
import { router } from './router'
import App from './App.vue'
import './style.css'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import { setUnauthorizedRedirect } from './composables/auth/session'
import { resolveAuthSession } from './composables/auth/useAuth'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      gcTime: 1000 * 60 * 10,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

// After a deploy, a tab still running the previous build asks for chunks
// that no longer exist. Reload once to pick up the new build instead of
// surfacing "error loading dynamically imported module". The timestamp
// guard stops a reload loop if a chunk is genuinely missing.
window.addEventListener('vite:preloadError', (event) => {
  const key = 'st:chunk-reload-at'
  let last = 0
  try {
    last = Number(sessionStorage.getItem(key)) || 0
    sessionStorage.setItem(key, String(Date.now()))
  } catch {
    // storage unavailable: still reload, the guard just can't apply
  }
  if (Date.now() - last < 10_000) return
  event.preventDefault()
  window.location.reload()
})

setUnauthorizedRedirect(() => {
  if (router.currentRoute.value.name === 'signin') return
  void router.push({
    name: 'signin',
    query: { redirect: router.currentRoute.value.fullPath },
  })
})

const app = createApp(App)

app.use(router)
app.use(VueQueryPlugin, { queryClient })
app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      darkModeSelector: '[data-theme="dark"]',
      cssLayer: false,
    },
  },
})

void resolveAuthSession().then(() => {
  app.mount('#app')
})
