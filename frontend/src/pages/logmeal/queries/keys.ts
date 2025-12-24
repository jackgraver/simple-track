const logmealKeysBase = {
    all: ['logmeal'] as const,
} as const;

export const logmealKeys = {
    all: logmealKeysBase.all,
    meals: {
        all: [...logmealKeysBase.all, 'meals'] as const,
        detail: (id: number) => 
            [...logmealKeysBase.all, 'meals', 'detail', id] as const,
    },
    diet: {
        all: [...logmealKeysBase.all, 'diet'] as const,
        today: () => 
            [...logmealKeysBase.all, 'diet', 'today'] as const,
    },
    foods: {
        all: [...logmealKeysBase.all, 'foods'] as const,
    },
} as const;

