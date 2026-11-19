<script setup lang="ts">
import { computed, ref } from "vue";

/** Half-star rating input/display (0.5–5). Clicking the current value clears it. */
const props = withDefaults(
    defineProps<{ modelValue: number | null; readonly?: boolean; size?: "sm" | "lg" }>(),
    { readonly: false, size: "lg" },
);
const emit = defineEmits<{ "update:modelValue": [value: number | null] }>();

const hover = ref<number | null>(null);
const shown = computed(() => hover.value ?? props.modelValue ?? 0);

function fill(star: number) {
    return Math.max(0, Math.min(1, shown.value - (star - 1))) * 100;
}

function pick(value: number) {
    if (props.readonly) return;
    emit("update:modelValue", value === props.modelValue ? null : value);
}
</script>

<template>
    <div
        class="inline-flex select-none"
        :class="size === 'lg' ? 'text-4xl gap-0.5 lg:text-3xl' : 'text-sm'"
        role="group"
        :aria-label="readonly ? `Rated ${modelValue ?? 0} of 5` : 'Rating'"
        @mouseleave="hover = null"
    >
        <span v-for="star in 5" :key="star" class="relative leading-none">
            <span class="text-line-strong">★</span>
            <span
                class="absolute inset-0 overflow-hidden text-warn"
                :style="{ width: fill(star) + '%' }"
                >★</span
            >
            <template v-if="!readonly">
                <span
                    role="button"
                    tabindex="0"
                    class="absolute inset-y-0 left-0 w-1/2 cursor-pointer focus-visible:outline-2 focus-visible:outline-accent"
                    :aria-label="`${star - 0.5} stars`"
                    @mouseenter="hover = star - 0.5"
                    @click="pick(star - 0.5)"
                    @keydown.enter.prevent="pick(star - 0.5)"
                    @keydown.space.prevent="pick(star - 0.5)"
                ></span>
                <span
                    role="button"
                    tabindex="0"
                    class="absolute inset-y-0 right-0 w-1/2 cursor-pointer focus-visible:outline-2 focus-visible:outline-accent"
                    :aria-label="`${star} stars`"
                    @mouseenter="hover = star"
                    @click="pick(star)"
                    @keydown.enter.prevent="pick(star)"
                    @keydown.space.prevent="pick(star)"
                ></span>
            </template>
        </span>
    </div>
</template>
