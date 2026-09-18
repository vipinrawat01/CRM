import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // In dev, run the Go backend separately (see backend/README) and let
      // Vite proxy API calls to it so the frontend can use relative fetch()
      // URLs identically to the production build.
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    // Production build output the Go server serves as static files
    // (see FRONTEND_DIR in backend/.env).
    outDir: 'dist',
  },
})
