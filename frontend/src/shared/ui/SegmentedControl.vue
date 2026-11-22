<script setup lang="ts" generic="T extends string">
// 2-5 mutually exclusive options. Reka ToggleGroup (single) gives the roving focus and ARIA.
import { ToggleGroupItem, ToggleGroupRoot } from "reka-ui";

defineProps<{ options: { label: string; value: T }[]; label: string }>();
const model = defineModel<T>({ required: true });

// ToggleGroup lets the active item be un-pressed (emits ""); a segmented control always has one.
function onUpdate(v: unknown) {
    if (typeof v === "string" && v) model.value = v as T;
}
</script>

<template>
    <ToggleGroupRoot
        type="single"
        :model-value="model"
        :aria-label="label"
        class="inline-flex w-full gap-0.5 rounded-control border border-line bg-surface-2 p-0.5"
        @update:model-value="onUpdate"
    >
        <ToggleGroupItem
            v-for="o in options"
            :key="o.value"
            :value="o.value"
            class="min-h-10 flex-1 rounded-control-inner px-3 text-base text-fg-muted transition-colors duration-150 hover:text-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent data-[state=on]:bg-thumb data-[state=on]:text-fg data-[state=on]:shadow-thumb lg:min-h-8 lg:text-sm"
        >
            {{ o.label }}
        </ToggleGroupItem>
    </ToggleGroupRoot>
</template>
