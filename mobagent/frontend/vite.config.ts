import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: { '@': path.resolve(__dirname, 'src') },
  },
  server: {
    port: 5175,
    proxy: {
      '/api': { target: 'http://127.0.0.1:8790', changeOrigin: true },
      '/ws': { target: 'ws://127.0.0.1:8790', ws: true },
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
  },
})
