import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
    },
  },
  server: {
    // IPv4 루프백 고정: WSL mirrored 모드에서 Windows 브라우저가 확실히 닿는 주소
    host: '127.0.0.1',
    port: 5173,
    strictPort: true,
    // 개발 중에는 /api 요청을 Go 서버로 넘긴다.
    // 브라우저 입장에선 페이지와 API가 같은 출처(localhost:5173)라서
    // 쿠키·CSRF(Sec-Fetch-Site: same-origin) 모두 운영 환경과 똑같이 동작한다.
    proxy: {
      '/api': 'http://127.0.0.1:8080',
    },
  },
})
