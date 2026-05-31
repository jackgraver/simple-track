export const trackingKeys = {
    weight: ['tracking', 'weight'] as const,
    profile: ['tracking', 'profile'] as const,
    steps: ['tracking', 'steps'] as const,
    groceryItems: ['tracking', 'grocery', 'items'] as const,
    grocerySuggestions: (query: string) =>
        ['tracking', 'grocery', 'suggestions', query] as const,
    missed: ['tracking', 'missed'] as const,
    water: (date: string) => ['tracking', 'water', date] as const,
    waterPresets: ['tracking', 'water', 'presets'] as const,
};
