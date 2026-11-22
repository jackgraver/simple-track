import { queryOptions, useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
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
import axios from "axios";
import { getApiErrorMessage } from "~/api/httpError";
import { useOptimisticMutation } from "~/api/optimistic";
import {
    addGroceryItem,
    completeGroceryItemInList,
    groceryPathRef,
    reconcileGroceryItem,
    removeGroceryItemByClientId,
    restoreGroceryItem,
    type GroceryItemRef,
} from "~/api/tracking/groceryCache";
import {
    addWaterPatch,
    deleteWaterPatch,
    reconcileByDay,
    reconcileWater,
    revertAddWater,
    revertDeleteWater,
    revertUpsertByDay,
    saveStepsPatch,
    saveWeightPatch,
    type AddWaterVars,
    type DeleteWaterVars,
} from "~/api/tracking/patches";
import type { BodyWeightLog, GroceryItem, StepLog, WaterLog } from "~/api/tracking/types";
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
    return useOptimisticMutation<{ date: string; weightLbs: number }, BodyWeightLog>({
        mutationKey: trackingKeys.mutations.saveWeight,
        mutationFn: ({ date, weightLbs }) => saveWeight(date, weightLbs),
        // Upsert by date: same-day saves must reach the server in order.
        scope: (v) => ({ id: `weight:${v.date}` }),
        optimistic: () => [
            {
                queryKey: trackingKeys.weight,
                apply: (old: BodyWeightLog[], vars) => saveWeightPatch(old, vars),
                revert: (cur: BodyWeightLog[], vars, snap: BodyWeightLog[]) => revertUpsertByDay(cur, vars.date, snap),
            },
        ],
        reconcile: (saved, _v, qc, { siblingsPending }) => {
            if (siblingsPending(trackingKeys.weight) > 0) return;
            qc.setQueryData<BodyWeightLog[]>(trackingKeys.weight, (cur) => cur && reconcileByDay(cur, saved));
        },
        invalidate: () => [trackingKeys.weight],
        errorMessage: (err) => getApiErrorMessage(err, "Couldn't save weight"),
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
    return useOptimisticMutation<{ date: string; steps: number }, StepLog>({
        mutationKey: trackingKeys.mutations.saveSteps,
        mutationFn: ({ date, steps }) => saveSteps(date, steps),
        scope: (v) => ({ id: `steps:${v.date}` }),
        optimistic: () => [
            {
                queryKey: trackingKeys.steps,
                apply: (old: StepLog[], vars) => saveStepsPatch(old, vars),
                revert: (cur: StepLog[], vars, snap: StepLog[]) => revertUpsertByDay(cur, vars.date, snap),
            },
        ],
        reconcile: (saved, _v, qc, { siblingsPending }) => {
            if (siblingsPending(trackingKeys.steps) > 0) return;
            qc.setQueryData<StepLog[]>(trackingKeys.steps, (cur) => cur && reconcileByDay(cur, saved));
        },
        invalidate: () => [trackingKeys.steps],
        errorMessage: (err) => getApiErrorMessage(err, "Couldn't save steps"),
    });
}

/** Single source for hooks and route prefetch. */
export const trackingQueries = {
    groceryItems: () =>
        queryOptions({
            queryKey: trackingKeys.groceryItems,
            queryFn: fetchGroceryItems,
        }),
};

export function useGroceryItems() {
    return useQuery(trackingQueries.groceryItems());
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
const GROCERY_CREATE_KEY = ["tracking", "grocery", "create"] as const;
const groceryScope = (clientId: string) => "grocery:" + clientId;

export function useCreateGroceryItem() {
    return useOptimisticMutation<{ client_id: string; name: string }, GroceryItem>({
        mutationKey: GROCERY_CREATE_KEY,
        mutationFn: ({ client_id, name }) => createGroceryItem(name, client_id),
        // Safe to retry: the server de-duplicates by client_id.
        retry: 2,
        // A tick on the row waits for its create (same scope), then completes it by client_id.
        scope: (v) => ({ id: groceryScope(v.client_id) }),
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
    const qc = useQueryClient();
    /** Its create failed: the row never reached the server, so a failed complete must not bring it back. */
    const createFailed = (clientId: string | null | undefined) =>
        !!clientId &&
        qc
            .getMutationCache()
            .findAll({ mutationKey: GROCERY_CREATE_KEY, exact: true, status: "error" })
            .some((m) => (m.state.variables as { client_id?: string } | undefined)?.client_id === clientId);
    return useOptimisticMutation<GroceryItemRef, GroceryItem>({
        mutationKey: ["tracking", "grocery", "complete"],
        mutationFn: (v) => completeGroceryItem(groceryPathRef(v)),
        // PATCH complete is idempotent on the server.
        retry: 2,
        // Rows added in this session have a client_id: wait for their create before completing.
        scope: (v) => ({ id: v.client_id ? groceryScope(v.client_id) : "grocery:id:" + v.id }),
        optimistic: () => [
            {
                queryKey: trackingKeys.groceryItems,
                apply: (old: GroceryItem[], v) => completeGroceryItemInList(old, v),
                revert: (cur: GroceryItem[], v, snapshot) =>
                    createFailed(v.client_id) ? cur : restoreGroceryItem(cur, v, snapshot),
                alwaysRevert: true,
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

/** Log water. Vars need a `client_id` from `newClientId()` (the backend dedupes retries by it). */
export function useSaveWater() {
    return useOptimisticMutation<AddWaterVars, WaterLog>({
        mutationKey: trackingKeys.mutations.addWater,
        mutationFn: ({ date, amountOz, presetId, client_id }) => saveWater(date, amountOz, presetId, client_id),
        // Idempotent through client_id, so a transient failure may be retried safely.
        retry: 1,
        retryDelay: 800,
        optimistic: (v) => [
            {
                queryKey: trackingKeys.water(v.date),
                apply: (old: WaterLog[], vars) => addWaterPatch(old, vars),
                revert: (cur: WaterLog[], vars) => revertAddWater(cur, vars),
            },
        ],
        reconcile: (saved, v, qc) =>
            qc.setQueryData<WaterLog[]>(trackingKeys.water(v.date), (cur) => cur && reconcileWater(cur, v, saved)),
        invalidate: (v) => [trackingKeys.water(v.date)],
        errorMessage: (err) => getApiErrorMessage(err, "Couldn't log water"),
    });
}

export function useDeleteWater() {
    return useOptimisticMutation<DeleteWaterVars, void>({
        mutationKey: trackingKeys.mutations.deleteWater,
        // Already gone (404) is the state we wanted.
        mutationFn: async ({ id }) => {
            try {
                await deleteWater(id);
            } catch (err) {
                if (!(axios.isAxiosError(err) && err.response?.status === 404)) throw err;
            }
        },
        scope: (v) => ({ id: `water:${v.id}` }),
        optimistic: (v) => [
            {
                queryKey: trackingKeys.water(v.date),
                apply: (old: WaterLog[], vars) => deleteWaterPatch(old, vars),
                revert: (cur: WaterLog[], vars, snap: WaterLog[]) => revertDeleteWater(cur, vars, snap),
            },
        ],
        invalidate: (v) => [trackingKeys.water(v.date)],
        errorMessage: (err) => getApiErrorMessage(err, "Couldn't delete water entry"),
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
