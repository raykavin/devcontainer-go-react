import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 8080,
    // Fail instead of silently moving to another port: compose publishes exactly 8080
    strictPort: true,
    proxy: {
      // Forward API calls to the Go backend inside the same container
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
})
