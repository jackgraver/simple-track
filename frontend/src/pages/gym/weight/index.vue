<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useWeightLogs, useSaveWeight } from "~/api/tracking/queries";
import { formatDateLong, parseYmdDateQuery } from "~/utils/dateUtil";
import { toast } from "~/composables/toast/useToast";
import axios from "axios";

function localDateInputValue(d = new Date()) {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
}

const route = useRoute();
const dateStr = ref(
    parseYmdDateQuery(route.query.date) ?? localDateInputValue(),
);
watch(
    () => route.query.date,
    (d) => {
        const p = parseYmdDateQuery(d);
        if (p) dateStr.value = p;
    },
);
const weightInput = ref("");
const { data: logs, isPending, isError, error } = useWeightLogs();
const saveMutation = useSaveWeight();

const rows = computed(() => logs.value ?? []);

const handleSave = async () => {
    const w = Number.parseFloat(weightInput.value.replace(",", "."));
    if (Number.isNaN(w) || w <= 0) {
        toast.push("Enter a valid weight (lbs)", "error");
        return;
    }
    try {
        await saveMutation.mutateAsync({ date: dateStr.value, weightLbs: w });
        toast.push("Saved", "success");
        weightInput.value = "";
    } catch (err: unknown) {
        let msg = "Failed to save";
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
};

const saving = computed(() => saveMutation.isPending.value);
</script>
<template>
    <div class="flex w-full flex-col gap-6 max-w-3xl">
        <div
            class="flex items-center justify-between gap-4 border-b border-(--color-border) pb-3"
        >
            <h1 class="m-0 text-lg font-semibold text-textPrimary">
                Body weight
            </h1>
            <router-link
                :to="{ name: 'gym' }"
                class="text-sm text-textSecondary transition-colors hover:text-textPrimary"
                >Back</router-link
            >
        </div>
        <form
            class="flex flex-col gap-3 rounded-md border border-(--color-border) bg-firstBg p-4"
            @submit.prevent="handleSave"
        >
            <label class="flex flex-col gap-1 text-xs text-textSecondary"
                >Date
                <input
                    v-model="dateStr"
                    type="date"
                    class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                />
            </label>
            <label class="flex flex-col gap-1 text-xs text-textSecondary"
                >Weight (lbs)
                <input
                    v-model="weightInput"
                    type="text"
                    inputmode="decimal"
                    autocomplete="off"
                    placeholder="e.g. 165.5"
                    class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                />
            </label>
            <button
                type="submit"
                class="rounded-md bg-secondBg px-4 py-2 text-sm font-semibold text-textPrimary transition-colors hover:bg-thirdBg disabled:opacity-50"
                :disabled="saving"
            >
                {{ saving ? "Saving…" : "Save" }}
            </button>
        </form>
        <div v-if="isError" class="text-sm text-(--color-cf-red)">
            {{ error?.message ?? "Failed to load" }}
        </div>
        <div v-else-if="isPending" class="text-sm text-textSecondary">
            Loading…
        </div>
        <div v-else class="overflow-x-auto rounded-md border border-(--color-border)">
            <table class="w-full border-collapse text-left text-sm">
                <thead>
                    <tr class="border-b border-(--color-border) bg-secondBg">
                        <th class="px-3 py-2 font-medium text-textSecondary">
                            Date
                        </th>
                        <th class="px-3 py-2 font-medium text-textSecondary">
                            Weight (lbs)
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="r in rows" :key="r.ID" class="border-b border-(--color-border)">
                        <td class="px-3 py-2 text-textPrimary">
                            {{ formatDateLong(r.date) }}
                        </td>
                        <td class="px-3 py-2 text-textPrimary">
                            {{ r.weight_lbs }}
                        </td>
                    </tr>
                </tbody>
            </table>
            <p v-if="!rows.length" class="m-0 px-3 py-4 text-sm text-textSecondary">
                No entries yet.
            </p>
        </div>
    </div>
</template>
