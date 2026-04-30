export const trackingKeys = {
    weight: ['tracking', 'weight'] as const,
    steps: ['tracking', 'steps'] as const,
    missed: ['tracking', 'missed'] as const,
    water: (date: string) => ['tracking', 'water', date] as const,
    waterPresets: ['tracking', 'water', 'presets'] as const,
};
