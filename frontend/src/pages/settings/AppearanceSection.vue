<script setup lang="ts">
import { Button, Section, SegmentedControl } from "~/shared/ui";
import { toast } from "~/composables/toast/useToast";
import { useTheme, type ThemeMode } from "~/composables/useTheme";

const { mode, setThemeMode } = useTheme();
const options: { label: string; value: ThemeMode }[] = [
    { label: "Dark", value: "dark" },
    { label: "Light", value: "light" },
    { label: "System", value: "system" },
];

// Dev-only demo of the toast action (Retry) and duration, see optimistic-ui F2.
const isDev = import.meta.env.DEV;
function demoRetryToast() {
    toast.push("Couldn’t save (demo)", "error", {
        action: { label: "Retry", onClick: () => toast.push("Retried", "success") },
        duration: 6000,
    });
}
</script>

<template>
    <Section title="Appearance">
        <SegmentedControl label="Theme" :options="options" :model-value="mode" @update:model-value="setThemeMode" />
        <Button v-if="isDev" variant="secondary" @click="demoRetryToast">Demo retry toast (dev)</Button>
    </Section>
</template>
