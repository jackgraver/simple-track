<script setup lang="ts" generic="T extends string">
// A single choice inside a DropdownMenu (menuitemradio). Laid out as a compact row of tiles, like a
// segmented control; arrow keys still move through them with the rest of the menu. Choosing one
// keeps the menu open so the effect (e.g. a theme) can be seen.
import type { Component } from "vue";
import { DropdownMenuRadioGroup, DropdownMenuRadioItem } from "reka-ui";

defineProps<{ options: { label: string; value: T; icon?: Component }[]; label: string }>();
const model = defineModel<T>({ required: true });

function onUpdate(v: unknown) {
    if (typeof v === "string" && v) model.value = v as T;
}
</script>

<template>
    <DropdownMenuRadioGroup
        :model-value="model"
        :aria-label="label"
        class="grid auto-cols-fr grid-flow-col gap-1 px-1 py-1"
        @update:model-value="onUpdate"
    >
        <DropdownMenuRadioItem
            v-for="o in options"
            :key="o.value"
            :value="o.value"
            class="m-0 flex min-h-touch cursor-pointer flex-col items-center justify-center gap-1 rounded-control border border-line px-2 py-1.5 text-xs text-fg-muted outline-none select-none data-[highlighted]:border-line-strong data-[highlighted]:text-fg data-[state=checked]:border-transparent data-[state=checked]:bg-accent-soft data-[state=checked]:font-medium data-[state=checked]:text-accent lg:min-h-row"
            @select="(e: Event) => e.preventDefault()"
        >
            <component :is="o.icon" v-if="o.icon" aria-hidden="true" class="size-4" />
            {{ o.label }}
        </DropdownMenuRadioItem>
    </DropdownMenuRadioGroup>
</template>
