<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { activeGroup, itemLocation, navGroups, type NavGroup } from "../nav";
import { useNavMemory } from "./useNavMemory";

const route = useRoute();
const router = useRouter();
const { lastPath } = useNavMemory();

const current = computed(() => activeGroup(route));

function openGroup(group: NavGroup) {
    const home = itemLocation(group.items[0]!, route);
    // Re-tapping the active tab returns to that section's home, like native apps.
    if (current.value?.key === group.key) {
        void router.push(home);
        return;
    }
    void router.push(lastPath.value[group.key] ?? home);
}
</script>

<template>
    <nav
        aria-label="Main"
        class="fixed inset-x-0 bottom-0 z-nav border-t border-line bg-surface/95 pb-safe backdrop-blur-sm"
    >
        <div class="mx-auto flex h-tabbar max-w-md items-stretch">
            <button
                v-for="group in navGroups"
                :key="group.key"
                type="button"
                class="m-0! flex flex-1 flex-col items-center justify-center gap-0.5 rounded-none! bg-transparent p-0! text-2xs shadow-none! transition-colors"
                :class="current?.key === group.key ? 'text-fg' : 'text-fg-muted'"
                :aria-current="current?.key === group.key ? 'page' : undefined"
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
