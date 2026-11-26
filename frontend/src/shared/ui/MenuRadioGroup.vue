<script setup lang="ts" generic="T extends string">
// Tiles in a row, so Left/Right move between them as well as the menu's Up/Down. Picking one keeps
// the menu open so the effect (e.g. a theme) can be seen.
import type { Component } from "vue";
import { DropdownMenuRadioGroup, DropdownMenuRadioItem } from "reka-ui";

const props = defineProps<{ options: { label: string; value: T; icon?: Component }[]; label: string }>();
const model = defineModel<T>({ required: true });

function onUpdate(v: unknown) {
    const picked = props.options.find((o) => o.value === v);
    if (picked) model.value = picked.value;
}

function onArrow(e: KeyboardEvent, step: 1 | -1) {
    const tile = e.currentTarget as HTMLElement;
    const sibling = step === 1 ? tile.nextElementSibling : tile.previousElementSibling;
    if (sibling instanceof HTMLElement) sibling.focus();
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
            @keydown.right.prevent="onArrow($event, 1)"
            @keydown.left.prevent="onArrow($event, -1)"
        >
            <component :is="o.icon" v-if="o.icon" aria-hidden="true" class="size-4" />
            {{ o.label }}
        </DropdownMenuRadioItem>
    </DropdownMenuRadioGroup>
</template>
