import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";
import { dietKeys } from "~/api/diet/keys";
import type { SavedMeal } from "~/types/diet";

/** The saved-meal list, shared (same query key) by the day editor and the range planner. */
export function useSavedMeals() {
    const query = useQuery({
        queryKey: dietKeys.savedMeals.list(),
        queryFn: () =>
            apiClient
                .get<{ saved_meals: SavedMeal[] }>("/diet/meals/saved-meal/all")
                .then((r) => r.data.saved_meals),
    });
    const savedMeals = computed((): SavedMeal[] => query.data.value ?? []);
    return { ...query, savedMeals };
}
