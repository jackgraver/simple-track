<script setup lang="ts">
import type { RouteLocationRaw } from "vue-router";

export type ModuleTitleLink = {
    name: string;
    to?: string;
    offset?: boolean;
    query?: Record<string, string>;
    onClick?: () => void;
};

const props = defineProps<{
    title: string;
    dayOffset?: number;
    links: ModuleTitleLink[];
}>();

const linkTo = (link: ModuleTitleLink): RouteLocationRaw | undefined => {
    if (!link.to) return undefined;
    if (link.query) {
        return { name: link.to, query: link.query };
    }
    const offset = props.dayOffset ?? 0;
    if (link.offset && offset !== 0) {
        return { name: link.to, query: { offset: String(offset) } };
    }
    return { name: link.to };
};
</script>
<template>
    <div
        class="flex items-center justify-between gap-4 border-b border-(--color-border) pb-3"
    >
        <h1 class="m-0 text-md font-semibold text-textSecondary">
            {{ title }}
        </h1>
        <nav
            v-if="links.length"
            class="flex items-center gap-3 text-sm text-textSecondary"
        >
            <template v-for="(link, index) in links" :key="link.name">
                <router-link
                    v-if="link.to"
                    :to="linkTo(link)!"
                    class="transition-colors hover:text-textPrimary"
                    >{{ link.name }}</router-link
                >
                <button
                    v-else
                    type="button"
                    class="p-0! transition-colors hover:text-textPrimary"
                    @click="link.onClick?.()"
                >
                    {{ link.name }}
                </button>
                <span
                    v-if="index < links.length - 1"
                    aria-hidden="true"
                    class="text-textSecondary/50"
                    >·</span
                >
            </template>
        </nav>
    </div>
</template>
