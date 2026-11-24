<script setup lang="ts">
import { computed, ref } from "vue";
import { Badge, Button, Field, NumberField } from "~/shared/ui";
import type { MachineSetting } from "~/types/workout";
import { overloadReason, type OverloadStatus } from "../domain/liveLogging";
import {
    loadFromPlates,
    plateMismatch,
    type PlateConfig,
    type PlateCounts,
    platesForWeight,
} from "../domain/plates";
import PlatePicker from "../components/PlatePicker.vue";

/** `plates` is {} for exercises that aren't plate-loaded. */
export type WeightSheetResult = { action: "set"; weight: number; plates: PlateCounts } | { action: "undo-progression" };

// Weight and everything that depends on it, in one sheet: reachable from the screen, never in the way.
// Plate-loaded lifts lead with the plates: tapping plates sets the weight. Typing a weight loads the
// nearest plates; one the plates can't make is kept as typed, with the difference shown.
const props = defineProps<{
    weight: number;
    step: number;
    overload: OverloadStatus;
    canUndoProgression: boolean;
    /** Null when the exercise isn't plate-loaded. */
    plateConfig: PlateConfig | null;
    plates: PlateCounts;
    machineSettings: MachineSetting[];
    cues: string;
    onResolve?: (value: WeightSheetResult) => void;
}>();

const draft = ref(props.weight);
const draftPlates = ref<PlateCounts>({ ...props.plates });
const valid = computed(() => Number.isFinite(draft.value) && draft.value >= 0);
const reason = computed(() => overloadReason(props.overload));

const load = computed(() => (props.plateConfig ? loadFromPlates(props.plateConfig, draftPlates.value) : null));
const mismatch = computed(() => (load.value == null ? null : plateMismatch(draft.value, load.value)));

const setPlates = (plates: PlateCounts) => {
    draftPlates.value = plates;
    if (props.plateConfig) draft.value = loadFromPlates(props.plateConfig, plates);
};
const setWeight = (value: number | null | undefined) => {
    draft.value = value ?? 0;
    if (props.plateConfig && valid.value) {
        // From the plates the sheet opened with, not the draft's: typing "1", "14", "145" passes
        // through weights that would strip the bar.
        draftPlates.value = platesForWeight(draft.value, props.plateConfig, props.plates).plates;
    }
};

const done = () => {
    if (!valid.value) return;
    props.onResolve?.({ action: "set", weight: draft.value, plates: draftPlates.value });
};
</script>

<template>
    <div class="flex w-full min-w-0 flex-col gap-5">
        <PlatePicker
            v-if="plateConfig"
            :plates="draftPlates"
            :config="plateConfig"
            @update:plates="setPlates"
        />
        <Field :label="plateConfig ? 'Or type a weight (lb)' : 'Weight (lb)'">
            <NumberField
                :model-value="draft"
                label="weight"
                :step="step"
                :min="0"
                :decimals="1"
                :size="plateConfig ? 'md' : 'lg'"
                unit="lb"
                @update:model-value="setWeight"
            />
        </Field>
        <div v-if="mismatch != null && load != null" class="-mt-3 flex flex-wrap items-center gap-2">
            <p class="m-0 text-sm text-warn">
                Plates make {{ load }} lb ({{ mismatch > 0 ? "+" : "−" }}{{ Math.abs(mismatch) }} lb logged)
            </p>
            <Button variant="ghost" size="sm" @click="draft = load">Use {{ load }} lb</Button>
        </div>
        <div class="flex flex-col gap-1">
            <p class="m-0 text-sm text-fg-muted">{{ reason }}</p>
            <Button
                v-if="canUndoProgression"
                variant="ghost"
                size="sm"
                class="self-start"
                @click="onResolve?.({ action: 'undo-progression' })"
            >
                Undo the increase
            </Button>
        </div>
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
