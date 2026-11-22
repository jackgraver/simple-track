export const trackingKeys = {
    all: ['tracking'] as const,
    weight: ['tracking', 'weight'] as const,
    profile: ['tracking', 'profile'] as const,
    steps: ['tracking', 'steps'] as const,
    groceryItems: ['tracking', 'grocery', 'items'] as const,
    /** Prefix of every suggestions query (any search term). */
    grocerySuggestionsAll: ['tracking', 'grocery', 'suggestions'] as const,
    grocerySuggestions: (query: string) =>
        ['tracking', 'grocery', 'suggestions', query] as const,
    /** Prefix of every per-date water log query. Does not include presets. */
    waterLogsAll: ['tracking', 'water', 'logs'] as const,
    water: (date: string) => ['tracking', 'water', 'logs', date] as const,
    waterPresets: ['tracking', 'water', 'presets'] as const,
    mutations: {
        addWater: ['tracking', 'mutation', 'water-add'] as const,
        deleteWater: ['tracking', 'mutation', 'water-delete'] as const,
        saveWeight: ['tracking', 'mutation', 'weight-save'] as const,
        saveSteps: ['tracking', 'mutation', 'steps-save'] as const,
    },
};
