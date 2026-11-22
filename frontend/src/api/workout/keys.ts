/** One root (`['workout']`). Day-scoped keys use the absolute date, never the offset. */
export type WorkoutActivityKeyOpts = {
    mode: 'year' | 'rolling';
    weeks: number | null;
    days: number | null;
};

export const workoutKeys = {
    all: ['workout'] as const,
    days: {
        all: ['workout', 'days'] as const,
        /** `GET /workout/logs/previous` for one date. Readers needing a slice use `select`. */
        day: (date: string) => ['workout', 'days', 'day', date] as const,
    },
    activity: {
        all: ['workout', 'activity'] as const,
        list: (opts: WorkoutActivityKeyOpts) =>
            ['workout', 'activity', opts.mode, opts.weeks, opts.days] as const,
    },
    exercises: {
        all: ['workout', 'exercises'] as const,
        list: () => ['workout', 'exercises', 'all'] as const,
        paginated: (search: string) => ['workout', 'exercises', 'paginated', search] as const,
    },
    plans: {
        all: ['workout', 'plans'] as const,
        list: () => ['workout', 'plans', 'all'] as const,
    },
    programs: ['workout', 'programs'] as const,
    /** Mutation keys (not query keys): pending/failed state is read from the mutation cache. */
    mutations: {
        sets: ['workout', 'mutations', 'sets'] as const,
        logSet: () => ['workout', 'mutations', 'sets', 'create'] as const,
        updateSet: () => ['workout', 'mutations', 'sets', 'update'] as const,
        deleteSet: () => ['workout', 'mutations', 'sets', 'delete'] as const,
        notes: () => ['workout', 'mutations', 'sets', 'notes'] as const,
        addExercise: () => ['workout', 'mutations', 'exercise', 'add'] as const,
        removeExercise: () => ['workout', 'mutations', 'exercise', 'remove'] as const,
        cardio: () => ['workout', 'mutations', 'cardio'] as const,
        mobilityPre: () => ['workout', 'mutations', 'mobility', 'pre'] as const,
        mobilityPost: () => ['workout', 'mutations', 'mobility', 'post'] as const,
        switchPlan: () => ['workout', 'mutations', 'switch-plan'] as const,
    },
} as const;
