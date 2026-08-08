import { mergeConfig } from "vite";
import { defineConfig } from "vitest/config";
import viteConfig from "./vite.config";

export default defineConfig(({ mode }) =>
    mergeConfig(
        viteConfig({
            command: "serve",
            mode,
            isSsrBuild: false,
            isPreview: false,
        }),
        {
            test: {
                environment: "node",
                include: ["src/**/*.test.ts"],
            },
        },
    ),
);
