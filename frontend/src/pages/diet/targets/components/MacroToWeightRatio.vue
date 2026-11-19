<script setup lang="ts">
import { computed, ref } from "vue";
import { Button, Field, Input } from "~/shared/ui";

const props = defineProps<{
    label: string;
    weight: number;
}>();

const emit = defineEmits<{ setRatio: [ratio: number] }>();

const startingRatio = computed(() => {
    if (props.label.toLocaleLowerCase() === "protein") return 0.8;
    if (props.label.toLocaleLowerCase() === "fat") return 0.3;
    return 0;
});

const ratio = ref<number | string>(startingRatio.value);
const grams = computed(() => Math.round((Number(ratio.value) || 0) * props.weight));
</script>

<template>
    <div class="flex flex-wrap items-end gap-x-3 gap-y-2">
        <div class="w-32">
            <Field :label="`${label} (g/lb)`">
                <Input v-model="ratio" type="number" inputmode="decimal" />
            </Field>
        </div>
        <Button size="sm" @click="emit('setRatio', Number(ratio) || 0)">
            Apply {{ grams }}g
        </Button>
        <span class="pb-3 text-xs text-fg-muted lg:pb-2"
            >{{ grams }}g at {{ weight }} lbs</span
        >
    </div>
</template>
