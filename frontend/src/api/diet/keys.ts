export const dietKeys = {
    all: ['diet'] as const,
    days: {
        /** Prefix covering per-day logs and month planned summaries. */
        all: ['diet', 'days'] as const,
        day: (offset: number) => ['diet', 'days', 'day', offset] as const,
        monthPlannedSummaryAll: ['diet', 'days', 'month-planned-summary'] as const,
        monthPlannedSummary: (monthOffset: number) =>
            ['diet', 'days', 'month-planned-summary', monthOffset] as const,
    },
    plans: {
        all: ['diet', 'plans'] as const,
    },
    meals: {
        detail: (id: number) => ['diet', 'meals', 'detail', id] as const,
    },
    savedMeals: {
        all: ['diet', 'saved-meals'] as const,
        list: () => ['diet', 'saved-meals', 'list'] as const,
        detail: (id: number) => ['diet', 'saved-meals', 'detail', id] as const,
    },
} as const;
