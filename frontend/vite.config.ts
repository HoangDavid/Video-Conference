import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import mkcert from "vite-plugin-mkcert";

export default defineConfig({
  plugins: [react(), mkcert()],
  server: {
    https: {},
    port: 5173,
    proxy: {
      "/api": {
        target: "https://localhost:8443",
        changeOrigin: true,
        secure: false,
      },
      "/ws": {
        target: "wss://localhost:8443",
        changeOrigin: true,
        secure: false,
        ws: true,
      },
    },
  },
});
