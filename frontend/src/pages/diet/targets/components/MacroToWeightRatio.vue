<script setup lang="ts">
import { computed, ref } from "vue";

const props = defineProps<{
    label: string;
    weight: number;
    onSetRatio: (ratio: number) => void;
}>();

const startingRatio = computed(() => {
    if (props.label.toLocaleLowerCase() === "protein") return 0.8;
    if (props.label.toLocaleLowerCase() === "fat") return 0.3;
    return 0;
});

const ratio = ref(startingRatio.value);

function onRatioInput(e: Event) {
    ratio.value = Number((e.target as HTMLInputElement).value) || 0;
}
</script>
<template>
    <div class="flex flex-wrap items-end gap-2 pl-0.5">
        <label class="flex min-w-28 flex-col gap-1">
            <span class="text-xs font-medium text-zinc-500"
                >{{ label }} (g/lb)</span
            >
            <input
                :value="ratio"
                type="number"
                min="0"
                step="0.05"
                class="max-w-24 rounded-md border border-zinc-600 bg-zinc-900 px-2 py-1 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                @input="onRatioInput"
            />
        </label>
        <button
            type="button"
            class="rounded-md border border-zinc-600 px-2 py-1 text-xs font-medium text-zinc-300 hover:border-amber-600/60 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-50"
            @click="onSetRatio(ratio)"
        >
            Apply {{ ratio * weight }}g
        </button>
        <span class="pb-1 text-xs text-zinc-600"
            >{{ ratio * weight }}g at {{ weight }} lbs</span
        >
    </div>
</template>
