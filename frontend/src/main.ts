import { createApp } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'
import { router } from './router'
import App from './App.vue'
import '@fontsource-variable/plus-jakarta-sans/wght.css'
import './style.css'
import { setUnauthorizedRedirect } from './composables/auth/session'
import { resolveAuthSession } from './composables/auth/useAuth'
import { queryClient } from './api/queryClient'
import { initTheme } from './composables/useTheme'
import { clearUserStorage } from './utils/userStorage'

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
  // Session expired: drop the previous user's cached data before the sign-in page.
  queryClient.clear()
  clearUserStorage()
  if (router.currentRoute.value.name === 'signin') return
  void router.push({
    name: 'signin',
    query: { redirect: router.currentRoute.value.fullPath },
  })
})

initTheme()

const app = createApp(App)

app.use(router)
app.use(VueQueryPlugin, { queryClient })

// The landing route's queries start in the router's first `beforeEach` (see router/index.ts).
void resolveAuthSession().then(() => {
  app.mount('#app')
})
