<script setup lang="ts">
import { useRoute } from "vue-router";
import { SlidingHighlight } from "~/shared/ui";
import NavLink from "./NavLink.vue";
import { isItemActive, itemLocation, navGroups } from "./nav";

// The group/sub-page tree shared by the desktop sidebar and the mobile drawer.
// Rows are touch-sized below lg and compact from lg. Hover and active backgrounds are drawn by
// SlidingHighlight (they slide between rows), so rows only set their text colour.
const route = useRoute();
</script>

<template>
    <SlidingHighlight as="nav" aria-label="Main">
        <section v-for="group in navGroups" :key="group.key" class="mb-4">
            <!-- The header links to the group's home (its first item). -->
            <NavLink
                :to="itemLocation(group.items[0]!, route)"
                data-highlight
                class="relative flex min-h-touch items-center gap-2 rounded-control px-2 py-1.5 text-xs font-semibold tracking-wide uppercase lg:min-h-0"
                :class="
                    isItemActive(group.items[0]!, route)
                        ? 'text-fg'
                        : 'text-fg-muted hover:text-fg'
                "
                :aria-current="isItemActive(group.items[0]!, route) ? 'page' : undefined"
            >
                <component :is="group.icon" class="size-4 lg:size-3.5" />
                {{ group.label }}
            </NavLink>
            <ul v-if="group.items.length > 1" class="m-0 mt-0.5 flex list-none flex-col p-0">
                <li v-for="item in group.items.slice(1)" :key="item.to">
                    <NavLink
                        :to="itemLocation(item, route)"
                        data-highlight
                        class="relative flex min-h-touch items-center rounded-control px-2 py-1.5 pl-7.5 text-base lg:block lg:min-h-0 lg:text-sm"
                        :class="
                            isItemActive(item, route)
                                ? 'font-medium text-fg'
                                : 'text-fg-muted hover:text-fg'
                        "
                        :aria-current="isItemActive(item, route) ? 'page' : undefined"
                        >{{ item.label }}</NavLink
                    >
                </li>
            </ul>
        </section>
    </SlidingHighlight>
</template>
