<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from "vue";
import { useField } from "./fieldContext";

const model = defineModel<string>({ required: true });
const props = withDefaults(
    defineProps<{
        placeholder?: string;
        rows?: number;
        invalid?: boolean;
        disabled?: boolean;
        readonly?: boolean;
        required?: boolean;
        ariaLabel?: string;
    }>(),
    { rows: 3 },
);

const field = useField();
const el = ref<HTMLTextAreaElement>();
// Auto-grow: reset, then fit to content.
function grow() {
    if (!el.value) return;
    el.value.style.height = "auto";
    el.value.style.height = `${el.value.scrollHeight}px`;
}
onMounted(grow);
watch(model, () => nextTick(grow));
</script>

<template>
    <textarea
        :id="field?.id"
        ref="el"
        v-model="model"
        :rows="rows"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :required="required"
        :aria-label="ariaLabel"
        :aria-invalid="props.invalid || field?.invalid() || undefined"
        :aria-describedby="field?.describedBy()"
        class="m-0! block min-h-touch w-full resize-none rounded-control! border! bg-surface-2! px-3! py-2! text-base text-fg! outline-none! transition-colors duration-150 placeholder:text-fg-subtle focus:ring-1 disabled:cursor-not-allowed disabled:opacity-50 read-only:text-fg-muted lg:text-sm"
        :class="
            props.invalid || field?.invalid()
                ? 'border-danger! ring-danger focus:border-danger!'
                : 'border-line-strong! ring-accent focus:border-accent!'
        "
    ></textarea>
</template>
