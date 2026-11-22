import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import { computed, type MaybeRefOrGetter, toValue } from "vue";
import {
    completeGroceryItem,
    createGroceryItem,
    createGroceryItems,
    createDrinkSizePreset,
    deleteGroceryItem,
    deleteDrinkSizePreset,
    deleteWater,
    fetchGroceryItems,
    fetchGrocerySuggestions,
    parseGroceryList,
    fetchDrinkSizePresets,
    fetchProfile,
    fetchStepLogs,
    fetchWaterLogs,
    fetchWeightLogs,
    saveProfile,
    saveSteps,
    saveWater,
    saveWeight,
    importWeightCsv,
    updateDrinkSizePreset,
} from "~/api/tracking/api";
import type { BodyWeightLog, GroceryItem } from "~/api/tracking/types";
import { getApiErrorMessage } from "~/api/httpError";
import { useOptimisticMutation } from "~/api/optimistic";
import {
    addGroceryItem,
    completeGroceryItemInList,
    reconcileGroceryItem,
    removeGroceryItemByClientId,
    restoreGroceryItem,
} from "~/api/tracking/groceryCache";
import { trackingKeys } from "~/api/tracking/keys";

export function latestWeightLbs(logs: BodyWeightLog[] | undefined): number | null {
    const first = logs?.[0];
    return first != null && first.weight_lbs > 0 ? first.weight_lbs : null;
}

export function useProfile() {
    return useQuery({
        queryKey: trackingKeys.profile,
        queryFn: fetchProfile,
    });
}

export function useSaveProfile() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: saveProfile,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: trackingKeys.profile });
        },
    });
}

export function useWeightLogs() {
    return useQuery({
        queryKey: trackingKeys.weight,
        queryFn: () => fetchWeightLogs(),
    });
}

export function useSaveWeight() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({
            date,
            weightLbs,
        }: {
            date: string;
            weightLbs: number;
        }) => saveWeight(date, weightLbs),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: trackingKeys.weight });
        },
    });
}

export function useImportWeightCsv() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (csv: string) => importWeightCsv(csv),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: trackingKeys.weight });
        },
    });
}

export function useStepLogs() {
    return useQuery({
        queryKey: trackingKeys.steps,
        queryFn: () => fetchStepLogs(),
    });
}

export function useSaveSteps() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ date, steps }: { date: string; steps: number }) =>
            saveSteps(date, steps),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: trackingKeys.steps });
        },
    });
}

export function useGroceryItems() {
    return useQuery({
        queryKey: trackingKeys.groceryItems,
        queryFn: fetchGroceryItems,
    });
}

export function useGrocerySuggestions(query: MaybeRefOrGetter<string>) {
    const trimmed = computed(() => toValue(query).trim());
    return useQuery({
        queryKey: computed(() =>
            trackingKeys.grocerySuggestions(trimmed.value.toLowerCase()),
        ),
        queryFn: () => fetchGrocerySuggestions(trimmed.value),
        enabled: () => trimmed.value.length > 0,
        staleTime: 5 * 60_000,
    });
}

/** Optimistic add: appends a temp row (client_id), swaps in the server row on success. */
export function useCreateGroceryItem() {
    return useOptimisticMutation<{ client_id: string; name: string }, GroceryItem>({
        mutationKey: ["tracking", "grocery", "create"],
        mutationFn: ({ client_id, name }) => createGroceryItem(name, client_id),
        // Safe to retry: the server de-duplicates by client_id.
        retry: 2,
        optimistic: () => [
            {
                queryKey: trackingKeys.groceryItems,
                apply: (old: GroceryItem[], vars) => addGroceryItem(old, vars),
                revert: (cur: GroceryItem[], vars) => removeGroceryItemByClientId(cur, vars),
            },
        ],
        reconcile: (item, v, qc) =>
            qc.setQueryData<GroceryItem[]>(trackingKeys.groceryItems, (cur) =>
                cur ? reconcileGroceryItem(cur, item, v.client_id) : cur,
            ),
        invalidate: () => [trackingKeys.groceryItems],
        errorMessage: (err) => getApiErrorMessage(err, "Couldn’t add item"),
    });
}

export function useParseGroceryList() {
    return useMutation({
        mutationFn: (text: string) => parseGroceryList(text),
    });
}

export function useCreateGroceryItems() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (names: string[]) => createGroceryItems(names),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: trackingKeys.groceryItems });
        },
    });
}

/** Optimistic complete: the row leaves the list now. Suggestions (built from completed items) refetch on settle. */
export function useCompleteGroceryItem() {
    return useOptimisticMutation<{ id: number }, GroceryItem>({
        mutationKey: ["tracking", "grocery", "complete"],
        mutationFn: ({ id }) => completeGroceryItem(id),
        // PATCH complete is idempotent on the server.
        retry: 2,
        optimistic: () => [
            {
                queryKey: trackingKeys.groceryItems,
                apply: (old: GroceryItem[], v) => completeGroceryItemInList(old, v),
                revert: (cur: GroceryItem[], v, snapshot) => restoreGroceryItem(cur, v, snapshot),
            },
        ],
        invalidate: () => [trackingKeys.groceryItems, trackingKeys.grocerySuggestionsAll],
        errorMessage: (err) => getApiErrorMessage(err, "Couldn’t complete item"),
    });
}

export function useDeleteGroceryItem() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: number) => deleteGroceryItem(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: trackingKeys.groceryItems });
        },
    });
}

export function useWaterLogs(dateStr: MaybeRefOrGetter<string>) {
    return useQuery({
        queryKey: computed(() => trackingKeys.water(toValue(dateStr))),
        queryFn: () => fetchWaterLogs(toValue(dateStr)),
    });
}

export function useSaveWater() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({
            date,
            amountOz,
            presetId,
        }: {
            date: string;
            amountOz: number;
            presetId?: number;
        }) => saveWater(date, amountOz, presetId),
        onSuccess: (_data, vars) => {
            queryClient.invalidateQueries({
                queryKey: trackingKeys.water(vars.date),
            });
        },
    });
}

export function useDeleteWater() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async ({ id, date }: { id: number; date: string }) => {
            await deleteWater(id);
            return date;
        },
        onSuccess: (date) => {
            queryClient.invalidateQueries({
                queryKey: trackingKeys.water(date),
            });
        },
    });
}

export function useDrinkSizePresets() {
    return useQuery({
        queryKey: trackingKeys.waterPresets,
        queryFn: () => fetchDrinkSizePresets(),
    });
}

export function useCreateDrinkSizePreset() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({
            name,
            amountOz,
        }: {
            name: string;
            amountOz: number;
        }) => createDrinkSizePreset(name, amountOz),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: trackingKeys.waterPresets,
            });
        },
    });
}

export function useUpdateDrinkSizePreset() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({
            id,
            name,
            amountOz,
        }: {
            id: number;
            name: string;
            amountOz: number;
        }) => updateDrinkSizePreset(id, name, amountOz),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: trackingKeys.waterPresets,
            });
            queryClient.invalidateQueries({ queryKey: trackingKeys.waterLogsAll });
        },
    });
}

export function useDeleteDrinkSizePreset() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: number) => deleteDrinkSizePreset(id),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: trackingKeys.waterPresets,
            });
            queryClient.invalidateQueries({ queryKey: trackingKeys.waterLogsAll });
        },
    });
}
