import {
    queryOptions,
    useMutation,
    useQueries,
    useQuery,
    useQueryClient,
    type QueryClient,
} from '@tanstack/vue-query';
import { computed, toValue, type MaybeRefOrGetter } from 'vue';
import { useRouter } from 'vue-router';
import {
    addPlannedMealFromSaved,
    copyPlannedDay,
    createCompositeFood,
    createMeal,
    createSavedMeal,
    deleteLoggedMeal,
    deletePlannedMeal,
    deleteSavedMeal,
    getAllPlans,
    getDietLogsToday,
    getMealById,
    getSavedMealById,
    getSavedMeals,
    logEditedMeal,
    logPlannedMeal,
    quickLog,
    reorderPlannedMeals,
    updateLoggedMeal,
    updatePlanMacros,
    updateSavedMeal,
    type QuickLogPayload,
    type SavedMealItemPayload,
    type UpdatePlanMacrosPayload,
} from './api';
import { createFood } from './food';
import { dietKeys } from './keys';
import { getApiErrorMessage } from '~/api/httpError';
import { toast } from '~/composables/toast/useToast';
import { searchListKeys } from '~/shared/searchListKeys';
import { parseDietDayOffsetQuery, ymdForDayOffset } from '~/utils/dateUtil';
import type { Food, Meal } from '~/types/diet';

const DAY_STALE_MS = 1000 * 60 * 2;
const DETAIL_STALE_MS = 1000 * 60 * 5;

// ---- Queries ----

/**
 * Query options for one day. The offset is the API argument; the cache key is the absolute
 * date it resolves to *now*, so the key and the request always describe the same day.
 */
export const dietQueries = {
    day: (offset: number) =>
        queryOptions({
            queryKey: dietKeys.days.day(ymdForDayOffset(offset)),
            queryFn: () => getDietLogsToday(offset),
            staleTime: DAY_STALE_MS,
        }),
    plans: () => queryOptions({ queryKey: dietKeys.plans.all, queryFn: getAllPlans }),
    savedMeals: () =>
        queryOptions({ queryKey: dietKeys.savedMeals.list(), queryFn: getSavedMeals }),
    savedMeal: (id: number) =>
        queryOptions({
            queryKey: dietKeys.savedMeals.detail(id),
            queryFn: () => getSavedMealById(id),
            staleTime: DETAIL_STALE_MS,
        }),
    meal: (id: number) =>
        queryOptions({
            queryKey: dietKeys.meals.detail(id),
            queryFn: () => getMealById(id),
            staleTime: DETAIL_STALE_MS,
        }),
};

export function useDietLogsToday(offset: MaybeRefOrGetter<number>) {
    return useQuery(computed(() => dietQueries.day(toValue(offset))));
}

/** The `n` days before today (offset 1..n), sharing cache entries with `useDietLogsToday`. */
export function useRecentDietDays(n: number) {
    return useQueries({
        queries: Array.from({ length: n }, (_, i) => dietQueries.day(i + 1)),
    });
}

/** Several days by offset (e.g. a week), sharing cache entries with `useDietLogsToday`. */
export function useDietDays(offsets: MaybeRefOrGetter<number[]>) {
    return useQueries({
        queries: computed(() => toValue(offsets).map((o) => dietQueries.day(o))),
    });
}

export function usePlans() {
    return useQuery(dietQueries.plans());
}

/** The saved-meal list, shared (same query key) by the day editor and the range planner. */
export function useSavedMeals() {
    const query = useQuery(dietQueries.savedMeals());
    const savedMeals = computed(() => query.data.value ?? []);
    return { ...query, savedMeals };
}

function idOrNull(id: MaybeRefOrGetter<number | null>): number | null {
    const v = toValue(id);
    return v ? v : null;
}

export function useMeal(id: MaybeRefOrGetter<number | null>) {
    return useQuery(
        computed(() => {
            const mealId = idOrNull(id);
            return { ...dietQueries.meal(mealId ?? 0), enabled: mealId !== null };
        }),
    );
}

export function useSavedMeal(id: MaybeRefOrGetter<number | null>) {
    return useQuery(
        computed(() => {
            const savedMealId = idOrNull(id);
            return { ...dietQueries.savedMeal(savedMealId ?? 0), enabled: savedMealId !== null };
        }),
    );
}

// ---- Day mutations: the target day travels in the variables ----

function invalidateMonthPlannedSummaries(queryClient: QueryClient) {
    queryClient.invalidateQueries({ queryKey: dietKeys.days.monthPlannedSummaryAll });
}

/** Resolve the key date when the mutation starts, not when it settles (the user may have swiped or crossed midnight). */
const dayContext = (vars: { offset: number }) => ({ date: ymdForDayOffset(vars.offset) });

function invalidateDay(queryClient: QueryClient, date: string) {
    queryClient.invalidateQueries({ queryKey: dietKeys.days.day(date) });
    invalidateMonthPlannedSummaries(queryClient);
}

export function useLogPlannedMeal() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ mealId }: { mealId: number; offset: number }) => logPlannedMeal(mealId),
        onMutate: dayContext,
        onSuccess: (day, _vars, ctx) => {
            queryClient.setQueryData(dietKeys.days.day(ctx.date), day);
            invalidateMonthPlannedSummaries(queryClient);
        },
    });
}

export function useDeleteLoggedMeal() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ mealId, dayId }: { mealId: number; dayId: number; offset: number }) =>
            deleteLoggedMeal(mealId, dayId),
        onMutate: dayContext,
        onSuccess: (day, _vars, ctx) => {
            queryClient.setQueryData(dietKeys.days.day(ctx.date), day);
        },
    });
}

// Planning across days: the week board edits 7 days, so each call names its own day.

export function useAddPlannedFromSavedAt() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ savedMealId, offset }: { savedMealId: number; offset: number }) =>
            addPlannedMealFromSaved(savedMealId, offset),
        onMutate: dayContext,
        onSuccess: (_d, _vars, ctx) => invalidateDay(queryClient, ctx.date),
    });
}

export function useDeletePlannedMealAt() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ plannedMealId, offset }: { plannedMealId: number; offset: number }) =>
            deletePlannedMeal(plannedMealId, offset),
        onMutate: dayContext,
        onSuccess: (_d, _vars, ctx) => invalidateDay(queryClient, ctx.date),
    });
}

export function useReorderPlannedMealsAt() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ ids, offset }: { ids: number[]; offset: number }) =>
            reorderPlannedMeals(ids, offset),
        onMutate: dayContext,
        onSuccess: (_d, _vars, ctx) => invalidateDay(queryClient, ctx.date),
    });
}

export function useCopyPlannedDay() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ from, to }: { from: number; to: number[] }) => copyPlannedDay(from, to),
        onMutate: ({ to }) => ({ dates: to.map(ymdForDayOffset) }),
        onSuccess: (_d, _vars, ctx) => {
            for (const date of ctx.dates) {
                queryClient.invalidateQueries({ queryKey: dietKeys.days.day(date) });
            }
            invalidateMonthPlannedSummaries(queryClient);
        },
    });
}

// ---- Plan targets ----

export function useSavePlanMacros() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ planId, macros }: { planId: number; macros: UpdatePlanMacrosPayload }) =>
            updatePlanMacros(planId, macros),
        onSuccess: async () => {
            await Promise.all([
                queryClient.invalidateQueries({ queryKey: dietKeys.days.all }),
                queryClient.invalidateQueries({ queryKey: dietKeys.plans.all }),
            ]);
        },
    });
}

// ---- Meal logging / saved meals / foods ----

const SAVED_MEALS_SEARCH_PATH = '/diet/meals/saved-meal/all';

function invalidateSavedMeals(queryClient: QueryClient) {
    queryClient.invalidateQueries({ queryKey: dietKeys.savedMeals.all });
    queryClient.invalidateQueries({
        queryKey: searchListKeys.path(SAVED_MEALS_SEARCH_PATH),
    });
}

function invalidateDietQueries(queryClient: QueryClient) {
    queryClient.invalidateQueries({ queryKey: dietKeys.days.all });
    invalidateSavedMeals(queryClient);
}

function afterMealLogNavigate(router: ReturnType<typeof useRouter>) {
    const route = router.currentRoute.value;
    const name = route.name;
    if (name === 'diet-log') {
        const offset = parseDietDayOffsetQuery(route.query.offset);
        router.push({
            name: 'diet',
            query: offset === 0 ? {} : { offset: String(offset) },
        });
        return;
    }
    if (name === 'home') {
        return;
    }
    router.push({ name: 'gym' });
}

export function useQuickLog() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: QuickLogPayload) => quickLog(payload),
        onSuccess: () => {
            invalidateDietQueries(queryClient);
        },
    });
}

export function useCreateSavedMeal() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: { name: string; items: SavedMealItemPayload[] }) =>
            createSavedMeal(payload),
        onSuccess: () => {
            invalidateSavedMeals(queryClient);
        },
    });
}

export function useDeleteSavedMeal() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (vars: { savedMealId: number; force?: boolean }) =>
            deleteSavedMeal(vars.savedMealId, { force: vars.force }),
        onSuccess: () => {
            invalidateDietQueries(queryClient);
        },
    });
}

export function useCreateCompositeFood() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: { name: string; items: { food_id: number; amount: number }[] }) =>
            createCompositeFood(payload),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: searchListKeys.all });
        },
    });
}

export function useCreateMeal() {
    const queryClient = useQueryClient();
    const router = useRouter();

    return useMutation({
        mutationFn: ({
            meal,
            log,
            saveToLibrary,
            offset,
        }: {
            meal: Meal;
            log: boolean;
            saveToLibrary?: boolean;
            offset?: number;
        }) => createMeal(meal, log, { saveToLibrary, offset }),
        onSuccess: (_, variables) => {
            invalidateDietQueries(queryClient);
            if (variables.log) {
                afterMealLogNavigate(router);
            }
        },
    });
}

export function useLogEditedMeal() {
    const queryClient = useQueryClient();
    const router = useRouter();

    return useMutation({
        mutationFn: (vars: {
            meal: Meal;
            plannedSourceMealId?: number;
            dayId?: number;
        }) =>
            logEditedMeal(vars.meal, {
                plannedSourceMealId: vars.plannedSourceMealId,
                dayId: vars.dayId,
            }),
        onSuccess: () => {
            invalidateDietQueries(queryClient);
            afterMealLogNavigate(router);
        },
    });
}

export function useUpdateLoggedMeal() {
    const queryClient = useQueryClient();
    const router = useRouter();

    return useMutation({
        mutationFn: ({
            meal,
            oldMealId,
            dayId,
        }: {
            meal: Meal;
            oldMealId: number;
            dayId?: number;
        }) => updateLoggedMeal(meal, oldMealId, { dayId }),
        onSuccess: () => {
            invalidateDietQueries(queryClient);
            afterMealLogNavigate(router);
        },
    });
}

export function useUpdateSavedMeal() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (payload: {
            savedMealId: number;
            name: string;
            items: SavedMealItemPayload[];
        }) =>
            updateSavedMeal(payload.savedMealId, {
                name: payload.name,
                items: payload.items,
            }),
        onSuccess: (_data, variables) => {
            queryClient.invalidateQueries({
                queryKey: dietKeys.savedMeals.detail(variables.savedMealId),
            });
            invalidateSavedMeals(queryClient);
        },
    });
}


export function useCreateFood() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (vars: { food: Food; relatedFoodId?: number }) =>
            createFood(vars.food, vars.relatedFoodId),
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: searchListKeys.all });
            toast.push("Food Created Successfully!", "success");
        },
        onError: (error) => {
            toast.push(getApiErrorMessage(error, "Create food failed"), "error");
        },
    });
}

