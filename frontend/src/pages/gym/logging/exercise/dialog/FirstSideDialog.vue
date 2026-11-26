<script setup lang="ts">
import { ref } from "vue";
import type { SetSide } from "~/types/workout";
import { Button, SegmentedControl } from "~/shared/ui";

// Unilateral, first log of the exercise: which side was that? Asked on "Log Set 1" so it can't be
// missed and doesn't get in the way of opening the exercise just to look. Preselected with last
// session's first side, so usually one tap. Every later side follows by alternating.
const props = defineProps<{
    side: SetSide;
    onResolve?: (side: SetSide) => void;
}>();

const picked = ref<SetSide>(props.side);
const options: { label: string; value: SetSide }[] = [
    { label: "Left", value: "left" },
    { label: "Right", value: "right" },
];
</script>

<template>
    <div class="flex w-full min-w-0 flex-col gap-4">
        <p class="m-0 text-sm text-fg-muted">Sides alternate from here.</p>
        <SegmentedControl v-model="picked" :options="options" label="First side" />
        <Button variant="primary" size="lg" block @click="onResolve?.(picked)">Confirm</Button>
    </div>
</template>
