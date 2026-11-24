<script setup lang="ts">
import { X } from "lucide-vue-next";
import { watch } from "vue";
import { useRoute } from "vue-router";
import { useAuth } from "~/composables/auth/useAuth";
import { isDevEnvironment } from "~/composables/auth/session";
import { Drawer, IconButton } from "~/shared/ui";
import NavLink from "../NavLink.vue";
import NavTree from "../NavTree.vue";
import { drawerWidth, useNavDrawer } from "./useNavDrawer";

const { open, dragX, dragging, openedBySwipe } = useNavDrawer();
const route = useRoute();
const { getUsername } = useAuth();

// Any navigation (a tapped link, back/forward) closes the drawer.
watch(
    () => route.fullPath,
    () => (open.value = false),
);
</script>

<template>
    <Drawer
        v-model:open="open"
        title="Navigation"
        :offset="dragX"
        :dragging="dragging"
        :width="drawerWidth()"
        :animate-enter="!openedBySwipe"
    >
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

        <div class="flex items-center border-t border-line px-3 py-1">
            <NavLink
                :to="{ name: 'settings-profile' }"
                class="flex min-h-touch min-w-0 flex-1 items-center rounded-control px-2 text-base text-fg-muted hover:bg-surface-3 hover:text-fg"
                active-class="text-fg"
                ><span class="truncate">{{ getUsername() || "Profile" }}</span></NavLink
            >
        </div>
    </Drawer>
</template>
