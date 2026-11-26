<script setup lang="ts">
import { computed, ref } from "vue";
import { Badge, Button, Field, NumberField } from "~/shared/ui";
import type { MachineSetting } from "~/types/workout";
import { overloadReason, type OverloadStatus } from "../domain/liveLogging";

export type WeightSheetResult =
    | { action: "set"; weight: number }
    | { action: "plates"; weight: number };

// Weight and everything that depends on it, in one sheet: reachable from the screen, never in the way.
const props = defineProps<{
    weight: number;
    step: number;
    overload: OverloadStatus;
    /** Stored plate summary for plate-loaded lifts, null otherwise. */
    plates: string | null;
    machineSettings: MachineSetting[];
    cues: string;
    onResolve?: (value: WeightSheetResult) => void;
}>();

const draft = ref(props.weight);
const valid = computed(() => Number.isFinite(draft.value) && draft.value >= 0);
const reason = computed(() => overloadReason(props.overload));

const done = () => {
    if (!valid.value) return;
    props.onResolve?.({ action: "set", weight: draft.value });
};
</script>

<template>
    <div class="flex w-full min-w-0 flex-col gap-5">
        <Field label="Weight (lb)">
            <NumberField v-model="draft" label="weight" :step="step" :min="0" :decimals="1" size="lg" unit="lb" />
        </Field>
        <p class="m-0 text-sm text-fg-muted">{{ reason }}</p>
        <section v-if="plates !== null" class="flex flex-col gap-2">
            <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Plates</p>
            <Button variant="secondary" block @click="onResolve?.({ action: 'plates', weight: draft })">
                <span class="flex w-full min-w-0 items-baseline justify-between gap-2">
                    <span class="truncate font-normal">{{ plates || "Not set" }}</span>
                    <span class="text-sm text-fg-muted">Edit</span>
                </span>
            </Button>
        </section>
        <section class="flex flex-col gap-2">
            <p class="m-0 flex items-center gap-2 text-xs font-semibold tracking-wide text-fg-muted uppercase">
                Machine settings <Badge tone="warn">Preview</Badge>
            </p>
            <dl v-if="machineSettings.length" class="m-0 grid grid-cols-2 gap-2 sm:grid-cols-3">
                <div v-for="s in machineSettings" :key="s.label" class="rounded-control bg-surface-2 px-3 py-2">
                    <dt class="text-2xs tracking-wide text-fg-muted uppercase">{{ s.label }}</dt>
                    <dd class="m-0 text-lg font-semibold text-fg tabular-nums">{{ s.value.trim() || "—" }}</dd>
                </div>
            </dl>
            <p v-else class="m-0 text-sm text-fg-subtle">None yet. They'll be set in the plan, next to load type.</p>
        </section>
        <section v-if="cues" class="flex flex-col gap-1">
            <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Cues</p>
            <p class="m-0 text-sm whitespace-pre-wrap text-fg-muted">{{ cues }}</p>
        </section>
        <Button variant="primary" size="lg" block :disabled="!valid" @click="done">Done</Button>
    </div>
</template>
