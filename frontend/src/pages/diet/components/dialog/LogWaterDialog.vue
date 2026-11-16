<script setup lang="ts">
import axios from "axios";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import type { DrinkSizePreset } from "~/api/tracking/types";
import { useDrinkSizePresets, useSaveWater } from "~/api/tracking/queries";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    dateStr: string;
    onResolve?: (v?: unknown) => void;
    onCancel?: () => void;
}>();

const router = useRouter();
const { displayUnit, ozFromDisplayAmount } = useWaterPrefs();
const { data: presets, isPending: presetsPending } = useDrinkSizePresets();
const saveMutation = useSaveWater();

const customInput = ref("");
const savingCustom = computed(() => saveMutation.isPending.value);

async function logPreset(p: DrinkSizePreset) {
    try {
        await saveMutation.mutateAsync({
            date: props.dateStr,
            amountOz: p.amount_oz,
            presetId: p.ID,
        });
        toast.push("Logged", "success");
        props.onResolve?.(true);
    } catch (err: unknown) {
        let msg = "Failed to log";
        if (
            axios.isAxiosError(err) &&
            err.response?.data &&
            typeof err.response.data === "object" &&
            "error" in err.response.data
        ) {
            const e0 = (err.response.data as { error?: string }).error;
            if (e0) msg = e0;
        } else if (err instanceof Error) {
            msg = err.message;
        }
        toast.push(msg, "error");
    }
}

async function submitCustom() {
    const raw = customInput.value.replace(/\s/g, "");
    const n = Number.parseFloat(raw);
    if (Number.isNaN(n) || n <= 0) {
        toast.push("Enter a valid amount", "error");
        return;
    }
    const oz = ozFromDisplayAmount(n, displayUnit.value);
    try {
        await saveMutation.mutateAsync({
            date: props.dateStr,
            amountOz: oz,
        });
        toast.push("Logged", "success");
        customInput.value = "";
        props.onResolve?.(true);
    } catch (err: unknown) {
        let msg = "Failed to log";
        if (
            axios.isAxiosError(err) &&
            err.response?.data &&
            typeof err.response.data === "object" &&
            "error" in err.response.data
        ) {
            const e0 = (err.response.data as { error?: string }).error;
            if (e0) msg = e0;
        } else if (err instanceof Error) {
            msg = err.message;
        }
        toast.push(msg, "error");
    }
}

function manageSizes() {
    props.onCancel?.();
    router.push({ name: "diet-water-presets" });
}
</script>

<template>
    <div class="flex flex-col gap-4 text-left text-zinc-100">
        <p class="m-0 text-xs text-zinc-400">
            Quick-add from a saved size or enter a custom amount.
        </p>
        <div v-if="presetsPending" class="text-sm text-zinc-400">Loading sizes…</div>
        <div v-else class="flex flex-wrap gap-2">
            <button
                v-for="p in presets ?? []"
                :key="p.ID"
                type="button"
                class="rounded-md border border-sky-700/50 bg-sky-950/50 px-3 py-2 text-sm font-medium text-sky-100 hover:bg-sky-950/80 disabled:opacity-50"
                :disabled="savingCustom"
                @click="logPreset(p)"
            >
                {{ p.name }}
            </button>
            <p
                v-if="!(presets ?? []).length"
                class="m-0 w-full text-sm text-zinc-500"
            >
                No drink sizes yet — add some under Manage drink sizes.
            </p>
        </div>
        <div class="flex flex-col gap-2 border-t border-zinc-700 pt-3">
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Custom amount ({{ displayUnit }})
                <input
                    v-model="customInput"
                    type="text"
                    inputmode="decimal"
                    autocomplete="off"
                    placeholder="e.g. 12"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                    @keyup.enter.prevent="submitCustom"
                />
            </label>
            <button
                type="button"
                class="rounded-md bg-zinc-700 px-4 py-2 text-sm font-semibold text-zinc-100 hover:bg-zinc-600 disabled:opacity-50"
                :disabled="savingCustom"
                @click="submitCustom"
            >
                {{ savingCustom ? "Saving…" : "Log custom amount" }}
            </button>
        </div>
        <button
            type="button"
            class="text-sm text-sky-400 underline-offset-2 hover:text-sky-300 hover:underline"
            @click="manageSizes"
        >
            Manage drink sizes
        </button>
    </div>
</template>
