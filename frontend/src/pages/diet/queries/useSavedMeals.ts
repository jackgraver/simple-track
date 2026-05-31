import { useQuery } from "@tanstack/vue-query";
import { getAllSavedMeals } from "~/api/diet/api";

export const savedMealsQueryKey = ["savedMeals", "all"] as const;

export function useSavedMeals() {
    return useQuery({
        queryKey: savedMealsQueryKey,
        queryFn: getAllSavedMeals,
        staleTime: 1000 * 60 * 2,
    });
}
