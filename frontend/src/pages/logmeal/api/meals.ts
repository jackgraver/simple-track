import { apiClient } from '~/utils/axios';
import type { Meal, Day } from '~/types/diet';

export type MealResponse = {
    meal: Meal;
};

export type DietLogsTodayResponse = {
    day: Day;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
    totalCarbs: number;
};

export type CreateMealResponse = {
    meal_id: number;
};

export type LogMealResponse = {
    day: Day;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
};

export async function getMealById(id: number): Promise<MealResponse> {
    const response = await apiClient.get<MealResponse>(`/diet/meals/meal/${id}`);
    return response.data;
}

export async function getDietLogsToday(): Promise<DietLogsTodayResponse> {
    const response = await apiClient.get<DietLogsTodayResponse>('/diet/logs/today');
    return response.data;
}

export async function createMeal(meal: Meal, log: boolean): Promise<CreateMealResponse> {
    const response = await apiClient.post<CreateMealResponse>('/diet/meals/meal/new', {
        meal,
        log,
    });
    return response.data;
}

export async function logEditedMeal(meal: Meal): Promise<LogMealResponse> {
    const response = await apiClient.post<LogMealResponse>('/diet/meals/meal/logedited', {
        meal,
    });
    return response.data;
}

export async function updateLoggedMeal(meal: Meal, oldMealId: number): Promise<LogMealResponse> {
    const response = await apiClient.post<LogMealResponse>('/diet/meals/meal/editlogged', {
        meal,
        oldMealID: oldMealId,
    });
    return response.data;
}

