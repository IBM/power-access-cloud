import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const proxyTarget = env.VITE_PAC_GO_SERVER_TARGET || env.REACT_APP_PAC_GO_SERVER_TARGET || "http://localhost:8000";

  return {
    plugins: [react()],
    build: {
      outDir: "build",
    },
    server: {
      port: 3000,
      proxy: {
        "/pac-go-server": {
          target: proxyTarget,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/pac-go-server/, '/api/v1'),
        }
      },
    },
  };
});
