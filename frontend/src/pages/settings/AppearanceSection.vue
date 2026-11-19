<script setup lang="ts">
import { SegmentedControl } from "~/shared/ui";
import { useTheme, type ThemeMode } from "~/composables/useTheme";

const { mode, setThemeMode } = useTheme();
const options: { label: string; value: ThemeMode }[] = [
    { label: "Dark", value: "dark" },
    { label: "Light", value: "light" },
    { label: "System", value: "system" },
];
</script>

<template>
    <section class="flex max-w-md flex-col gap-3">
        <h2 class="m-0 text-base font-semibold text-fg">Appearance</h2>
        <SegmentedControl
            label="Theme"
            :options="options"
            :model-value="mode"
            @update:model-value="setThemeMode"
        />
    </section>
</template>
