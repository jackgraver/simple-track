<script setup lang="ts">
import { getApiErrorMessage } from "~/api/httpError";
import type { Meal } from "~/types/diet";
import { computed, ref, watch } from "vue";
import { Button, Field, Input } from "~/shared/ui";
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
        toast.push(getApiErrorMessage(err, "Failed to log"), "error");
    }
}
</script>

<template>
    <form class="flex flex-col gap-4 text-left" @submit.prevent="submit">
        <p v-if="!isEditing" class="m-0 text-sm text-fg-muted">
            Adds one meal row with totals you enter. Not saved as a reusable
            food.
        </p>
        <Field label="Name" required>
            <Input
                v-model="label"
                autocomplete="off"
                placeholder="e.g. Takeout burger"
            />
        </Field>
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-3">
            <Field label="Calories">
                <Input v-model="calories" type="number" inputmode="decimal" />
            </Field>
            <Field label="Protein (g)">
                <Input v-model="protein" type="number" inputmode="decimal" />
            </Field>
            <Field label="Carbs (g)">
                <Input v-model="carbs" type="number" inputmode="decimal" />
            </Field>
            <Field label="Fat (g)">
                <Input v-model="fat" type="number" inputmode="decimal" />
            </Field>
            <Field label="Fiber (g)">
                <Input v-model="fiber" type="number" inputmode="decimal" />
            </Field>
        </div>
        <Button type="submit" variant="primary" size="lg" block :loading="pending">
            {{ isEditing ? "Save changes" : "Log meal" }}
        </Button>
    </form>
</template>
