<script setup lang="ts">
import { Check, ChevronDown, ChevronLeft, CircleUser } from "lucide-vue-next";
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { isDevEnvironment } from "~/composables/auth/session";
import {
    activeGroup,
    activeItem,
    isItemActive,
    isNavRoot,
    itemLocation,
} from "../nav";

const route = useRoute();
const router = useRouter();

const group = computed(() => activeGroup(route));
const title = computed(
    () => route.meta.breadcrumb ?? activeItem(route)?.item.label ?? "simpletrack",
);
const showBack = computed(() => !isNavRoot(route));
const hasMenu = computed(() => (group.value?.items.length ?? 0) > 1);

function goBack() {
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

// Section menu: the title doubles as a dropdown of the section's pages.
const menuOpen = ref(false);
const menuRoot = ref<HTMLElement | null>(null);
watch(() => route.fullPath, () => (menuOpen.value = false));
function onDocClick(ev: MouseEvent) {
    if (menuRoot.value && !menuRoot.value.contains(ev.target as Node)) {
        menuOpen.value = false;
    }
}
function onKey(ev: KeyboardEvent) {
    if (ev.key === "Escape") menuOpen.value = false;
}
onMounted(() => {
    document.addEventListener("click", onDocClick);
    document.addEventListener("keydown", onKey);
});
onUnmounted(() => {
    document.removeEventListener("click", onDocClick);
    document.removeEventListener("keydown", onKey);
});
</script>

<template>
    <header class="flex min-h-touch items-center gap-1 pt-2 pb-3">
        <button
            v-if="showBack"
            type="button"
            class="-ml-2 m-0! flex size-touch shrink-0 items-center justify-center rounded-control p-0! text-fg-muted shadow-none! hover:bg-surface-3 hover:text-fg"
            aria-label="Back"
            @click="goBack"
        >
            <ChevronLeft class="size-6" />
        </button>

        <div ref="menuRoot" class="relative min-w-0 flex-1">
            <button
                v-if="hasMenu"
                type="button"
                class="m-0! flex max-w-full items-center gap-1 rounded-control bg-transparent p-0! text-left shadow-none!"
                :aria-expanded="menuOpen"
                aria-haspopup="menu"
                @click="menuOpen = !menuOpen"
            >
                <h1 class="m-0 truncate text-lg font-semibold tracking-tight text-fg">
                    {{ title }}
                </h1>
                <ChevronDown
                    class="size-4 shrink-0 text-fg-muted transition-transform"
                    :class="{ 'rotate-180': menuOpen }"
                />
            </button>
            <h1 v-else class="m-0 truncate text-lg font-semibold tracking-tight text-fg">
                {{ title }}
            </h1>

            <div
                v-if="menuOpen && group"
                role="menu"
                :aria-label="`${group.label} pages`"
                class="absolute top-full left-0 z-popover mt-1 min-w-52 rounded-card border border-line bg-surface py-1 shadow-raised"
            >
                <p class="m-0 px-3 pt-1.5 pb-1 text-xs font-semibold tracking-wide text-fg-muted uppercase">
                    {{ group.label }}
                </p>
                <router-link
                    v-for="item in group.items"
                    :key="item.to"
                    :to="itemLocation(item, route)"
                    role="menuitem"
                    class="flex items-center justify-between gap-4 px-3 py-2.5 text-sm"
                    :class="
                        isItemActive(item, route)
                            ? 'font-medium text-fg'
                            : 'text-fg-muted hover:bg-surface-3 hover:text-fg'
                    "
                    :aria-current="isItemActive(item, route) ? 'page' : undefined"
                >
                    {{ item.label }}
                    <Check v-if="isItemActive(item, route)" class="size-4" />
                </router-link>
            </div>
        </div>

        <span v-if="isDevEnvironment" class="rounded-control bg-surface-3 px-1.5 text-xs text-fg-muted"
            >dev</span
        >
        <router-link
            :to="{ name: 'settings-profile' }"
            class="flex size-touch items-center justify-center rounded-control text-fg-muted hover:bg-surface-3 hover:text-fg"
            aria-label="Profile"
        >
            <CircleUser class="size-6" />
        </router-link>
    </header>
</template>
