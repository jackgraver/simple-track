<script setup lang="ts">
import { computed } from 'vue';
import { Field, Input, SegmentedControl } from '~/shared/ui';

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

const sexOptions: { label: string; value: 'male' | 'female' }[] = [
    { label: 'Male', value: 'male' },
    { label: 'Female', value: 'female' },
];
</script>

<template>
    <div class="flex flex-col gap-4">
        <Field label="Age">
            <Input
                :model-value="age"
                type="number"
                inputmode="numeric"
                @update:model-value="emit('update:age', Number($event) || 0)"
            />
        </Field>
        <div class="grid grid-cols-2 gap-3">
            <Field label="Height (ft)">
                <Input :model-value="feet" type="number" inputmode="numeric" @update:model-value="feet = Number($event)" />
            </Field>
            <Field label="Height (in)">
                <Input
                    :model-value="inchesOnly"
                    type="number"
                    inputmode="decimal"
                    @update:model-value="inchesOnly = Number($event)"
                />
            </Field>
        </div>
        <Field label="Sex">
            <SegmentedControl
                label="Sex"
                :options="sexOptions"
                :model-value="sex"
                @update:model-value="emit('update:sex', $event)"
            />
        </Field>
    </div>
</template>
