import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { viteSingleFile } from "vite-plugin-singlefile";
import { fileURLToPath, URL } from "node:url";

// MedClaim Vue — dibangun sebagai satu file HTML (inline JS/CSS)
// agar dapat disajikan mulus oleh backend Go di /app/.
export default defineConfig({
  plugins: [vue(), viteSingleFile()],
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
  },
  base: process.env.VERCEL ? "/" : "/app/",
  build: {
    outDir: process.env.VERCEL ? "dist" : "dist",
    emptyOutDir: true,
    target: "es2020",
    cssCodeSplit: false,
    rollupOptions: {
      output: {
        inlineDynamicImports: true,
      },
    },
  },
});
