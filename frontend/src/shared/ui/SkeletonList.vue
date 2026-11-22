<script setup lang="ts">
import Skeleton from "./Skeleton.vue";

// Matches ListRow height (touch below lg, row from lg) so the swap to real rows doesn't jump.
withDefaults(defineProps<{ rows?: number; leading?: boolean }>(), { rows: 5, leading: true });
</script>

<template>
    <div role="status" aria-busy="true" class="flex flex-col divide-y divide-line">
        <span class="sr-only">Loading</span>
        <div v-for="i in rows" :key="i" class="flex min-h-touch items-center gap-3 px-3 py-2 lg:min-h-row lg:px-4">
            <Skeleton v-if="leading" shape="circle" />
            <div class="flex min-w-0 flex-1 flex-col gap-1.5">
                <Skeleton width="55%" />
                <Skeleton width="30%" height="0.625rem" />
            </div>
        </div>
    </div>
</template>
