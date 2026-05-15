import { readFileSync } from "node:fs";
import { resolve } from "node:path";

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
      server.middlewares.use((request, response, next) => {
        const url = request.url ?? "/";
        const pathname = url.split("?")[0];

        if (pathname === "/app") {
          request.url = "/spa.html";
          next();
          return;
        }

        if (pathname.startsWith("/app/") && isKnownAppPath(pathname)) {
          request.url = "/spa.html";
          next();
          return;
        }

        if (shouldServeDevNotFound(pathname, request.headers.accept)) {
          response.statusCode = 404;
          response.setHeader("Content-Type", "text/html; charset=utf-8");
          response.end(
            readFileSync(resolve(server.config.root, "public/404.html"), "utf-8")
          );
          return;
        }

        next();
      });
    },
  };
}

function shouldServeDevNotFound(pathname: string, acceptHeader?: string | string[]) {
  const acceptsHtml = Array.isArray(acceptHeader)
    ? acceptHeader.some((value) => value.includes("text/html"))
    : acceptHeader?.includes("text/html");

  if (!acceptsHtml) {
    return false;
  }

  if (
    pathname === "/" ||
    pathname === "/index.html" ||
    pathname === "/spa.html" ||
    pathname === "/404.html"
  ) {
    return false;
  }

  return !(
    pathname.startsWith("/api") ||
    pathname.startsWith("/assets/") ||
    pathname.startsWith("/src/") ||
    pathname.startsWith("/@") ||
    pathname === "/favicon.ico"
  );
}

function isKnownAppPath(pathname: string) {
  const normalizedPathname =
    pathname.length > 1 ? pathname.replace(/\/+$/, "") : pathname;

  return (
    normalizedPathname === "/app" ||
    normalizedPathname === "/app/login" ||
    normalizedPathname === "/app/profile"
  );
}
