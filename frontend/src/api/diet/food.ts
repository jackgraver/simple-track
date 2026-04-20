import { apiClient } from '~/api/client';
import type { Food } from '~/types/diet';

export type CreateFoodResponse = {
    food: Food;
};

export async function createFood(food: Food): Promise<CreateFoodResponse> {
    const response = await apiClient.post<CreateFoodResponse>('/diet/foods', food);
    return response.data;
}

/** USDA FDC short row from GET /diet/foods/search */
export type FdcFoodSummary = {
    fdc_id: number;
    name: string;
    brand?: string;
    serving_size: number;
    serving_size_unit: string;
    calories: number;
    protein_g: number;
    carbs_g: number;
    fat_g: number;
    fiber_g: number;
    sugar_g: number;
};

export type FdcFoodSearchResponse = {
    foods: FdcFoodSummary[];
};

export async function searchFoods(
    q: string,
    signal?: AbortSignal,
): Promise<FdcFoodSearchResponse> {
    const response = await apiClient.get<FdcFoodSearchResponse>('/diet/foods/search', {
        params: { q },
        signal,
    });
    return response.data;
}
