<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";
import type {
    MealItem,
    PlannedMeal,
    SavedMeal,
    SavedMealItem,
} from "~/types/diet";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
import {
    useAddPlannedFromSaved,
    useDeletePlannedMeal,
} from "~/pages/home/queries/useMealMutations";
import { toast } from "~/composables/toast/useToast";
import SimpleMacros from "~/shared/SimpleMacros.vue";

function macroTotals(items: (MealItem | SavedMealItem)[] | undefined): {
    calories: number;
    protein: number;
    fiber: number;
    carbs: number;
} {
    if (!items?.length) return { calories: 0, protein: 0, fiber: 0, carbs: 0 };
    return items.reduce(
        (acc, i) => ({
            calories: acc.calories + (i.food?.calories ?? 0) * i.amount,
            protein: acc.protein + (i.food?.protein ?? 0) * i.amount,
            fiber: acc.fiber + (i.food?.fiber ?? 0) * i.amount,
            carbs: acc.carbs + (i.food?.carbs ?? 0) * i.amount,
        }),
        { calories: 0, protein: 0, fiber: 0, carbs: 0 },
    );
}

const route = useRoute();
const router = useRouter();

const dateOffset = computed(() => {
    const o = Number(route.query.offset);
    return Number.isFinite(o) ? o : 0;
});

const {
    data: dayData,
    isPending: dayPending,
    error: dayError,
} = useDietLogsToday(dateOffset);

const plannedMeals = computed((): PlannedMeal[] => {
    return dayData.value?.day.plannedMeals ?? [];
});

const {
    data: savedRaw,
    isPending: savedPending,
    error: savedError,
} = useQuery({
    queryKey: ["savedMeals", "all"],
    queryFn: async () => {
        const res = await apiClient.get<unknown>("/diet/meals/saved-meal/all");
        return res.data;
    },
});

const savedMeals = computed((): SavedMeal[] => {
    const v = savedRaw.value;
    if (!v) return [];
    if (Array.isArray(v)) return v as SavedMeal[];
    const arr = Object.values(v as object).find((x) => Array.isArray(x));
    return (arr ?? []) as SavedMeal[];
});

const deletePlanned = useDeletePlannedMeal(dateOffset);
const addFromSaved = useAddPlannedFromSaved(dateOffset);

const isRemoving = computed(() => deletePlanned.isPending.value);
const isAdding = computed(() => addFromSaved.isPending.value);

const removePlanned = async (pm: PlannedMeal) => {
    try {
        await deletePlanned.mutateAsync(pm.ID);
        toast.push("Removed from planned", "success");
    } catch {
        toast.push("Could not remove planned meal", "error");
    }
};

const addSaved = async (sm: SavedMeal) => {
    try {
        await addFromSaved.mutateAsync(sm.ID);
        toast.push(`Added “${sm.name}” to planned`, "success");
    } catch {
        toast.push("Could not add to planned", "error");
    }
};

const goBack = () => {
    router.push({ name: "diet" });
};
</script>

<template>
    <div class="mx-auto flex max-w-3xl flex-col gap-6 pb-8 pt-2">
        <div class="flex flex-wrap items-center gap-3">
            <button
                type="button"
                class="text-sm text-zinc-400 underline hover:text-zinc-200"
                @click="goBack"
            >
                ← Diet
            </button>
            <h1 class="m-0 text-xl font-semibold tracking-tight text-zinc-100">
                Edit planned meals
            </h1>
        </div>
        <div v-if="dayPending" class="text-zinc-500">Loading day…</div>
        <div v-else-if="dayError" class="text-red-400">
            {{ dayError.message }}
        </div>
        <section v-else class="flex flex-col gap-3">
            <h2 class="m-0 text-base font-semibold text-zinc-200">
                Currently planned
            </h2>
            <p
                v-if="plannedMeals.length === 0"
                class="m-0 text-sm text-zinc-500"
            >
                Nothing planned yet. Add a saved meal below.
            </p>
            <ul v-else class="m-0 flex list-none flex-col gap-2 p-0">
                <li
                    v-for="pm in plannedMeals"
                    :key="pm.ID"
                    class="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-zinc-700 bg-zinc-900/50 px-3 py-2"
                >
                    <div class="min-w-0 flex-1">
                        <p class="m-0 font-medium text-zinc-100">
                            {{ pm.meal.name }}
                        </p>
                        <SimpleMacros
                            v-if="pm.meal.items?.length"
                            class="mt-1"
                            font-size="0.8rem"
                            v-bind="macroTotals(pm.meal.items)"
                        />
                    </div>
                    <button
                        type="button"
                        class="shrink-0 rounded-md border border-red-900/60 bg-red-950/40 px-2 py-1 text-sm font-medium text-red-200 hover:bg-red-950/70 disabled:opacity-50"
                        :disabled="isRemoving"
                        @click="removePlanned(pm)"
                    >
                        Remove
                    </button>
                </li>
            </ul>
        </section>
        <section class="flex flex-col gap-3">
            <h2 class="m-0 text-base font-semibold text-zinc-200">
                Add from saved meals
            </h2>
            <p v-if="savedError" class="m-0 text-sm text-red-400">
                Could not load saved meals.
            </p>
            <p v-else-if="savedPending" class="m-0 text-sm text-zinc-500">
                Loading saved meals…
            </p>
            <p
                v-else-if="savedMeals.length === 0"
                class="m-0 text-sm text-zinc-500"
            >
                No saved meals yet. Create one from the Diet page.
            </p>
            <div v-else class="flex flex-wrap gap-2">
                <button
                    v-for="sm in savedMeals"
                    :key="sm.ID"
                    type="button"
                    class="flex max-w-full flex-col items-start gap-1 rounded-md border border-zinc-600 bg-zinc-800 px-3 py-2 text-left hover:bg-zinc-700 disabled:opacity-50"
                    :disabled="isAdding"
                    @click="addSaved(sm)"
                >
                    <span class="font-medium text-zinc-100">{{ sm.name }}</span>
                    <SimpleMacros
                        v-if="sm.items?.length"
                        font-size="0.8rem"
                        v-bind="macroTotals(sm.items)"
                    />
                </button>
            </div>
        </section>
    </div>
</template>
