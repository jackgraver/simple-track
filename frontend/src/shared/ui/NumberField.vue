<script setup lang="ts">
// Reka NumberField: stepping, clamping, keyboard (arrows, PageUp/Down, Home/End) and wheel
// come from the library. The look and the labelled steppers are ours.
import { Minus, Plus } from "lucide-vue-next";
import { NumberFieldDecrement, NumberFieldIncrement, NumberFieldInput, NumberFieldRoot } from "reka-ui";
import { computed } from "vue";
import { useField } from "./fieldContext";

const model = defineModel<number | null | undefined>({ default: null });

const props = withDefaults(
    defineProps<{
        /** Accessible name when there is no surrounding Field, and the noun in "Decrease <label>". */
        label?: string;
        step?: number;
        min?: number;
        max?: number;
        /** Max fraction digits shown. Default 2. */
        decimals?: number;
        /** Unit after the number ("kg", "reps"). */
        unit?: string;
        placeholder?: string;
        size?: "md" | "lg";
        invalid?: boolean;
        disabled?: boolean;
    }>(),
    { step: 1, decimals: 2, size: "md" },
);

const field = useField();
const noun = computed(() => props.label ?? "value");
const formatOptions = computed(() => ({ maximumFractionDigits: props.decimals, useGrouping: false }));

const stepper =
    "m-0! flex size-touch shrink-0 cursor-pointer items-center justify-center rounded-control! border! border-line-strong! bg-surface-2 p-0! text-fg shadow-none! transition-colors duration-100 select-none enabled:hover:bg-surface-3 enabled:active:bg-line focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:cursor-not-allowed disabled:opacity-50 lg:size-row";
</script>

<template>
    <NumberFieldRoot
        :id="field?.id"
        v-model="model"
        :step="step"
        :min="min"
        :max="max"
        :format-options="formatOptions"
        :disabled="disabled"
        :aria-label="field ? undefined : label"
        class="flex items-center gap-2"
    >
        <NumberFieldDecrement :class="stepper" :aria-label="`Decrease ${noun}`" :title="`Decrease ${noun}`">
            <Minus aria-hidden="true" class="size-5" />
        </NumberFieldDecrement>
        <div
            class="flex min-w-0 flex-1 items-center justify-center gap-2 rounded-control border bg-surface-2 px-3 transition-colors duration-150 focus-within:ring-1"
            :class="[
                'min-h-touch lg:min-h-row',
                props.invalid || field?.invalid()
                    ? 'border-danger ring-danger focus-within:border-danger'
                    : 'border-line-strong ring-accent focus-within:border-accent',
                disabled ? 'cursor-not-allowed opacity-50' : '',
            ]"
        >
            <NumberFieldInput
                inputmode="decimal"
                :placeholder="placeholder"
                :aria-invalid="props.invalid || field?.invalid() || undefined"
                :aria-describedby="field?.describedBy()"
                class="m-0! w-full min-w-0 flex-1 rounded-none! border-0! bg-transparent! p-0! text-center text-fg! tabular-nums outline-none! placeholder:text-fg-subtle"
                :class="size === 'lg' ? 'text-lg font-semibold' : 'text-base lg:text-sm'"
                @focus="($event.target as HTMLInputElement).select()"
            />
            <span v-if="unit" class="shrink-0 text-sm text-fg-muted">{{ unit }}</span>
        </div>
        <NumberFieldIncrement :class="stepper" :aria-label="`Increase ${noun}`" :title="`Increase ${noun}`">
            <Plus aria-hidden="true" class="size-5" />
        </NumberFieldIncrement>
    </NumberFieldRoot>
</template>
