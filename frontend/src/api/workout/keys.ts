const liveworkoutKeysBase = {
    all: ['liveworkout'] as const,
} as const;

export const liveworkoutKeys = {
    all: liveworkoutKeysBase.all,
    workouts: {
        all: [...liveworkoutKeysBase.all, 'workouts'] as const,
        day: (offset: number) =>
            [...liveworkoutKeysBase.all, 'workouts', 'day', offset] as const,
        previous: (offset: number) =>
            [...liveworkoutKeysBase.all, 'workouts', 'previous', offset] as const,
        today: () => [...liveworkoutKeysBase.all, 'workouts', 'today'] as const,
        activity: (opts: {
            mode: 'year' | 'rolling';
            weeks: number | null;
            days: number | null;
        }) =>
            [
                ...liveworkoutKeysBase.all,
                'workouts',
                'activity',
                opts.mode,
                opts.weeks,
                opts.days,
            ] as const,
        activityPrefix: () =>
            [...liveworkoutKeysBase.all, 'workouts', 'activity'] as const,
    },
    exercises: {
        all: [...liveworkoutKeysBase.all, 'exercises'] as const,
        allList: () => [...liveworkoutKeysBase.all, 'exercises', 'all'] as const,
        paginated: (search: string) =>
            [...liveworkoutKeysBase.all, 'exercises', 'paginated', search] as const,
    },
} as const;
