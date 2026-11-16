import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'

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
      AutoImport({
        imports: [
          'vue',
          'vue-router',
          {
            '@tanstack/vue-query': ['useQuery', 'useMutation', 'useQueryClient'],
          },
        ],
        dts: true,
        vueTemplate: true,
      }),
      Components({
        dirs: ['src/components'],
        dts: true,
      }),
    ],
    resolve: {
      alias: {
        '~': fileURLToPath(new URL('./src', import.meta.url)),
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

