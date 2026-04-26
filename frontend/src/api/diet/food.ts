import { apiClient } from '~/api/client';
import type { Food } from '~/types/diet';

export type CreateFoodResponse = {
    food: Food;
};

export async function createFood(
    food: Food,
    relatedFoodId?: number,
): Promise<CreateFoodResponse> {
    const body: Record<string, unknown> = {
        name: food.name,
        serving_type: food.serving_type,
        serving_amount: food.serving_amount,
        calories: food.calories,
        protein: food.protein,
        fiber: food.fiber,
        carbs: food.carbs,
    };
    if (relatedFoodId != null && relatedFoodId > 0) {
        body.related_food_id = relatedFoodId;
    }
    const response = await apiClient.post<CreateFoodResponse>("/diet/foods", body);
    return response.data;
}
