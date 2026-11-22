<script setup lang="ts">
import { newClientId } from "~/api/ids";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import type { DrinkSizePreset } from "~/api/tracking/types";
import { useDrinkSizePresets, useSaveWater } from "~/api/tracking/queries";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { Button, EmptyState, Field, Input, Skeleton, ErrorState } from "~/shared/ui";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    dateStr: string;
    onResolve?: (v?: unknown) => void;
    onCancel?: () => void;
}>();

const router = useRouter();
const { displayUnit, ozFromDisplayAmount } = useWaterPrefs();
const {
    data: presets,
    isPending: presetsPending,
    isError: presetsError,
    error: presetsQueryError,
    refetch: refetchPresets,
} = useDrinkSizePresets();
const saveMutation = useSaveWater();

const customInput = ref("");
const savingCustom = computed(() => saveMutation.isPending.value);

function logPreset(p: DrinkSizePreset) {
    saveMutation.mutate({
        client_id: newClientId(),
        date: props.dateStr,
        amountOz: p.amount_oz,
        presetId: p.ID,
        preset: p,
    });
    toast.push("Logged", "success");
    props.onResolve?.(true);
}

function submitCustom() {
    const raw = customInput.value.replace(/\s/g, "");
    const n = Number.parseFloat(raw);
    if (Number.isNaN(n) || n <= 0) {
        toast.push("Enter a valid amount", "error");
        return;
    }
    const oz = ozFromDisplayAmount(n, displayUnit.value);
    saveMutation.mutate({ client_id: newClientId(), date: props.dateStr, amountOz: oz });
    toast.push("Logged", "success");
    customInput.value = "";
    props.onResolve?.(true);
}

function manageSizes() {
    props.onCancel?.();
    router.push({ name: "diet-water-presets" });
}
</script>

<template>
    <div class="flex flex-col gap-4 text-left">
        <p class="m-0 text-sm text-fg-muted">
            Quick-add from a saved size or enter a custom amount.
        </p>
        <ErrorState
            v-if="presetsError"
            compact
            title="Couldn't load drink sizes"
            :error="presetsQueryError"
            @retry="refetchPresets()"
        />
        <div v-else-if="presetsPending" class="flex flex-wrap gap-2">
            <Skeleton v-for="n in 3" :key="n" shape="block" width="6rem" height="2.75rem" />
        </div>
        <div v-else-if="(presets ?? []).length" class="flex flex-wrap gap-2">
            <Button
                v-for="p in presets ?? []"
                :key="p.ID"
                variant="secondary"
                :disabled="savingCustom"
                @click="logPreset(p)"
            >
                {{ p.name }}
            </Button>
        </div>
        <EmptyState
            v-else
            title="No drink sizes yet"
            description="Add some under Manage drink sizes."
        />
        <form
            class="flex flex-col gap-3 border-t border-line pt-4"
            @submit.prevent="submitCustom"
        >
            <Field :label="`Custom amount (${displayUnit})`">
                <Input
                    v-model="customInput"
                    inputmode="decimal"
                    autocomplete="off"
                    placeholder="e.g. 12"
                />
            </Field>
            <Button type="submit" variant="primary" :loading="savingCustom">
                Log custom amount
            </Button>
        </form>
        <div>
            <Button variant="ghost" size="sm" @click="manageSizes">
                Manage drink sizes
            </Button>
        </div>
    </div>
</template>
