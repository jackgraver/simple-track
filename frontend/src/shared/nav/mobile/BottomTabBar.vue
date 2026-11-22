<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { activeGroup, itemLocation, navGroups, type NavGroup } from "../nav";
import { queryClient } from "~/api/queryClient";
import { prefetchRoute } from "~/router/prefetch";
import { useNavMemory } from "./useNavMemory";

const route = useRoute();
const router = useRouter();
const { lastPath } = useNavMemory();

const current = computed(() => activeGroup(route));

function targetOf(group: NavGroup) {
    const home = itemLocation(group.items[0]!, route);
    // Re-tapping the active tab returns to that section's home, like native apps.
    if (current.value?.key === group.key) return home;
    return lastPath.value[group.key] ?? home;
}

function openGroup(group: NavGroup) {
    void router.push(targetOf(group));
}

function warm(group: NavGroup) {
    try {
        prefetchRoute(queryClient, router.resolve(targetOf(group)));
    } catch {
        // unresolvable remembered path: nothing to warm
    }
}
</script>

<template>
    <nav
        aria-label="Main"
        class="fixed inset-x-0 bottom-0 z-nav border-t border-line bg-surface/95 pb-safe backdrop-blur-sm"
    >
        <div class="mx-auto flex h-tabbar max-w-md items-stretch">
            <!-- eslint-disable-next-line vue/no-restricted-html-elements -- tab bar item (icon over label), not a button style -->
            <button
                v-for="group in navGroups"
                :key="group.key"
                type="button"
                class="flex flex-1 flex-col items-center justify-center gap-0.5 bg-transparent text-2xs transition-colors"
                :class="current?.key === group.key ? 'text-fg' : 'text-fg-muted'"
                :aria-current="current?.key === group.key ? 'page' : undefined"
                @pointerenter="warm(group)"
                @touchstart.passive="warm(group)"
                @click="openGroup(group)"
            >
                <component
                    :is="group.icon"
                    class="size-5.5"
                    :stroke-width="current?.key === group.key ? 2.4 : 1.8"
                />
                {{ group.label }}
            </button>
        </div>
    </nav>
</template>
