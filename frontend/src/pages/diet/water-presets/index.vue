<script setup lang="ts">
import axios from "axios";
import { computed, ref, watch } from "vue";
import type { DrinkSizePreset } from "~/api/tracking/types";
import {
    useCreateDrinkSizePreset,
    useDeleteDrinkSizePreset,
    useDrinkSizePresets,
    useUpdateDrinkSizePreset,
} from "~/api/tracking/queries";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import { useWaterPrefs } from "~/composables/water/useWaterPrefs";

const {
    goalOz,
    displayUnit,
    ozFromDisplayAmount,
    formatVolumeFromOz,
} = useWaterPrefs();

const goalInput = ref("");

const newName = ref("");
const newAmount = ref("");

const editingId = ref<number | null>(null);
const editName = ref("");
const editAmount = ref("");

const { data: presets, isPending, isError, error } = useDrinkSizePresets();
const createMut = useCreateDrinkSizePreset();
const updateMut = useUpdateDrinkSizePreset();
const deleteMut = useDeleteDrinkSizePreset();

const rows = computed(() => presets.value ?? []);

function syncGoalInputFromPrefs() {
    const f = formatVolumeFromOz(goalOz.value);
    goalInput.value = String(f.value);
}

syncGoalInputFromPrefs();

watch(displayUnit, () => {
    const f = formatVolumeFromOz(goalOz.value);
    goalInput.value = String(f.value);
});

function applyGoalFromInput() {
    const n = Number.parseFloat(goalInput.value.replace(/\s/g, ""));
    if (Number.isNaN(n) || n <= 0) {
        toast.push("Enter a positive goal", "error");
        syncGoalInputFromPrefs();
        return;
    }
    const oz = ozFromDisplayAmount(n, displayUnit.value);
    goalOz.value = oz;
    syncGoalInputFromPrefs();
    toast.push("Goal saved", "success");
}

async function createPreset() {
    const name = newName.value.trim();
    const n = Number.parseFloat(newAmount.value.replace(/\s/g, ""));
    if (!name) {
        toast.push("Name is required", "error");
        return;
    }
    if (Number.isNaN(n) || n <= 0) {
        toast.push("Enter a valid amount", "error");
        return;
    }
    const oz = ozFromDisplayAmount(n, displayUnit.value);
    try {
        await createMut.mutateAsync({ name, amountOz: oz });
        toast.push("Size created", "success");
        newName.value = "";
        newAmount.value = "";
    } catch (err: unknown) {
        toast.push(axiosMsg(err), "error");
    }
}

function startEdit(p: DrinkSizePreset) {
    editingId.value = p.ID;
    editName.value = p.name;
    const f = formatVolumeFromOz(p.amount_oz);
    editAmount.value = String(f.value);
}

function cancelEdit() {
    editingId.value = null;
}

async function saveEdit() {
    const id = editingId.value;
    if (id == null) return;
    const name = editName.value.trim();
    const n = Number.parseFloat(editAmount.value.replace(/\s/g, ""));
    if (!name) {
        toast.push("Name is required", "error");
        return;
    }
    if (Number.isNaN(n) || n <= 0) {
        toast.push("Enter a valid amount", "error");
        return;
    }
    const oz = ozFromDisplayAmount(n, displayUnit.value);
    try {
        await updateMut.mutateAsync({ id, name, amountOz: oz });
        toast.push("Updated", "success");
        editingId.value = null;
    } catch (err: unknown) {
        toast.push(axiosMsg(err), "error");
    }
}

async function removePreset(id: number) {
    const ok = await dialogManager.confirm({
        title: "Delete drink size?",
        message: "Quick-add buttons will no longer include this size.",
        confirmText: "Delete",
        cancelText: "Cancel",
    });
    if (!ok) return;
    try {
        await deleteMut.mutateAsync(id);
        toast.push("Deleted", "success");
        if (editingId.value === id) editingId.value = null;
    } catch (err: unknown) {
        toast.push(axiosMsg(err), "error");
    }
}

function axiosMsg(err: unknown): string {
    if (
        axios.isAxiosError(err) &&
        err.response?.data &&
        typeof err.response.data === "object" &&
        "error" in err.response.data
    ) {
        const e0 = (err.response.data as { error?: string }).error;
        if (e0) return e0;
    }
    if (err instanceof Error) return err.message;
    return "Request failed";
}

function displayAmountLabel(amountOz: number): string {
    const f = formatVolumeFromOz(amountOz);
    return `${f.value} ${f.unit}`;
}
</script>

<template>
    <div class="mx-auto flex w-full max-w-3xl flex-col gap-6 pb-10 pt-2">
        <div
            class="flex items-center justify-between gap-4 border-b border-zinc-700 pb-3"
        >
            <h1 class="m-0 text-lg font-semibold text-zinc-100">
                Water sizes & goal
            </h1>
            <router-link
                :to="{ name: 'diet' }"
                class="text-sm text-zinc-400 hover:text-zinc-200"
                >Back to diet</router-link
            >
        </div>
        <section
            class="flex flex-col gap-3 rounded-md border border-zinc-700 bg-zinc-900/50 p-4"
        >
            <h2 class="m-0 text-sm font-semibold text-zinc-200">
                Display & daily goal
            </h2>
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Unit
                <select
                    v-model="displayUnit"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                >
                    <option value="oz">oz</option>
                    <option value="ml">ml</option>
                </select>
            </label>
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Daily goal ({{ displayUnit }})
                <input
                    v-model="goalInput"
                    type="text"
                    inputmode="decimal"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                    @change="applyGoalFromInput"
                    @keyup.enter.prevent="applyGoalFromInput"
                />
            </label>
            <button
                type="button"
                class="self-start rounded-md bg-zinc-700 px-4 py-2 text-sm font-semibold text-zinc-100 hover:bg-zinc-600"
                @click="applyGoalFromInput"
            >
                Save goal
            </button>
            <p class="m-0 text-xs text-zinc-500">
                Amounts are stored in oz; switching units only changes how numbers are shown.
            </p>
        </section>
        <section
            class="flex flex-col gap-3 rounded-md border border-zinc-700 bg-zinc-900/50 p-4"
        >
            <h2 class="m-0 text-sm font-semibold text-zinc-200">
                New drink size
            </h2>
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Label
                <input
                    v-model="newName"
                    type="text"
                    autocomplete="off"
                    placeholder="e.g. Large bottle"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                />
            </label>
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Amount ({{ displayUnit }})
                <input
                    v-model="newAmount"
                    type="text"
                    inputmode="decimal"
                    placeholder="e.g. 24"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                    @keyup.enter.prevent="createPreset"
                />
            </label>
            <button
                type="button"
                class="rounded-md border border-sky-700/60 bg-sky-950/40 px-4 py-2 text-sm font-semibold text-sky-100 hover:bg-sky-950/70 disabled:opacity-50"
                :disabled="createMut.isPending.value"
                @click="createPreset"
            >
                {{ createMut.isPending.value ? "Saving…" : "Add size" }}
            </button>
        </section>
        <section class="flex flex-col gap-2">
            <h2 class="m-0 text-sm font-semibold text-zinc-200">
                Saved sizes
            </h2>
            <div v-if="isError" class="text-sm text-red-400">
                {{ error?.message ?? "Failed to load" }}
            </div>
            <p v-else-if="isPending" class="m-0 text-sm text-zinc-500">
                Loading…
            </p>
            <div
                v-else
                class="overflow-x-auto rounded-md border border-zinc-700"
            >
                <table class="w-full border-collapse text-left text-sm">
                    <thead>
                        <tr class="border-b border-zinc-700 bg-zinc-900">
                            <th class="px-3 py-2 font-medium text-zinc-400">
                                Name
                            </th>
                            <th class="px-3 py-2 font-medium text-zinc-400">
                                Amount
                            </th>
                            <th class="px-3 py-2 font-medium text-zinc-400" />
                        </tr>
                    </thead>
                    <tbody>
                        <tr
                            v-for="p in rows"
                            :key="p.ID"
                            class="border-b border-zinc-800"
                        >
                            <template v-if="editingId === p.ID">
                                <td class="px-3 py-2 align-top">
                                    <input
                                        v-model="editName"
                                        type="text"
                                        class="w-full min-w-[8rem] rounded border border-zinc-600 bg-zinc-950 px-2 py-1 text-zinc-100"
                                    />
                                </td>
                                <td class="px-3 py-2 align-top">
                                    <input
                                        v-model="editAmount"
                                        type="text"
                                        inputmode="decimal"
                                        class="w-full min-w-[6rem] rounded border border-zinc-600 bg-zinc-950 px-2 py-1 text-zinc-100"
                                    />
                                </td>
                                <td class="px-3 py-2 align-top">
                                    <div class="flex flex-wrap gap-2">
                                        <button
                                            type="button"
                                            class="rounded bg-zinc-700 px-2 py-1 text-xs text-white hover:bg-zinc-600 disabled:opacity-50"
                                            :disabled="updateMut.isPending.value"
                                            @click="saveEdit"
                                        >
                                            Save
                                        </button>
                                        <button
                                            type="button"
                                            class="rounded bg-zinc-800 px-2 py-1 text-xs text-zinc-300 hover:bg-zinc-700"
                                            @click="cancelEdit"
                                        >
                                            Cancel
                                        </button>
                                    </div>
                                </td>
                            </template>
                            <template v-else>
                                <td class="px-3 py-2 text-zinc-100">
                                    {{ p.name }}
                                </td>
                                <td class="px-3 py-2 text-zinc-300">
                                    {{ displayAmountLabel(p.amount_oz) }}
                                </td>
                                <td class="px-3 py-2">
                                    <div class="flex flex-wrap gap-2">
                                        <button
                                            type="button"
                                            class="text-xs text-sky-400 hover:text-sky-300"
                                            @click="startEdit(p)"
                                        >
                                            Edit
                                        </button>
                                        <button
                                            type="button"
                                            class="text-xs text-red-400 hover:text-red-300 disabled:opacity-50"
                                            :disabled="deleteMut.isPending.value"
                                            @click="removePreset(p.ID)"
                                        >
                                            Delete
                                        </button>
                                    </div>
                                </td>
                            </template>
                        </tr>
                    </tbody>
                </table>
                <p
                    v-if="!rows.length"
                    class="m-0 px-3 py-4 text-sm text-zinc-500"
                >
                    No sizes yet — add one above.
                </p>
            </div>
        </section>
    </div>
</template>
