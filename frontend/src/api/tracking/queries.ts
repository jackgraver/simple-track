import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { isAuthenticated } from "~/composables/auth/session";
import {
    createDrinkSizePreset,
    deleteDrinkSizePreset,
    deleteWater,
    fetchDrinkSizePresets,
    fetchMissedTracking,
    fetchStepLogs,
    fetchWaterLogs,
    fetchWeightLogs,
    saveSteps,
    saveWater,
    saveWeight,
    updateDrinkSizePreset,
} from "~/api/tracking/api";
import { trackingKeys } from "~/api/tracking/keys";

export function useWeightLogs() {
    return useQuery({
        queryKey: trackingKeys.weight,
        queryFn: () => fetchWeightLogs(),
    });
}

export function useMissedTracking() {
    return useQuery({
        queryKey: trackingKeys.missed,
        queryFn: () => fetchMissedTracking(),
        staleTime: 5 * 60_000,
        enabled: computed(() => isAuthenticated.value),
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
            queryClient.invalidateQueries({ queryKey: trackingKeys.missed });
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
            queryClient.invalidateQueries({ queryKey: trackingKeys.missed });
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
            queryClient.invalidateQueries({ queryKey: ["tracking", "water"] });
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
            queryClient.invalidateQueries({ queryKey: ["tracking", "water"] });
        },
    });
}
