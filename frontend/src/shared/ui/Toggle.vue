<script setup lang="ts">
// Sample of the shared/ui wrapper pattern: Reka supplies the behaviour (role=switch, keyboard,
// aria-checked), the look is ours (tokens only). Only shared/ui/ may import "reka-ui".
import { SwitchRoot, SwitchThumb } from "reka-ui";

const model = defineModel<boolean>({ default: false });
defineProps<{ label: string; disabled?: boolean }>();
</script>

<template>
    <label class="flex min-h-touch items-center justify-between gap-3 text-base text-fg lg:min-h-row lg:text-sm">
        <span>{{ label }}</span>
        <SwitchRoot
            v-model="model"
            :disabled="disabled"
            class="relative inline-flex h-6 w-11 shrink-0 items-center rounded-full bg-surface-3 transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-accent"
        >
            <SwitchThumb
                class="block size-5 translate-x-0.5 rounded-full bg-on-solid shadow-raised transition-transform duration-150 data-[state=checked]:translate-x-5.5"
            />
        </SwitchRoot>
    </label>
</template>
