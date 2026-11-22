<script setup lang="ts">
import { computed, ref, useId } from "vue";

/** Half-star rating input/display (0.5–5). Clicking the current value (or Backspace/Delete on the group) clears it. */
const props = withDefaults(
    defineProps<{ modelValue: number | null; readonly?: boolean; size?: "sm" | "lg" }>(),
    { readonly: false, size: "lg" },
);
const emit = defineEmits<{ "update:modelValue": [value: number | null] }>();

const name = useId();
const hover = ref<number | null>(null);
const shown = computed(() => hover.value ?? props.modelValue ?? 0);

function fill(star: number) {
    return Math.max(0, Math.min(1, shown.value - (star - 1))) * 100;
}

function pick(value: number) {
    emit("update:modelValue", value === props.modelValue ? null : value);
}

function clear(e: KeyboardEvent) {
    if (e.key === "Backspace" || e.key === "Delete") {
        e.preventDefault();
        emit("update:modelValue", null);
    }
}

const sizeClass = computed(() => (props.size === "lg" ? "text-4xl gap-0.5 lg:text-3xl" : "text-sm"));
const stars = [1, 2, 3, 4, 5];
</script>

<template>
    <div v-if="readonly" role="img" class="inline-flex select-none" :class="sizeClass" :aria-label="`Rated ${modelValue ?? 0} of 5`">
        <span v-for="star in stars" :key="star" class="relative leading-none">
            <span class="text-line-strong">★</span>
            <span class="absolute inset-0 overflow-hidden text-warn" :style="{ width: fill(star) + '%' }">★</span>
        </span>
    </div>
    <fieldset v-else class="m-0 inline-flex min-w-0 border-0 p-0 select-none" :class="sizeClass" @mouseleave="hover = null" @keydown="clear">
        <legend class="sr-only">Rating</legend>
        <span v-for="star in stars" :key="star" class="relative leading-none">
            <span class="text-line-strong">★</span>
            <span class="absolute inset-0 overflow-hidden text-warn" :style="{ width: fill(star) + '%' }">★</span>
            <label
                v-for="(value, i) in [star - 0.5, star]"
                :key="value"
                class="absolute inset-y-0 w-1/2 cursor-pointer has-focus-visible:outline-2 has-focus-visible:outline-accent"
                :class="i === 0 ? 'left-0' : 'right-0'"
                @mouseenter="hover = value"
            >
                <!-- eslint-disable-next-line vue/no-restricted-html-elements -- sr-only native radio per half star (radio-card pattern, PRIMITIVES.md) -->
                <input
                    type="radio"
                    :name="name"
                    :value="value"
                    :checked="modelValue === value"
                    class="sr-only"
                    @click="pick(value)"
                />
                <span class="sr-only">{{ value }} stars</span>
            </label>
        </span>
    </fieldset>
</template>
