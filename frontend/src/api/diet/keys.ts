export const dietKeys = {
    all: ['diet'] as const,
    days: {
        /** Prefix covering per-day logs and month planned summaries. */
        all: ['diet', 'days'] as const,
        /** Keyed by absolute local date (`YYYY-MM-DD`), never a relative offset. */
        day: (date: string) => ['diet', 'days', 'day', date] as const,
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
    mutations: {
        logPlanned: () => ['diet', 'mutations', 'log-planned'] as const,
        unlog: () => ['diet', 'mutations', 'unlog'] as const,
        quickLog: () => ['diet', 'mutations', 'quick-log'] as const,
        addPlanned: () => ['diet', 'mutations', 'add-planned'] as const,
        deletePlanned: () => ['diet', 'mutations', 'delete-planned'] as const,
        reorderPlanned: () => ['diet', 'mutations', 'reorder-planned'] as const,
    },
    savedMeals: {
        all: ['diet', 'saved-meals'] as const,
        list: () => ['diet', 'saved-meals', 'list'] as const,
        detail: (id: number) => ['diet', 'saved-meals', 'detail', id] as const,
    },
} as const;
