import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import wails from "@wailsio/runtime/plugins/vite";
import tailwindcss from "@tailwindcss/vite";
import { fileURLToPath } from "node:url";

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  resolve: { alias: {
    "@claudepro/ui": fileURLToPath(new URL("./src/components/ui/index.ts", import.meta.url)),
    cn: fileURLToPath(new URL("./src/components/ui/cn.ts", import.meta.url)),
  } },
  plugins: [react(), tailwindcss(), wails("./bindings")],
});
