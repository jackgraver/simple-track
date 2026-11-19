import type { StorybookConfig } from "@storybook/vue3-vite";

const config: StorybookConfig = {
    stories: ["../src/**/*.stories.ts"],
    // addon-mcp exposes the component and story docs to coding agents at /mcp (dev server only).
    addons: ["@storybook/addon-a11y", "@storybook/addon-themes", "@storybook/addon-mcp"],
    framework: { name: "@storybook/vue3-vite", options: {} },
    core: { disableTelemetry: true },
};
export default config;
