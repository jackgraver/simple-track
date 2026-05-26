<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
    heightIn: number;
    age: number;
    sex: 'male' | 'female';
}>();

const emit = defineEmits<{
    'update:heightIn': [v: number];
    'update:age': [v: number];
    'update:sex': [v: 'male' | 'female'];
}>();

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
</script>

<template>
    <div class="flex flex-col gap-3">
        <label class="flex flex-col gap-1">
            <span class="text-xs font-medium text-zinc-400">Age</span>
            <input
                :value="age"
                type="number"
                min="1"
                max="130"
                step="1"
                class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                @input="emit('update:age', Number(($event.target as HTMLInputElement).value) || 0)"
            />
        </label>
        <div class="flex flex-col gap-1">
            <span class="text-xs font-medium text-zinc-400">Height</span>
            <div class="flex items-end gap-2">
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
    </div>
</template>
