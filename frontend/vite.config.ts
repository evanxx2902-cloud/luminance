import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist', // Docker COPY --from=frontend-builder /build/frontend/dist/ /opt/luminance/web/
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8000',
      '/ai':  'http://localhost:8001',
    },
  },
})
