<script setup lang="ts">
// Selection or completion in lists. Reka Checkbox gives role=checkbox, keyboard and indeterminate.
import { CheckboxIndicator, CheckboxRoot } from "reka-ui";
import { Check, Minus } from "lucide-vue-next";

const model = defineModel<boolean | "indeterminate">({ default: false });
defineProps<{ label: string; disabled?: boolean }>();
</script>

<template>
    <label
        class="flex min-h-touch items-center gap-3 text-base text-fg lg:min-h-row lg:text-sm"
        :class="disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'"
    >
        <CheckboxRoot
            v-model="model"
            :disabled="disabled"
            class="m-0! flex size-5 shrink-0 items-center justify-center rounded-control! border! border-line-strong! bg-surface-2 p-0! shadow-none! transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent data-[state=checked]:border-accent! data-[state=checked]:bg-accent data-[state=indeterminate]:border-accent! data-[state=indeterminate]:bg-accent"
        >
            <CheckboxIndicator class="flex text-fg-inverse">
                <Minus v-if="model === 'indeterminate'" aria-hidden="true" class="size-3.5" />
                <Check v-else aria-hidden="true" class="size-3.5" />
            </CheckboxIndicator>
        </CheckboxRoot>
        <span><slot>{{ label }}</slot></span>
    </label>
</template>
