import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";

// The build goes into the server, which embeds it with go:embed.
const outDir = fileURLToPath(
    new URL("../server/internal/web/dist", import.meta.url)
);

export default defineConfig({
    plugins: [react(), tailwindcss()],
    resolve: {
        alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
    },
    server: {
        proxy: {
            // A server started with `make run` in server/.
            "/api": process.env.KUBIN_API ?? "http://localhost:8080",
        },
    },
    build: {
        outDir,
        emptyOutDir: true,
    },
});
