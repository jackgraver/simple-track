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
      darkModeSelector: '.dark',
      cssLayer: false,
    },
  },
})

void resolveAuthSession().then(() => {
  app.mount('#app')
})
