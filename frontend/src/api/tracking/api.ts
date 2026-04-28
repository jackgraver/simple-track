import { apiGET, apiPOST } from '~/api/client';
import type { BodyWeightLog, MissedTracking, StepLog } from '~/api/tracking/types';

export async function fetchWeightLogs(limit?: number): Promise<BodyWeightLog[]> {
    const params = limit != null && limit > 0 ? { limit } : undefined;
    const res = await apiGET<{ logs: BodyWeightLog[] }>('/tracking/weight', { params });
    return res.logs ?? [];
}

export async function saveWeight(date: string, weightLbs: number): Promise<BodyWeightLog> {
    const res = await apiPOST<{ log: BodyWeightLog }>('/tracking/weight', {
        date: date || undefined,
        weight_lbs: weightLbs,
    });
    return res.log;
}

export async function fetchStepLogs(limit?: number): Promise<StepLog[]> {
    const params = limit != null && limit > 0 ? { limit } : undefined;
    const res = await apiGET<{ logs: StepLog[] }>('/tracking/steps', { params });
    return res.logs ?? [];
}

export async function fetchMissedTracking(): Promise<MissedTracking> {
    return apiGET<MissedTracking>('/tracking/missed');
}

export async function saveSteps(date: string, steps: number): Promise<StepLog> {
    const res = await apiPOST<{ log: StepLog }>('/tracking/steps', {
        date: date || undefined,
        steps,
    });
    return res.log;
}
