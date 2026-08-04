<script setup lang="ts">
import type { LoggedSetWithStatus } from "../../store/useWorkoutStore";
import { Loader, X, RotateCcw } from "lucide-vue-next";

defineProps<{
    loggedSets: LoggedSetWithStatus[];
}>();

const emit = defineEmits<{
    retry: [index: number];
    delete: [index: number];
    edit: [index: number];
}>();
</script>

<template>
    <div v-if="loggedSets.length > 0" class="flex flex-col gap-2">
        <h3
            class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500"
        >
            Logged today
        </h3>
        <ul class="m-0 flex list-none flex-col gap-1.5 p-0">
            <li
                v-for="(set, index) in loggedSets"
                :key="set.tempId || index"
                :class="[
                    'flex items-center justify-between gap-2 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 transition-colors',
                    {
                        'cursor-pointer hover:border-zinc-500 hover:bg-zinc-800':
                            set.status === 'success',
                        'border-red-800 bg-red-950/40': set.status === 'error',
                    },
                ]"
                @click="set.status === 'success' && emit('edit', index)"
            >
                <div
                    class="flex min-w-0 flex-1 items-center justify-between gap-2"
                >
                    <div class="min-w-0 flex flex-col gap-0.5">
                        <span class="whitespace-nowrap">
                            Set {{ index + 1 }}: {{ set.weight }}lbs ×
                            {{ set.reps }}
                        </span>
                        <span
                            v-if="set.weight_setup?.trim()"
                            class="truncate text-xs text-[rgb(130,130,130)]"
                            >{{ set.weight_setup }}</span
                        >
                    </div>
                    <div class="flex shrink-0 items-center gap-2">
                        <div
                            v-if="set.status === 'pending'"
                            class="flex items-center justify-center"
                        >
                            <Loader
                                class="animate-spin text-zinc-400"
                                :size="16"
                            />
                        </div>
                        <div
                            v-else-if="set.status === 'error'"
                            class="flex items-center justify-center"
                        >
                            <button
                                class="flex items-center justify-center rounded p-1 text-red-300 hover:bg-red-950"
                                type="button"
                                title="Retry"
                                @click.stop="emit('retry', index)"
                            >
                                <RotateCcw :size="16" />
                            </button>
                        </div>
                        <button
                            v-if="set.status === 'success'"
                            class="flex items-center justify-center rounded p-1 text-red-300 hover:bg-red-950"
                            type="button"
                            title="Delete set"
                            @click.stop="emit('delete', index)"
                        >
                            <X :size="16" />
                        </button>
                    </div>
                </div>
            </li>
        </ul>
    </div>
</template>
