const liveworkoutKeysBase = {
    all: ['liveworkout'] as const,
} as const;

export const liveworkoutKeys = {
    all: liveworkoutKeysBase.all,
    workouts: {
        all: [...liveworkoutKeysBase.all, 'workouts'] as const,
        previous: (offset: number) => 
            [...liveworkoutKeysBase.all, 'workouts', 'previous', offset] as const,
    },
    exercises: {
        all: [...liveworkoutKeysBase.all, 'exercises'] as const,
        allList: () => 
            [...liveworkoutKeysBase.all, 'exercises', 'all'] as const,
    },
} as const;

