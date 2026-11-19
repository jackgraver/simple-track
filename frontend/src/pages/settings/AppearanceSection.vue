<script setup lang="ts">
import { Section, SegmentedControl } from "~/shared/ui";
import { useTheme, type ThemeMode } from "~/composables/useTheme";

const { mode, setThemeMode } = useTheme();
const options: { label: string; value: ThemeMode }[] = [
    { label: "Dark", value: "dark" },
    { label: "Light", value: "light" },
    { label: "System", value: "system" },
];
</script>

<template>
    <Section title="Appearance">
        <SegmentedControl label="Theme" :options="options" :model-value="mode" @update:model-value="setThemeMode" />
    </Section>
</template>
