<script setup lang="ts">
import { computed } from "vue";
import type { WorkoutStep } from "~/pages/gym/logging/domain/workoutSteps";

// One segment per step, so the bar shows the shape of the session (warmup, lifts, cardio,
// cooldown), not just a percentage. Done = success, next = accent.
const props = defineProps<{ steps: WorkoutStep[]; next: number }>();
const done = computed(() => props.steps.filter((s) => s.done).length);
</script>

<template>
    <div
        role="progressbar"
        :aria-valuenow="done"
        aria-valuemin="0"
        :aria-valuemax="steps.length"
        :aria-label="`${done} of ${steps.length} done`"
        class="flex h-1.5 w-full gap-1"
    >
        <span
            v-for="(s, i) in steps"
            :key="s.key"
            class="h-full min-w-1 flex-1 rounded-full transition-colors duration-150"
            :class="s.done ? 'bg-success' : i === next ? 'bg-accent' : 'bg-surface-3'"
        />
    </div>
</template>
