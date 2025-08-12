import { defineConfig } from 'vite'
import mkcert from 'vite-plugin-mkcert'

// https://vite.dev/config/
export default defineConfig({
  plugins: [mkcert()],
  server: {
    https: {},
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8443",
        changeOrigin: true,
      },
    },
  },
})
