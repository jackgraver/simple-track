<script setup lang="ts">
import type { MobilityLogged } from "~/types/workout";
import LoggingHeader from "./LoggingHeader.vue";
import { ref, watch, computed } from "vue";
import { useLoggingRouteContext } from "../../composables/useLoggingRouteContext";
import { toast } from "~/composables/toast/useToast";
import { useRoute } from "vue-router";

const props = defineProps<{
    loggedPreMobility: MobilityLogged | null;
    loggedPostMobility: MobilityLogged | null;
    savePreMobility: (checked: string[]) => Promise<void>;
    savePostMobility: (checked: string[]) => Promise<void>;
}>();

const route = useRoute();
const { goBackToLogging } = useLoggingRouteContext();

const mobilitySlot = computed<"pre" | "post" | null>(() => {
    const raw = route.params.slot;
    const value = typeof raw === "string" ? raw : "";
    if (value === "pre" || value === "post") return value;
    return null;
});

const effectiveLogged = computed(() =>
    mobilitySlot.value === "post"
        ? props.loggedPostMobility
        : props.loggedPreMobility,
);

const localChecked = ref<Set<string>>(new Set());

watch(
    effectiveLogged,
    (m) => {
        localChecked.value = new Set(m?.checked ?? []);
    },
    { immediate: true },
);

const items = computed(() => effectiveLogged.value?.items ?? []);
const title = computed(() => effectiveLogged.value?.title ?? "");

const doneCount = computed(
    () => items.value.filter((name) => localChecked.value.has(name)).length,
);

const headerProgress = computed(() => {
    const n = items.value.length;
    return n ? `${doneCount.value}/${n}` : "";
});

const toggle = async (name: string) => {
    const next = new Set(localChecked.value);
    if (next.has(name)) {
        next.delete(name);
    } else {
        next.add(name);
    }
    const previous = new Set(localChecked.value);
    localChecked.value = next;
    try {
        if (mobilitySlot.value === "post") {
            await props.savePostMobility([...next]);
        } else {
            await props.savePreMobility([...next]);
        }
    } catch (err: any) {
        localChecked.value = previous;
        toast.push(err.message || "Failed to save", "error");
    }
};

const isChecked = (name: string) => localChecked.value.has(name);
</script>

<template>
    <div class="flex w-full max-w-full flex-col gap-6 px-1 pb-2 box-border">
        <LoggingHeader :title="title" @back="goBackToLogging">
            <template #right>
                <span
                    v-if="headerProgress"
                    class="text-sm font-medium text-zinc-300 tabular-nums"
                    >{{ headerProgress }}</span
                >
            </template>
        </LoggingHeader>
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
            @click="goBackToLogging"
        >
            Done
        </button>
    </div>
</template>
