<script setup lang="ts">
import { computed } from "vue";
import { X } from "lucide-vue-next";
import { Button } from "~/shared/ui";
import {
    addPlate,
    AVAILABLE_PLATES_LBS,
    loadFromPlates,
    type PlateConfig,
    type PlateCounts,
} from "../domain/plates";
import PlateBar from "./PlateBar.vue";

// Tap a plate to load it, tap a loaded plate to take one off. The total follows.
const props = defineProps<{ config: PlateConfig }>();
const plates = defineModel<PlateCounts>("plates", { required: true });

const perSide = computed(() => props.config.mode === "per_side");
const total = computed(() => loadFromPlates(props.config, plates.value));
const loaded = computed(() =>
    Object.entries(plates.value)
        .filter(([, n]) => n > 0)
        .map(([lbs, n]) => ({ lbs: Number(lbs), n }))
        .sort((a, b) => b.lbs - a.lbs),
);
const baseLine = computed(() => {
    const where = perSide.value ? "per side" : "in total";
    return props.config.base > 0 ? `Plates ${where}, on ${props.config.base} lb` : `Plates ${where}`;
});

const change = (lbs: number, delta: number) => {
    plates.value = addPlate(plates.value, lbs, delta);
};
</script>

<template>
    <section class="flex flex-col gap-3" aria-label="Plates">
        <div class="flex items-center justify-between gap-3 rounded-card bg-surface-2 px-3 py-3">
            <div class="flex min-w-0 flex-col">
                <span class="text-2xl font-semibold text-fg tabular-nums">{{ total }} lb</span>
                <span class="text-sm text-fg-muted">{{ baseLine }}</span>
            </div>
            <PlateBar v-if="perSide" :plates="plates" size="md" />
        </div>

        <div class="flex flex-col gap-2">
            <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">
                {{ perSide ? "Add a plate each side" : "Add a plate" }}
            </p>
            <div class="grid grid-cols-3 gap-2 lg:grid-cols-6">
                <Button
                    v-for="lbs in AVAILABLE_PLATES_LBS"
                    :key="lbs"
                    variant="secondary"
                    :aria-label="`Add a ${lbs} lb plate${perSide ? ' each side' : ''}`"
                    @click="change(lbs, 1)"
                >
                    <span class="tabular-nums">+{{ lbs }}</span>
                </Button>
            </div>
        </div>

        <div class="flex flex-col gap-2">
            <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">
                {{ perSide ? "On each side" : "Loaded" }}
            </p>
            <div v-if="loaded.length" class="flex flex-wrap gap-2">
                <Button
                    v-for="p in loaded"
                    :key="p.lbs"
                    variant="secondary"
                    size="sm"
                    :aria-label="`Remove a ${p.lbs} lb plate${perSide ? ' each side' : ''} (${p.n} loaded)`"
                    @click="change(p.lbs, -1)"
                >
                    <span class="tabular-nums">{{ p.n > 1 ? `${p.n} × ${p.lbs}` : p.lbs }}</span>
                    <X class="size-4 text-fg-subtle" aria-hidden="true" />
                </Button>
                <Button variant="ghost" size="sm" @click="plates = {}">Clear</Button>
            </div>
            <p v-else class="m-0 text-sm text-fg-subtle">
                {{ config.base > 0 ? `Empty bar (${config.base} lb)` : "No plates" }}
            </p>
        </div>
    </section>
</template>
