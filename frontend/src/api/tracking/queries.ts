import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { computed } from 'vue';
import { isAuthenticated } from '~/composables/auth/session';
import {
    fetchMissedTracking,
    fetchStepLogs,
    fetchWeightLogs,
    saveSteps,
    saveWeight,
} from '~/api/tracking/api';
import { trackingKeys } from '~/api/tracking/keys';

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
        mutationFn: ({ date, weightLbs }: { date: string; weightLbs: number }) =>
            saveWeight(date, weightLbs),
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
