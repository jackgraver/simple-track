<script setup lang="ts">
import { ChevronLeft, Menu } from "lucide-vue-next";
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { isDevEnvironment } from "~/composables/auth/session";
import AccountMenu from "../AccountMenu.vue";
import { IconButton } from "~/shared/ui";
import { activeItem, isNavRoot, itemLocation } from "../nav";
import { useNavDrawer } from "./useNavDrawer";

const route = useRoute();
const router = useRouter();

const title = computed(
    () => route.meta.breadcrumb ?? activeItem(route)?.item.label ?? "simpletrack",
);
const showBack = computed(() => !isNavRoot(route));
const { openDrawer } = useNavDrawer();

function goBack() {
    // Pages with a fixed logical parent skip history, so Back can't loop between two pages.
    const parent = route.meta.backTo;
    if (parent) {
        const offset = route.query.offset;
        void router.push({
            name: parent,
            query: typeof offset === "string" && offset !== "" ? { offset } : {},
        });
        return;
    }
    // Prefer real history (keeps scroll/state); fall back to the section's
    // nav entry when the page was opened directly (e.g. PWA cold start).
    if (window.history.state?.back) {
        router.back();
        return;
    }
    const current = activeItem(route);
    if (current && current.item.to !== route.name) {
        void router.push(itemLocation(current.item, route));
    } else {
        void router.push("/");
    }
}
</script>

<template>
    <header class="flex min-h-touch items-center gap-1 pt-2 pb-3">
        <IconButton label="Open menu" class="-ml-2" @click="openDrawer"><Menu /></IconButton>
        <IconButton v-if="showBack" label="Back" @click="goBack"><ChevronLeft /></IconButton>

        <h1 class="m-0 min-w-0 flex-1 truncate text-lg font-semibold tracking-tight text-fg">
            {{ title }}
        </h1>

        <span v-if="isDevEnvironment" class="rounded-control bg-surface-3 px-1.5 text-xs text-fg-muted"
            >dev</span
        >
        <AccountMenu placement="header" />
    </header>
</template>
