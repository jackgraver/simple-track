<script setup lang="ts">
// For labels, separators or a radio group, compose DropdownMenu directly.
import type { Component } from "vue";
import DropdownMenu from "./DropdownMenu.vue";
import MenuAction from "./MenuAction.vue";

export type MenuItem = {
    key: string;
    label: string;
    icon?: Component;
    tone?: "neutral" | "danger";
    disabled?: boolean;
};

defineProps<{ items: MenuItem[]; align?: "start" | "center" | "end" }>();
const emit = defineEmits<{ select: [key: string] }>();
</script>

<template>
    <DropdownMenu :align="align ?? 'end'">
        <template #trigger><slot /></template>
        <MenuAction
            v-for="item in items"
            :key="item.key"
            :icon="item.icon"
            :tone="item.tone"
            :disabled="item.disabled"
            @select="emit('select', item.key)"
        >
            {{ item.label }}
        </MenuAction>
    </DropdownMenu>
</template>
