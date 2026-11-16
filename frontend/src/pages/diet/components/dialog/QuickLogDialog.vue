<script setup lang="ts">
import axios from "axios";
import type { Meal } from "~/types/diet";
import { computed, ref, watch } from "vue";
import { toast } from "~/composables/toast/useToast";
import { useQuickLog } from "~/pages/diet/logmeal/queries/useMealMutations";

const props = defineProps<{
    dateOffset: number;
    editingMeal?: Meal | null;
    onResolve?: (v?: unknown) => void;
    onCancel?: () => void;
}>();

const mutation = useQuickLog();

const label = ref("");
const calories = ref<number | string>("");
const protein = ref<number | string>(0);
const carbs = ref<number | string>(0);
const fat = ref<number | string>(0);
const fiber = ref<number | string>(0);

const isEditing = computed(() => !!(props.editingMeal?.ID ?? 0));
const pending = computed(() => mutation.isPending.value);

function totalsFromMeal(m: Meal) {
    let calories = 0;
    let protein = 0;
    let carbs = 0;
    let fat = 0;
    let fiber = 0;
    for (const item of m.items ?? []) {
        const a = Number(item.amount);
        const f = item.food;
        if (!f) continue;
        calories += (f.calories ?? 0) * a;
        protein += (f.protein ?? 0) * a;
        carbs += (f.carbs ?? 0) * a;
        fat += (f.fat ?? 0) * a;
        fiber += (f.fiber ?? 0) * a;
    }
    return { calories, protein, carbs, fat, fiber };
}

function resetEmptyForm() {
    label.value = "";
    calories.value = "";
    protein.value = 0;
    carbs.value = 0;
    fat.value = 0;
    fiber.value = 0;
}

function hydrateFromEditingMeal(m: Meal | null | undefined) {
    if (!m?.items?.length) {
        resetEmptyForm();
        return;
    }
    label.value = (m.name || "").trim();
    const t = totalsFromMeal(m);
    calories.value = Number.isInteger(t.calories)
        ? t.calories
        : Math.round(t.calories * 100) / 100;
    protein.value =
        Number.isInteger(t.protein) ? t.protein : Math.round(t.protein * 100) / 100;
    carbs.value =
        Number.isInteger(t.carbs) ? t.carbs : Math.round(t.carbs * 100) / 100;
    fat.value =
        Number.isInteger(t.fat) ? t.fat : Math.round(t.fat * 100) / 100;
    fiber.value =
        Number.isInteger(t.fiber)
            ? t.fiber
            : Math.round(t.fiber * 100) / 100;
}

watch(
    () => props.editingMeal,
    (m) => {
        hydrateFromEditingMeal(m ?? null);
    },
    { immediate: true },
);

function num(v: number | string, fallback = 0): number {
    const x = typeof v === "string" ? Number.parseFloat(v) : v;
    if (Number.isNaN(x)) return fallback;
    return x;
}

async function submit() {
    const name = label.value.trim();
    if (!name) {
        toast.push("Enter a label for this meal", "error");
        return;
    }
    try {
        const editId = props.editingMeal?.ID;
        await mutation.mutateAsync({
            name,
            calories: num(calories.value, 0),
            protein: num(protein.value, 0),
            carbs: num(carbs.value, 0),
            fat: num(fat.value, 0),
            fiber: num(fiber.value, 0),
            offset: props.dateOffset ?? 0,
            replace_meal_id:
                editId != null && editId > 0 ? editId : undefined,
        });
        toast.push(isEditing.value ? "Updated" : "Logged", "success");
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
</script>

<template>
    <div class="flex flex-col gap-4 text-left text-zinc-100">
        <p v-if="!isEditing" class="m-0 text-xs text-zinc-400">
            Adds one meal row with totals you enter. Not saved as a reusable
            food.
        </p>
        <label class="flex flex-col gap-1 text-xs text-zinc-400">
            Name
            <input
                v-model="label"
                type="text"
                autocomplete="off"
                class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                placeholder="e.g. Takeout burger"
                @keyup.enter.prevent="submit"
            />
        </label>
        <div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Calories
                <input
                    v-model.number="calories"
                    type="number"
                    min="0"
                    step="any"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                />
            </label>
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Protein (g)
                <input
                    v-model.number="protein"
                    type="number"
                    min="0"
                    step="any"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                />
            </label>
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Carbs (g)
                <input
                    v-model.number="carbs"
                    type="number"
                    min="0"
                    step="any"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                />
            </label>
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Fat (g)
                <input
                    v-model.number="fat"
                    type="number"
                    min="0"
                    step="any"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                />
            </label>
            <label class="flex flex-col gap-1 text-xs text-zinc-400">
                Fiber (g)
                <input
                    v-model.number="fiber"
                    type="number"
                    min="0"
                    step="any"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
                />
            </label>
        </div>
        <button
            type="button"
            class="rounded-md bg-zinc-700 px-4 py-2 text-sm font-semibold text-zinc-100 hover:bg-zinc-600 disabled:opacity-50"
            :disabled="pending"
            @click="submit"
        >
            {{ pending ? "Saving…" : isEditing ? "Save changes" : "Log meal" }}
        </button>
    </div>
</template>
