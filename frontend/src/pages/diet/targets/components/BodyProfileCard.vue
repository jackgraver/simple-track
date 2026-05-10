<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import type { ActivityLevel } from '~/api/tracking/types';

const props = defineProps<{
    weightLbs: number;
    heightIn: number;
    age: number;
    sex: 'male' | 'female';
    activityLevel: ActivityLevel;
    latestLogWeightLbs: number | null;
}>();

const emit = defineEmits<{
    'update:weightLbs': [v: number];
    'update:heightIn': [v: number];
    'update:age': [v: number];
    'update:sex': [v: 'male' | 'female'];
    'update:activityLevel': [v: ActivityLevel];
}>();

const useMetric = ref(false);

const heightCm = computed({
    get: () => Math.round(props.heightIn * 2.54 * 10) / 10,
    set: (v) => {
        if (typeof v === 'number' && v > 0) emit('update:heightIn', v / 2.54);
    },
});

const feet = computed({
    get: () => Math.floor(props.heightIn / 12),
    set: (ft) => {
        const inches = props.heightIn % 12;
        const n = Number(ft);
        if (!Number.isFinite(n) || n < 0) return;
        emit('update:heightIn', n * 12 + inches);
    },
});

const inchesOnly = computed({
    get: () => Math.round((props.heightIn % 12) * 10) / 10,
    set: (inch) => {
        const f = Math.floor(props.heightIn / 12);
        const n = Number(inch);
        if (!Number.isFinite(n) || n < 0) return;
        emit('update:heightIn', f * 12 + n);
    },
});

const activityOptions: { id: ActivityLevel; label: string }[] = [
    { id: 'sedentary', label: 'Sedentary' },
    { id: 'lightly_active', label: 'Lightly active' },
    { id: 'moderately_active', label: 'Moderate' },
    { id: 'very_active', label: 'Very active' },
    { id: 'extra_active', label: 'Extra active' },
];

watch(
    () => props.latestLogWeightLbs,
    (w) => {
        if (w != null && w > 0 && props.weightLbs <= 0) emit('update:weightLbs', w);
    },
    { immediate: true },
);

function applyLatestWeight() {
    if (props.latestLogWeightLbs != null && props.latestLogWeightLbs > 0) {
        emit('update:weightLbs', props.latestLogWeightLbs);
    }
}
</script>

<template>
    <div class="flex flex-col gap-3">
        <p
            v-if="latestLogWeightLbs != null && latestLogWeightLbs > 0"
            class="m-0 text-xs text-zinc-500"
        >
            Latest logged weight: {{ latestLogWeightLbs }} lb
            <button
                type="button"
                class="ml-1 text-zinc-400 underline underline-offset-2 hover:text-zinc-200"
                @click="applyLatestWeight"
            >
                Use
            </button>
        </p>
        <div class="grid grid-cols-2 gap-3">
            <label class="flex flex-col gap-1">
                <span class="text-xs font-medium text-zinc-400">Weight (lb)</span>
                <input
                    :value="weightLbs"
                    type="number"
                    min="1"
                    step="0.1"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                    @input="
                        emit(
                            'update:weightLbs',
                            Number(($event.target as HTMLInputElement).value) || 0,
                        )
                    "
                />
            </label>
            <label class="flex flex-col gap-1">
                <span class="text-xs font-medium text-zinc-400">Age</span>
                <input
                    :value="age"
                    type="number"
                    min="1"
                    max="130"
                    step="1"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                    @input="
                        emit('update:age', Number(($event.target as HTMLInputElement).value) || 0)
                    "
                />
            </label>
        </div>
        <div class="flex flex-col gap-1">
            <div class="flex items-center gap-2">
                <span class="text-xs font-medium text-zinc-400">Height</span>
                <label class="flex items-center gap-1.5 text-xs text-zinc-500">
                    <input v-model="useMetric" type="checkbox" class="accent-amber-600" />
                    cm
                </label>
            </div>
            <div v-if="useMetric">
                <input
                    :value="heightCm"
                    type="number"
                    min="1"
                    step="0.1"
                    class="w-full rounded-md border border-zinc-600 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                    @input="heightCm = Number(($event.target as HTMLInputElement).value) || 0"
                />
            </div>
            <div v-else class="flex items-end gap-2">
                <label class="flex flex-col gap-1">
                    <span class="text-xs text-zinc-500">Ft</span>
                    <input
                        :value="feet"
                        type="number"
                        min="0"
                        step="1"
                        class="w-20 rounded-md border border-zinc-600 bg-zinc-900 px-2 py-1.5 text-sm text-zinc-100"
                        @input="feet = Number(($event.target as HTMLInputElement).value)"
                    />
                </label>
                <label class="flex flex-col gap-1">
                    <span class="text-xs text-zinc-500">In</span>
                    <input
                        :value="inchesOnly"
                        type="number"
                        min="0"
                        step="0.1"
                        class="w-20 rounded-md border border-zinc-600 bg-zinc-900 px-2 py-1.5 text-sm text-zinc-100"
                        @input="inchesOnly = Number(($event.target as HTMLInputElement).value)"
                    />
                </label>
            </div>
        </div>
        <div class="flex flex-col gap-1.5">
            <span class="text-xs font-medium text-zinc-400">Sex</span>
            <div class="flex gap-2">
                <button
                    type="button"
                    class="rounded-md border px-3 py-1 text-xs font-medium"
                    :class="
                        sex === 'male'
                            ? 'border-amber-600 bg-amber-950/50 text-amber-100'
                            : 'border-zinc-600 text-zinc-300 hover:border-zinc-500'
                    "
                    @click="emit('update:sex', 'male')"
                >
                    Male
                </button>
                <button
                    type="button"
                    class="rounded-md border px-3 py-1 text-xs font-medium"
                    :class="
                        sex === 'female'
                            ? 'border-amber-600 bg-amber-950/50 text-amber-100'
                            : 'border-zinc-600 text-zinc-300 hover:border-zinc-500'
                    "
                    @click="emit('update:sex', 'female')"
                >
                    Female
                </button>
            </div>
        </div>
        <label class="flex flex-col gap-1">
            <span class="text-xs font-medium text-zinc-400">Activity</span>
            <select
                :value="activityLevel"
                class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                @change="
                    emit(
                        'update:activityLevel',
                        ($event.target as HTMLSelectElement).value as ActivityLevel,
                    )
                "
            >
                <option v-for="o in activityOptions" :key="o.id" :value="o.id">
                    {{ o.label }}
                </option>
            </select>
        </label>
    </div>
</template>
