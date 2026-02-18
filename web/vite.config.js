import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "build",
  },
  server: {
    port: 3000,
    proxy: {
      "/pac-go-server": {
        target: "http://localhost:8000",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/pac-go-server/, '/api/v1'),
      }
    },
  },
});
