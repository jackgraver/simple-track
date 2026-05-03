<script setup lang="ts">
import { ChevronRightIcon } from "lucide-vue-next";
import { computed } from "vue";
import { useRoute } from "vue-router";

type Crumb = {
    key: string;
    label: string;
    to: { path: string };
    isCurrent: boolean;
};

const MAX_VISIBLE = 3;

const route = useRoute();

const crumbs = computed<Crumb[]>(() => {
    const matched = route.matched.filter((r) => r.meta?.breadcrumb);
    return matched.map((r, i) => ({
        key: String(r.name ?? r.path) + ":" + i,
        label: r.meta!.breadcrumb as string,
        to: { path: r.path },
        isCurrent: i === matched.length - 1,
    }));
});

const showBreadcrumbs = computed(() => crumbs.value.length > 1);

type Item = Crumb | { ellipsis: true; key: string };

const visibleItems = computed<Item[]>(() => {
    const all = crumbs.value;
    if (all.length <= MAX_VISIBLE) return all;
    return [
        all[0],
        { ellipsis: true, key: "ellipsis" },
        ...all.slice(-(MAX_VISIBLE - 1)),
    ];
});

const isCrumb = (item: Item): item is Crumb =>
    (item as Crumb).label !== undefined;
</script>

<template>
    <nav v-if="showBreadcrumbs" aria-label="Breadcrumb" class="min-w-0 pl-2 lg:pl-0">
        <ol class="flex min-w-0 items-center gap-1 text-sm">
            <li class="shrink-0">
                <router-link to="/" class="font-semibold tracking-tight text-textPrimary hover:underline underline-offset-4" aria-label="simpletrack home">st</router-link>
            </li>
            <template v-for="item in visibleItems" :key="item.key">
                <li aria-hidden="true" class="shrink-0 text-textSecondary/70 flex items-center">
                    <ChevronRightIcon class="size-3.5" />
                </li>
                <li v-if="!isCrumb(item)" class="shrink-0 text-textSecondary select-none">…</li>
                <li v-else class="min-w-0">
                    <router-link v-if="!item.isCurrent" :to="item.to" class="block max-w-[12ch] truncate text-textSecondary hover:text-textPrimary hover:underline underline-offset-4">{{ item.label }}</router-link>
                    <span v-else aria-current="page" class="block max-w-[14ch] truncate font-medium text-textPrimary">{{ item.label }}</span>
                </li>
            </template>
        </ol>
    </nav>
    <div v-else class="pl-2 lg:pl-0 font-semibold text-lg tracking-tight">
        <router-link to="/" aria-label="simpletrack home">simpletrack</router-link>
    </div>
</template>
