const homeKeysBase = {
    all: ['home'] as const,
} as const;

export const homeKeys = {
    all: homeKeysBase.all,
    workouts: {
        all: [...homeKeysBase.all, 'workouts'] as const,
        previous: (offset: number) => 
            [...homeKeysBase.all, 'workouts', 'previous', offset] as const,
    },
    diet: {
        all: [...homeKeysBase.all, 'diet'] as const,
        today: (offset: number) => 
            [...homeKeysBase.all, 'diet', 'today', offset] as const,
    },
} as const;

