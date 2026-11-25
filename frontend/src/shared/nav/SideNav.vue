<script setup lang="ts">
import { useRoute } from "vue-router";
import { isDevEnvironment } from "~/composables/auth/session";
import AccountMenu from "./AccountMenu.vue";
import NavLink from "./NavLink.vue";
import { isItemActive, itemLocation, navGroups } from "./nav";

const route = useRoute();
</script>

<template>
    <aside
        class="fixed inset-y-0 left-0 z-nav flex w-sidenav flex-col border-r border-line bg-surface"
    >
        <div class="flex items-center gap-2 px-5 pt-5 pb-4">
            <NavLink to="/" class="text-lg font-semibold tracking-tight text-fg"
                >simpletrack</NavLink
            >
            <span v-if="isDevEnvironment" class="rounded-control bg-surface-3 px-1.5 text-xs text-fg-muted"
                >dev</span
            >
        </div>

        <nav aria-label="Main" class="flex-1 overflow-y-auto px-3 pb-4">
            <section v-for="group in navGroups" :key="group.key" class="mb-4">
                <!-- The header links to the group's home (its first item). -->
                <NavLink
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
                </NavLink>
                <ul v-if="group.items.length > 1" class="m-0 mt-0.5 flex list-none flex-col p-0">
                    <li v-for="item in group.items.slice(1)" :key="item.to">
                        <NavLink
                            :to="itemLocation(item, route)"
                            class="block rounded-control px-2 py-1.5 pl-7.5 text-sm"
                            :class="
                                isItemActive(item, route)
                                    ? 'bg-surface-3 font-medium text-fg'
                                    : 'text-fg-muted hover:bg-surface-3 hover:text-fg'
                            "
                            :aria-current="isItemActive(item, route) ? 'page' : undefined"
                            >{{ item.label }}</NavLink
                        >
                    </li>
                </ul>
            </section>
        </nav>

        <div class="border-t border-line p-3">
            <AccountMenu placement="sidenav" />
        </div>
    </aside>
</template>
