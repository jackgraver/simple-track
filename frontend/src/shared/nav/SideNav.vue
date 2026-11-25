<script setup lang="ts">
import { isDevEnvironment } from "~/composables/auth/session";
import AccountMenu from "./AccountMenu.vue";
import NavLink from "./NavLink.vue";
import NavTree from "./NavTree.vue";
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

        <NavTree class="flex-1 overflow-y-auto px-3 pb-4" />

        <div class="border-t border-line p-3">
            <AccountMenu placement="sidenav" />
        </div>
    </aside>
</template>
