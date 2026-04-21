import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import {
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

export function useSaveWeight() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ date, weightLbs }: { date: string; weightLbs: number }) =>
            saveWeight(date, weightLbs),
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
