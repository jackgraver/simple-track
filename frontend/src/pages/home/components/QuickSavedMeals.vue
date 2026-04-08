<script setup lang="ts">
import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";
import type { Meal } from "~/types/diet";
import { useCreateMeal } from "~/pages/diet/logmeal/queries/useMealMutations";
import { cloneMealForNewLog } from "~/utils/cloneMealForLog";
import { toast } from "~/composables/toast/useToast";
import SimpleMacros from "~/shared/SimpleMacros.vue";

const {
    data,
    isPending: listPending,
    error,
    refetch,
} = useQuery({
    queryKey: ["savedMeals", "all"],
    queryFn: async () => {
        const res = await apiClient.get<unknown>("/diet/meals/meal/all");
        return res.data;
    },
});

const meals = computed((): Meal[] => {
    const v = data.value;
    if (!v) return [];
    if (Array.isArray(v)) return v as Meal[];
    const arr = Object.values(v as object).find((x) => Array.isArray(x));
    return (arr ?? []) as Meal[];
});

const { mutateAsync, isPending: savingMeal } = useCreateMeal();

const quickLog = async (m: Meal) => {
    try {
        await mutateAsync({
            meal: cloneMealForNewLog(m),
            log: true,
        });
        toast.push(`Logged “${m.name}”`, "success");
    } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Failed to log meal";
        toast.push(msg, "error");
    }
};
</script>

<template>
    <div class="rounded-lg border border-zinc-700 bg-zinc-900/50 p-3">
        <div class="mb-2 flex flex-wrap items-baseline justify-between gap-2">
            <h3 class="m-0 text-base font-semibold text-zinc-200">
                Quick log saved meal
            </h3>
            <button
                type="button"
                class="text-xs text-zinc-500 underline hover:text-zinc-300"
                @click="() => refetch()"
            >
                Refresh
            </button>
        </div>
        <p v-if="error" class="m-0 text-sm text-red-400">
            Could not load saved meals.
        </p>
        <p v-else-if="listPending" class="m-0 text-sm text-zinc-500">
            Loading…
        </p>
        <p v-else-if="meals.length === 0" class="m-0 text-sm text-zinc-500">
            No saved meals yet. Build one on the Diet page.
        </p>
        <div v-else class="flex flex-wrap gap-2">
            <button
                v-for="m in meals"
                :key="m.ID"
                type="button"
                class="flex max-w-full flex-col items-start gap-1 rounded-md border border-zinc-600 bg-zinc-800 px-3 py-2 text-left hover:bg-zinc-700 disabled:opacity-50"
                :disabled="savingMeal"
                @click="quickLog(m)"
            >
                <span class="truncate font-medium text-zinc-100">{{
                    m.name
                }}</span>
                <SimpleMacros
                    v-if="m.items?.length"
                    :calories="
                        m.items.reduce(
                            (t, i) => t + (i.food?.calories ?? 0) * i.amount,
                            0,
                        )
                    "
                    :protein="
                        m.items.reduce(
                            (t, i) => t + (i.food?.protein ?? 0) * i.amount,
                            0,
                        )
                    "
                    :fiber="
                        m.items.reduce(
                            (t, i) => t + (i.food?.fiber ?? 0) * i.amount,
                            0,
                        )
                    "
                    font-size="0.8rem"
                />
            </button>
        </div>
    </div>
</template>
