// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
    compatibilityDate: "2025-07-15",
    devtools: { enabled: true },
    app: {
        head: {
            title: "Simple Tracker",
            htmlAttrs: { lang: "en" },
            link: [{ rel: "icon", type: "image/x-icon", href: "/favicon.ico" }],
        },
    },
});
