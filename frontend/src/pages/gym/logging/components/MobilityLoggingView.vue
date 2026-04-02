<script setup lang="ts">
import type { MobilityLogged } from "~/types/workout";
import { ArrowLeft } from "lucide-vue-next";
import { ref, watch, computed } from "vue";

const props = defineProps<{
    loggedMobility: MobilityLogged;
}>();

const emit = defineEmits<{
    (e: "back"): void;
    (e: "save", checked: string[]): void;
}>();

const localChecked = ref<Set<string>>(new Set());

watch(
    () => props.loggedMobility,
    (m) => {
        localChecked.value = new Set(m.checked ?? []);
    },
    { immediate: true },
);

const items = computed(() => props.loggedMobility.items);
const title = computed(() => props.loggedMobility.title);

const doneCount = computed(() =>
    items.value.filter((name) => localChecked.value.has(name)).length,
);

const headerProgress = computed(() => {
    const n = items.value.length;
    return n ? `${doneCount.value}/${n}` : "";
});

const toggle = (name: string) => {
    const next = new Set(localChecked.value);
    if (next.has(name)) {
        next.delete(name);
    } else {
        next.add(name);
    }
    localChecked.value = next;
    emit("save", [...next]);
};

const isChecked = (name: string) => localChecked.value.has(name);
</script>

<template>
    <div class="flex w-full max-w-full flex-col gap-6 px-1 pb-2 box-border">
        <div class="grid grid-cols-[auto_1fr_auto] items-center gap-2 border-b border-zinc-700 pb-2 w-full">
            <button
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded border border-zinc-600 bg-transparent text-inherit hover:bg-zinc-800"
                type="button"
                @click="emit('back')"
            >
                <ArrowLeft :size="20" />
            </button>
            <h2 class="m-0 min-w-0 truncate text-center text-lg font-medium">{{ title }}</h2>
            <span v-if="headerProgress" class="shrink-0 text-sm font-medium text-zinc-300 tabular-nums">{{ headerProgress }}</span>
        </div>
        <ul class="m-0 flex list-none flex-col gap-2 p-0">
            <li v-for="item in items" :key="item">
                <label
                    class="flex cursor-pointer items-start gap-3 rounded border border-zinc-600 bg-zinc-900 px-4 py-3 hover:bg-zinc-800"
                >
                    <input
                        class="mt-1 h-4 w-4 shrink-0 rounded border-zinc-500"
                        type="checkbox"
                        :checked="isChecked(item)"
                        @change="toggle(item)"
                    />
                    <span class="text-base leading-snug">{{ item }}</span>
                </label>
            </li>
        </ul>
        <button
            class="w-full rounded bg-green-600 py-3 text-base font-medium text-white hover:bg-green-500"
            type="button"
            @click="emit('back')"
        >
            Done
        </button>
    </div>
</template>
