import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

const apiProxy = (proxyTarget: string) => ({
  '/api': {
    target: proxyTarget,
    changeOrigin: true,
    rewrite: (path: string) => path.replace(/^\/api/, '') || '/',
  },
})

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_PROXY_TARGET || 'http://127.0.0.1:8080'
  return {
    plugins: [
      tailwindcss(),
      vue(),
    ],
    resolve: {
      alias: {
        '~': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    build: {
      rolldownOptions: {
        output: {
          codeSplitting: {
            // Every page needs these. One stable chunk instead of letting automatic splitting cut
            // them into fragments that regroup (and add entry requests) whenever a lazy import
            // somewhere else changes.
            groups: [
              {
                name: 'framework',
                test: /[\\/]node_modules[\\/](@vue|vue|vue-router|@tanstack|axios)[\\/]/,
              },
            ],
          },
        },
      },
    },
    server: {
      host: true,
      port: 3000,
      proxy: apiProxy(proxyTarget),
    },
    preview: {
      proxy: apiProxy(proxyTarget),
    },
  }
})

