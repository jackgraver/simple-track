<script setup lang="ts">
import { useRouter, type RouteLocationRaw } from "vue-router";
import { queryClient } from "~/api/queryClient";
import { prefetchRoute } from "~/router/prefetch";

/**
 * `<router-link>` that warms the target route's `meta.queries` on hover, touch and keyboard focus,
 * so the page usually renders from cache. Attributes and slots pass through to the link.
 */
const props = defineProps<{ to: RouteLocationRaw }>();
const router = useRouter();

function warm() {
    try {
        prefetchRoute(queryClient, router.resolve(props.to));
    } catch {
        // an unresolvable target is the link's problem, not prefetch's
    }
}
</script>

<template>
    <router-link :to="to" @pointerenter="warm" @touchstart.passive="warm" @focus="warm">
        <slot />
    </router-link>
</template>
