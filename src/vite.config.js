import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const domain = env.VITE_APP_DOMAIN || 'https://007.hallo.run'

  return {
    plugins: [
      vue(),
      {
        name: 'html-domain-replace',
        transformIndexHtml(html) {
          return html.replace(/\{\{DOMAIN\}\}/g, domain)
        },
      },
    ],
    build: {
      outDir: '../cmd/web',
      emptyOutDir: true,
    },
    server: {
      proxy: {
        '/api': 'http://localhost:8088'
      },
    },
  }
})