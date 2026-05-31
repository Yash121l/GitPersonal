import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig(({ command }) => ({
  base: '/app/assets/',
  plugins: [
    vue(),
    tailwindcss(),
    ...(command === 'serve' ? [vueDevTools()] : []),
  ],
  build: {
    outDir: '../internal/server/ui/dist',
    emptyOutDir: true,
    manifest: 'manifest.json',
    cssCodeSplit: false,
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
}))
