<script setup lang="ts">
import { useField } from "./fieldContext";

const model = defineModel<string | number>({ required: true });

const props = withDefaults(
    defineProps<{
        type?: string;
        inputmode?: "text" | "decimal" | "numeric" | "email" | "tel" | "search" | "url";
        placeholder?: string;
        size?: "md" | "sm";
        invalid?: boolean;
        disabled?: boolean;
        readonly?: boolean;
        required?: boolean;
        autocomplete?: string;
        name?: string;
        ariaLabel?: string;
    }>(),
    { type: "text", size: "md" },
);

const field = useField();
const isNumeric = () => props.type === "number" || props.inputmode === "decimal" || props.inputmode === "numeric";
function onFocus(e: FocusEvent) {
    if (isNumeric()) (e.target as HTMLInputElement).select();
}

// `!` classes on the <input> beat the global `input, select {}` rule in App.vue (phase 6 removes it).
const control =
    "m-0! min-w-0 flex-1 rounded-none! border-0! bg-transparent! p-0! text-base text-fg! outline-none! placeholder:text-fg-subtle read-only:text-fg-muted lg:text-sm";
</script>

<template>
    <div
        class="flex items-center gap-2 rounded-control border bg-surface-2 px-3 transition-colors duration-150 focus-within:ring-1"
        :class="[
            props.size === 'sm' ? 'min-h-touch lg:min-h-control-sm' : 'min-h-touch lg:min-h-row',
            props.invalid || field?.invalid()
                ? 'border-danger ring-danger focus-within:border-danger'
                : 'border-line-strong ring-accent focus-within:border-accent',
            disabled ? 'cursor-not-allowed opacity-50' : '',
        ]"
    >
        <span v-if="$slots.prefix" class="shrink-0 text-sm text-fg-muted"><slot name="prefix" /></span>
        <input
            :id="field?.id"
            v-model="model"
            :type="type"
            :inputmode="inputmode"
            :placeholder="placeholder"
            :disabled="disabled"
            :readonly="readonly"
            :required="required"
            :autocomplete="autocomplete"
            :name="name"
            :aria-label="ariaLabel"
            :aria-invalid="invalid || field?.invalid() || undefined"
            :aria-describedby="field?.describedBy()"
            :class="[control, isNumeric() ? 'tabular-nums' : '', disabled ? 'cursor-not-allowed' : '']"
            @focus="onFocus"
        />
        <span v-if="$slots.suffix" class="shrink-0 text-sm text-fg-muted"><slot name="suffix" /></span>
    </div>
</template>
