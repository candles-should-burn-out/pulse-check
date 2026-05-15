import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import type { Plugin } from "vite";

export default defineConfig({
  plugins: [react(), appFallbackPlugin()],
  build: {
    rollupOptions: {
      input: {
        landing: "index.html",
        app: "spa.html",
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
});

function appFallbackPlugin(): Plugin {
  return {
    name: "app-fallback",
    configureServer(server) {
      server.middlewares.use((request, _response, next) => {
        if (request.url === "/app" || request.url?.startsWith("/app/")) {
          request.url = "/spa.html";
        }

        next();
      });
    },
  };
}
