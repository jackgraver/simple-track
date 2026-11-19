<script setup lang="ts">
import { MoreVertical } from "lucide-vue-next";
import { Button, IconButton, Menu } from "~/shared/ui";
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { useRouter, type RouteLocationRaw } from "vue-router";

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

const router = useRouter();
const NAV_GAP = 4; // gap-1 between the ghost buttons
const ROOT_GAP = 16;
const TITLE_LINK_MIN_GAP = 32;
const OVERFLOW_GAP = 8;
const OVERFLOW_BTN_WIDTH = 44; // IconButton: size-touch below lg (the larger of its two sizes)

const rootRef = ref<HTMLElement | null>(null);
const measureRef = ref<HTMLElement | null>(null);
const visibleCount = ref(props.links.length);
const linkWidths = ref<number[]>([]);

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
    width += NAV_GAP * (visible - 1);
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

function onOverflowSelect(key: string) {
    const link = overflowLinks.value[Number(key)];
    if (!link) return;
    if (link.to) void router.push(linkTo(link)!);
    else link.onClick?.();
}

const menuItems = computed(() => overflowLinks.value.map((l, i) => ({ key: String(i), label: l.name })));

let resizeObserver: ResizeObserver | null = null;

watch(
    () => props.links,
    () => {
        void nextTick(updateVisibleCount);
    },
    { deep: true },
);

onMounted(() => {
    resizeObserver = new ResizeObserver(() => {
        updateVisibleCount();
    });
    if (rootRef.value) resizeObserver.observe(rootRef.value);
    void nextTick(updateVisibleCount);
});

onUnmounted(() => {
    resizeObserver?.disconnect();
});
</script>
<template>
    <div
        ref="rootRef"
        class="flex min-w-0 items-center gap-4 border-b border-line pb-3"
        :class="links.length === 0 ? 'hidden lg:flex' : ''"
    >
        <!-- On mobile the app header already shows the page title. -->
        <h1 class="m-0 hidden shrink-0 text-xl font-semibold tracking-tight text-fg lg:block">
            {{ title }}
        </h1>
        <div v-if="links.length" class="min-w-8 flex-1 shrink" aria-hidden="true"></div>
        <div v-if="links.length" class="flex shrink-0 items-center gap-2">
            <div
                ref="measureRef"
                class="pointer-events-none fixed top-0 -left-[9999px] flex items-center gap-1"
                aria-hidden="true"
            >
                <Button v-for="link in links" :key="link.name" data-measure-link variant="ghost" size="sm" tabindex="-1">{{
                    link.name
                }}</Button>
            </div>
            <nav v-if="visibleLinks.length" aria-label="Page actions" class="flex min-w-0 items-center gap-1">
                <template v-for="link in visibleLinks" :key="link.name">
                    <Button v-if="link.to" variant="ghost" size="sm" :to="linkTo(link)!">{{ link.name }}</Button>
                    <Button v-else variant="ghost" size="sm" @click="link.onClick?.()">{{ link.name }}</Button>
                </template>
            </nav>
            <Menu v-if="overflowLinks.length" :items="menuItems" @select="onOverflowSelect">
                <IconButton label="More options" size="sm"><MoreVertical /></IconButton>
            </Menu>
        </div>
    </div>
</template>
