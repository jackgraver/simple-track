// ESLint: UI design-system rules (see docs/design/RULES.md).
//
// Existing violations are recorded in eslint-suppressions.json, so only NEW
// ones fail. After cleaning a file up, shrink the file with
//   npm run lint -- --prune-suppressions
// Never run --suppress-all / --suppress-rule to make new code pass.
//
// Rules are added as the overhaul progresses ("Planned rules" in RULES.md).
import pluginVue from "eslint-plugin-vue";
import betterTailwind from "eslint-plugin-better-tailwindcss";
import { getDefaultSelectors } from "eslint-plugin-better-tailwindcss/defaults";
import tseslint from "typescript-eslint";

const PALETTE =
    "red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|slate|gray|zinc|neutral|stone|black|white";
const COLOR_UTILITIES =
    "bg|text|border|ring|ring-offset|outline|fill|stroke|from|via|to|divide|decoration|caret|accent|shadow|placeholder";
// Optional variant prefixes (hover:, lg:, group-hover:...) and important marker.
const VARIANTS = "^!?(?:[\\w@\\[\\]&*>=:/.-]+:)*!?";

// Hex colours and rgb()/hsl() in JS strings and template expressions.
const COLOR_LITERAL = "/#[0-9a-fA-F]{3,8}\\b|\\b(?:rgba?|hsla?)\\(/";
const COLOR_LITERAL_MESSAGE =
    "No colour literals outside src/style.css. Use a token (docs/design/TOKENS.md); chart code reads tokens via shared/chart/theme.ts.";
const colorLiteralSelectors = [
    { selector: `Literal[value=${COLOR_LITERAL}]`, message: COLOR_LITERAL_MESSAGE },
    { selector: `TemplateElement[value.raw=${COLOR_LITERAL}]`, message: COLOR_LITERAL_MESSAGE },
];

export default [
    { ignores: ["dist/**", "public/**", "coverage/**"] },

    ...pluginVue.configs["flat/base"],

    {
        files: ["**/*.{ts,vue}"],
        languageOptions: {
            parserOptions: {
                parser: tseslint.parser,
                extraFileExtensions: [".vue"],
                sourceType: "module",
            },
        },
    },
    {
        files: ["**/*.ts"],
        languageOptions: { parser: tseslint.parser },
    },

    {
        files: ["src/**/*.{ts,vue}"],
        plugins: { "better-tailwindcss": betterTailwind },
        settings: {
            "better-tailwindcss": {
                entryPoint: "src/style.css",
                selectors: [
                    ...getDefaultSelectors(),
                    // finances/format.ts keeps class strings in a `ui` object.
                    { kind: "variable", name: "^ui$", match: [{ type: "objectValues" }] },
                ],
            },
        },
        rules: {
            // Tokens only: no Tailwind palette colours, no arbitrary values for
            // things the token layer covers.
            "better-tailwindcss/no-restricted-classes": [
                "error",
                {
                    restrict: [
                        {
                            pattern: `${VARIANTS}-?(?:${COLOR_UTILITIES})-(?:${PALETTE})(?:-\\d{2,3})?(?:/\\d+)?!?$`,
                            message:
                                "Tailwind palette colour. Use a semantic token (text-fg-muted, bg-surface, border-line, text-danger...): docs/design/TOKENS.md.",
                        },
                        {
                            pattern: `${VARIANTS}-?(?:${COLOR_UTILITIES}|rounded(?:-[a-z]+)?|z)-\\[.+\\]!?$`,
                            message:
                                "Arbitrary value for something tokens cover (colour, type size, radius, shadow, z). Use or add a token: docs/design/TOKENS.md.",
                        },
                    ],
                },
            ],
            // Correctness: classes Tailwind doesn't know (e.g. the undefined
            // `bg-primary`), contradictory or duplicate classes.
            "better-tailwindcss/no-unknown-classes": "error",
            "better-tailwindcss/no-conflicting-classes": "error",
            "better-tailwindcss/no-duplicate-classes": "error",
            "better-tailwindcss/no-deprecated-classes": "error",

            // No component-local CSS: Tailwind classes + tokens only.
            "vue/no-restricted-block": [
                "error",
                {
                    element: "style",
                    message:
                        "No <style> blocks. Use Tailwind classes with tokens, or a primitive (docs/design/RULES.md).",
                },
            ],
            "vue/no-restricted-static-attribute": [
                "error",
                {
                    key: "style",
                    value: "/#[0-9a-fA-F]{3,8}|rgba?\\(|hsla?\\(/",
                    message: COLOR_LITERAL_MESSAGE,
                },
            ],
            "no-restricted-syntax": ["error", ...colorLiteralSelectors],
            "vue/no-restricted-syntax": ["error", ...colorLiteralSelectors],
        },
    },
];
