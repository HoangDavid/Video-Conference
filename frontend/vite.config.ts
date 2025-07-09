import { defineConfig } from 'vite'
import mkcert from 'vite-plugin-mkcert'

// https://vite.dev/config/
export default defineConfig({
  plugins: [mkcert()],
  server: {
    https: {},
    port: 5173,
  },
})
