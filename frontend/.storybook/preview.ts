import type { Preview } from "@storybook/vue3-vite";
import { withThemeByDataAttribute } from "@storybook/addon-themes";
import "../src/style.css";

// Same contract as the app: tokens flip on <html data-theme="dark|light">.
const preview: Preview = {
    initialGlobals: { theme: "dark" },
    parameters: {
        layout: "centered",
        viewport: {
            options: {
                phone: { name: "Phone 390×844", styles: { width: "390px", height: "844px" }, type: "mobile" },
                desktop: { name: "Desktop 1280×800", styles: { width: "1280px", height: "800px" }, type: "desktop" },
            },
        },
        a11y: { test: "todo" },
        backgrounds: { disable: true },
    },
    decorators: [
        withThemeByDataAttribute({
            themes: { dark: "dark", light: "light" },
            defaultTheme: "dark",
            attributeName: "data-theme",
        }),
        (story, ctx) => {
            // The themes addon applies the attribute after mount; set it now so anything that
            // reads tokens during setup (charts) sees the right theme.
            document.documentElement.dataset.theme = (ctx.globals.theme as string) || "dark";
            return { components: { story }, template: `<story :key="ctx.globals.theme" />`, setup: () => ({ ctx }) };
        },
    ],
};

export default preview;
