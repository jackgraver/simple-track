<script setup lang="ts">
import { useRoute } from "vue-router";
import { useAuth } from "~/composables/auth/useAuth";
import { isDevEnvironment } from "~/composables/auth/session";
import { isItemActive, itemLocation, navGroups } from "./nav";

const route = useRoute();
const { getUsername } = useAuth();
</script>

<template>
    <aside
        class="fixed inset-y-0 left-0 z-nav flex w-sidenav flex-col border-r border-line bg-surface"
    >
        <div class="flex items-center gap-2 px-5 pt-5 pb-4">
            <router-link to="/" class="text-lg font-semibold tracking-tight text-fg"
                >simpletrack</router-link
            >
            <span v-if="isDevEnvironment" class="rounded-control bg-surface-3 px-1.5 text-xs text-fg-muted"
                >dev</span
            >
        </div>

        <nav aria-label="Main" class="flex-1 overflow-y-auto px-3 pb-4">
            <section v-for="group in navGroups" :key="group.key" class="mb-4">
                <!-- The header links to the group's home (its first item). -->
                <router-link
                    :to="itemLocation(group.items[0]!, route)"
                    class="flex items-center gap-2 rounded-control px-2 py-1.5 text-xs font-semibold tracking-wide uppercase"
                    :class="
                        isItemActive(group.items[0]!, route)
                            ? 'bg-surface-3 text-fg'
                            : 'text-fg-muted hover:bg-surface-3 hover:text-fg'
                    "
                    :aria-current="isItemActive(group.items[0]!, route) ? 'page' : undefined"
                >
                    <component :is="group.icon" class="size-3.5" />
                    {{ group.label }}
                </router-link>
                <ul v-if="group.items.length > 1" class="m-0 mt-0.5 flex list-none flex-col p-0">
                    <li v-for="item in group.items.slice(1)" :key="item.to">
                        <router-link
                            :to="itemLocation(item, route)"
                            class="block rounded-control px-2 py-1.5 pl-7.5 text-sm"
                            :class="
                                isItemActive(item, route)
                                    ? 'bg-surface-3 font-medium text-fg'
                                    : 'text-fg-muted hover:bg-surface-3 hover:text-fg'
                            "
                            :aria-current="isItemActive(item, route) ? 'page' : undefined"
                            >{{ item.label }}</router-link
                        >
                    </li>
                </ul>
            </section>
        </nav>

        <div class="flex items-center justify-between gap-2 border-t border-line px-5 py-3">
            <router-link
                v-if="getUsername()"
                :to="{ name: 'settings-profile' }"
                class="truncate text-sm text-fg-muted hover:text-fg"
                active-class="text-fg"
                >{{ getUsername() }}</router-link
            >
        </div>
    </aside>
</template>
