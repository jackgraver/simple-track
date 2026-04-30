import { apiDELETE, apiGET, apiPOST, apiPUT } from '~/api/client';
import type {
    BodyWeightLog,
    DrinkSizePreset,
    MissedTracking,
    StepLog,
    WaterLog,
} from '~/api/tracking/types';

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

export async function fetchWaterLogs(date?: string): Promise<WaterLog[]> {
    const params = date ? { date } : undefined;
    const res = await apiGET<{ logs: WaterLog[] }>('/tracking/water', { params });
    return res.logs ?? [];
}

export async function saveWater(
    date: string,
    amountOz: number,
    presetId?: number,
): Promise<WaterLog> {
    const body: Record<string, unknown> = {
        date: date || undefined,
        amount_oz: amountOz,
    };
    if (presetId != null) body.preset_id = presetId;
    const res = await apiPOST<{ log: WaterLog }>('/tracking/water', body);
    return res.log;
}

export async function deleteWater(id: number): Promise<void> {
    await apiDELETE(`/tracking/water/${id}`);
}

export async function fetchDrinkSizePresets(): Promise<DrinkSizePreset[]> {
    const res = await apiGET<{ presets: DrinkSizePreset[] }>('/tracking/water/presets');
    return res.presets ?? [];
}

export async function createDrinkSizePreset(
    name: string,
    amountOz: number,
): Promise<DrinkSizePreset> {
    const res = await apiPOST<{ preset: DrinkSizePreset }>('/tracking/water/presets', {
        name,
        amount_oz: amountOz,
    });
    return res.preset;
}

export async function updateDrinkSizePreset(
    id: number,
    name: string,
    amountOz: number,
): Promise<DrinkSizePreset> {
    const res = await apiPUT<{ preset: DrinkSizePreset }>(`/tracking/water/presets/${id}`, {
        name,
        amount_oz: amountOz,
    });
    return res.preset;
}

export async function deleteDrinkSizePreset(id: number): Promise<void> {
    await apiDELETE(`/tracking/water/presets/${id}`);
}
