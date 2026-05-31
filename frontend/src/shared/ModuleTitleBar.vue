<script setup lang="ts">
import { MoreVertical } from "lucide-vue-next";
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
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

const NAV_GAP = 12;
const ROOT_GAP = 16;
const TITLE_LINK_MIN_GAP = 32;
const OVERFLOW_GAP = 8;
const OVERFLOW_BTN_WIDTH = 36;

const rootRef = ref<HTMLElement | null>(null);
const measureRef = ref<HTMLElement | null>(null);
const menuRoot = ref<HTMLElement | null>(null);
const menuOpen = ref(false);
const visibleCount = ref(props.links.length);
const linkWidths = ref<number[]>([]);
const separatorWidth = ref(0);

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

const visibleLinks = computed(() => props.links.slice(0, visibleCount.value));
const overflowLinks = computed(() => props.links.slice(visibleCount.value));

const navWidthFor = (visible: number): number => {
    if (visible <= 0) return 0;
    let width = 0;
    for (let i = 0; i < visible; i++) {
        width += linkWidths.value[i] ?? 0;
    }
    width += separatorWidth.value * (visible - 1);
    width += NAV_GAP * Math.max(0, visible * 2 - 2);
    return width;
};

const getAvailableNavWidth = (): number => {
    if (!rootRef.value) return Infinity;
    const title = rootRef.value.querySelector("h1");
    if (!title) return Infinity;
    const rootWidth = rootRef.value.clientWidth;
    const titleWidth = title.getBoundingClientRect().width;
    return Math.max(
        0,
        rootWidth - titleWidth - ROOT_GAP - TITLE_LINK_MIN_GAP - ROOT_GAP,
    );
};

const measureLinks = () => {
    if (!measureRef.value) return;
    const linkNodes = measureRef.value.querySelectorAll("[data-measure-link]");
    linkWidths.value = Array.from(linkNodes).map(
        (node) => node.getBoundingClientRect().width,
    );
    const separator = measureRef.value.querySelector("[data-measure-sep]");
    separatorWidth.value = separator?.getBoundingClientRect().width ?? 0;
};

const updateVisibleCount = () => {
    measureLinks();
    const available = getAvailableNavWidth();
    const total = props.links.length;
    if (total === 0 || linkWidths.value.length !== total) {
        visibleCount.value = total;
        return;
    }
    let nextVisible = total;
    for (let visible = total; visible >= 0; visible--) {
        let required = navWidthFor(visible);
        if (visible < total) {
            required += OVERFLOW_GAP + OVERFLOW_BTN_WIDTH;
        }
        if (required <= available) {
            nextVisible = visible;
            break;
        }
    }
    visibleCount.value = nextVisible;
};

const onOverflowLinkClick = (link: ModuleTitleLink) => {
    menuOpen.value = false;
    link.onClick?.();
};

function onDocumentClick(ev: MouseEvent) {
    const el = menuRoot.value;
    if (!el || el.contains(ev.target as Node)) return;
    menuOpen.value = false;
}

let resizeObserver: ResizeObserver | null = null;

watch(
    () => props.links,
    () => {
        void nextTick(updateVisibleCount);
    },
    { deep: true },
);

watch(overflowLinks, (links) => {
    if (links.length === 0) menuOpen.value = false;
});

onMounted(() => {
    document.addEventListener("click", onDocumentClick);
    resizeObserver = new ResizeObserver(() => {
        updateVisibleCount();
    });
    if (rootRef.value) resizeObserver.observe(rootRef.value);
    void nextTick(updateVisibleCount);
});

onUnmounted(() => {
    document.removeEventListener("click", onDocumentClick);
    resizeObserver?.disconnect();
});
</script>
<template>
    <div
        ref="rootRef"
        class="flex min-w-0 items-center gap-4 border-b border-(--color-border) pb-3"
    >
        <h1 class="m-0 shrink-0 text-md font-semibold text-textSecondary">
            {{ title }}
        </h1>
        <div
            v-if="links.length"
            class="min-w-8 flex-1 shrink"
            aria-hidden="true"
        ></div>
        <div
            v-if="links.length"
            class="flex shrink-0 items-center gap-2 text-sm text-textSecondary"
        >
            <div
                ref="measureRef"
                class="pointer-events-none fixed top-0 -left-[9999px] flex items-center gap-3 text-sm"
                aria-hidden="true"
            >
                <template v-for="(link, index) in links" :key="link.name">
                    <span
                        data-measure-link
                        class="whitespace-nowrap transition-colors"
                        >{{ link.name }}</span
                    >
                    <span
                        v-if="index === 0"
                        data-measure-sep
                        class="text-textSecondary/50"
                        >·</span
                    >
                </template>
            </div>
            <nav
                v-if="visibleLinks.length"
                class="flex min-w-0 items-center gap-3"
            >
                <template
                    v-for="(link, index) in visibleLinks"
                    :key="link.name"
                >
                    <router-link
                        v-if="link.to"
                        :to="linkTo(link)!"
                        class="shrink-0 whitespace-nowrap transition-colors hover:text-textPrimary"
                        >{{ link.name }}</router-link
                    >
                    <button
                        v-else
                        type="button"
                        class="shrink-0 whitespace-nowrap p-0! transition-colors hover:text-textPrimary"
                        @click="link.onClick?.()"
                    >
                        {{ link.name }}
                    </button>
                    <span
                        v-if="index < visibleLinks.length - 1"
                        aria-hidden="true"
                        class="shrink-0 text-textSecondary/50"
                        >·</span
                    >
                </template>
            </nav>
            <div
                v-if="overflowLinks.length"
                ref="menuRoot"
                class="relative shrink-0"
            >
                <button
                    type="button"
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded text-textSecondary transition-colors hover:bg-secondBg hover:text-textPrimary"
                    :aria-expanded="menuOpen"
                    aria-haspopup="true"
                    aria-label="More options"
                    @click.stop="menuOpen = !menuOpen"
                >
                    <MoreVertical :size="20" />
                </button>
                <div
                    v-if="menuOpen"
                    class="absolute right-0 top-full z-20 mt-0.5 min-w-52 rounded-md border border-secondBg bg-firstBg py-1 shadow-lg"
                    role="menu"
                    @click.stop
                >
                    <template v-for="link in overflowLinks" :key="link.name">
                        <router-link
                            v-if="link.to"
                            :to="linkTo(link)!"
                            class="flex w-full items-center whitespace-nowrap px-3 py-2 text-left text-sm text-textPrimary transition-colors hover:bg-secondBg"
                            role="menuitem"
                            @click="menuOpen = false"
                            >{{ link.name }}</router-link
                        >
                        <button
                            v-else
                            type="button"
                            class="flex w-full items-center whitespace-nowrap px-3 py-2 text-left text-sm text-textPrimary transition-colors hover:bg-secondBg"
                            role="menuitem"
                            @click="onOverflowLinkClick(link)"
                        >
                            {{ link.name }}
                        </button>
                    </template>
                </div>
            </div>
        </div>
    </div>
</template>
