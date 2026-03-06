import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api/v2': {
        target: 'https://api-v2.zrok.io',
        changeOrigin: true,
      }
    },
    allowedHosts: [
      ".share.zrok.io"
    ]
  }
})
