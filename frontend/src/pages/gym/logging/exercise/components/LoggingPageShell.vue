<script setup lang="ts">
import { SearchX } from "lucide-vue-next";
import { Button, EmptyState, ErrorState, Skeleton, SkeletonList, useDelayedFlag } from "~/shared/ui";
import LoggingHeader from "./LoggingHeader.vue";

const props = defineProps<{
    pending?: boolean;
    error?: Error | null;
    empty?: boolean;
    emptyMessage?: string;
}>();

const emit = defineEmits<{
    (e: "back"): void;
    (e: "retry"): void;
}>();

// Fast loads show nothing instead of flashing a skeleton.
const showSkeleton = useDelayedFlag(() => !!props.pending);
</script>

<template>
    <div v-if="pending" class="flex w-full flex-col gap-6 self-stretch py-2 lg:py-0">
        <LoggingHeader @back="emit('back')">
            <template #center>
                <div v-if="showSkeleton" class="flex justify-center"><Skeleton width="8rem" /></div>
            </template>
        </LoggingHeader>
        <SkeletonList v-if="showSkeleton" :rows="3" />
    </div>
    <div v-else-if="error" class="flex w-full flex-col gap-4 self-stretch py-2 lg:py-0">
        <LoggingHeader @back="emit('back')" />
        <ErrorState :error="error" title="Couldn't load this workout" @retry="emit('retry')" />
    </div>
    <div v-else-if="empty" class="flex w-full flex-col gap-4 self-stretch py-2 lg:py-0">
        <slot name="empty">
            <EmptyState :icon="SearchX" :title="emptyMessage ?? 'Nothing to show'">
                <template #action>
                    <Button variant="secondary" @click="emit('back')">Back</Button>
                </template>
            </EmptyState>
        </slot>
    </div>
    <div v-else class="flex w-full flex-col gap-4 self-stretch py-2 lg:py-0">
        <slot />
    </div>
</template>
