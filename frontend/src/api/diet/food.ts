import { apiClient } from '~/utils/axios';
import type { Food } from '~/types/diet';

export type CreateFoodResponse = {
    food: Food;
};

export async function createFood(food: Food): Promise<CreateFoodResponse> {
    const response = await apiClient.post<CreateFoodResponse>('/diet/foods', food);
    return response.data;
}
