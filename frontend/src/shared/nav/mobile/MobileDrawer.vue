<script setup lang="ts">
import { X } from "lucide-vue-next";
import { watch } from "vue";
import { useRoute } from "vue-router";
import { isDevEnvironment } from "~/composables/auth/session";
import { Drawer, IconButton } from "~/shared/ui";
import AccountMenu from "../AccountMenu.vue";
import NavLink from "../NavLink.vue";
import NavTree from "../NavTree.vue";
import { useNavDrawer } from "./useNavDrawer";

const { panel, open } = useNavDrawer();
const route = useRoute();

// Any navigation (a tapped link, back/forward) closes the drawer.
watch(
    () => route.fullPath,
    () => (open.value = false),
);
</script>

<template>
    <Drawer title="Navigation" :panel="panel">
        <div class="flex items-center justify-between gap-2 py-2 pr-2 pl-5">
            <span class="flex items-center gap-2">
                <NavLink to="/" class="text-lg font-semibold tracking-tight text-fg">simpletrack</NavLink>
                <span v-if="isDevEnvironment" class="rounded-control bg-surface-3 px-1.5 text-xs text-fg-muted"
                    >dev</span
                >
            </span>
            <IconButton label="Close menu" @click="open = false"><X /></IconButton>
        </div>

        <NavTree class="min-h-0 flex-1 overflow-y-auto px-3 pt-2 pb-4" />

        <div class="border-t border-line p-3">
            <AccountMenu />
        </div>
    </Drawer>
</template>
