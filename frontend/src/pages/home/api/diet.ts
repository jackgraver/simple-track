import { apiClient } from '~/utils/axios';
import type { Day, Meal } from '~/types/diet';

export type DietLogsTodayResponse = {
    day: Day;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
    totalCarbs: number;
};

export async function getDietLogsToday(offset: number): Promise<DietLogsTodayResponse> {
    const response = await apiClient.get<DietLogsTodayResponse>(
        '/diet/logs/today',
        {
            params: { offset },
        }
    );
    return response.data;
}

export async function logPlannedMeal(mealId: number): Promise<DietLogsTodayResponse> {
    const response = await apiClient.post<DietLogsTodayResponse>(
        '/diet/meals/meal/log-planned',
        { meal_id: mealId }
    );
    return response.data;
}

export async function deleteLoggedMeal(mealId: number, dayId: number): Promise<DietLogsTodayResponse> {
    const response = await apiClient.delete<DietLogsTodayResponse>(
        '/diet/meals/meal/logged',
        {
            data: {
                meal_id: mealId,
                day_id: dayId,
            },
        }
    );
    return response.data;
}

export async function editLoggedMeal(meal: Meal, oldMealId: number): Promise<DietLogsTodayResponse> {
    const response = await apiClient.post<DietLogsTodayResponse>(
        '/diet/logs/meal/editlogged',
        {
            meal,
            oldMealID: oldMealId,
        }
    );
    return response.data;
}

