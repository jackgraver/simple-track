<script setup lang="ts" generic="T extends string">
// 2-5 mutually exclusive options. Reka ToggleGroup (single) gives the roving focus and ARIA.
import { ToggleGroupItem, ToggleGroupRoot } from "reka-ui";

type Option = { label: string; value: T };
// `sm`: a slim inline switch (next to a value, not a form row). It looks 30px tall, but each item's
// hit area reaches 44px, so it still meets the mobile touch-target rule.
withDefaults(defineProps<{ options: Option[]; label: string; size?: "md" | "sm" }>(), { size: "md" });
defineSlots<{ option?: (props: { option: Option }) => unknown }>();
const model = defineModel<T>({ required: true });

// ToggleGroup lets the active item be un-pressed (emits ""); a segmented control always has one.
function onUpdate(v: unknown) {
    if (typeof v === "string" && v) model.value = v as T;
}

const sizes = {
    md: "min-h-10 px-3 text-base lg:min-h-8 lg:text-sm",
    sm: "relative min-h-6 px-2 text-sm after:absolute after:inset-x-0 after:-inset-y-2.5",
} as const;
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
            class="flex-1 rounded-control-inner text-fg-muted transition-colors duration-150 hover:text-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent data-[state=on]:bg-thumb data-[state=on]:text-fg data-[state=on]:shadow-thumb"
            :class="sizes[size]"
        >
            <slot name="option" :option="o">{{ o.label }}</slot>
        </ToggleGroupItem>
    </ToggleGroupRoot>
</template>
