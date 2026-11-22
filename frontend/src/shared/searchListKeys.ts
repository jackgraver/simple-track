/** Query keys used by `SearchList.vue` (keyed by route only; exclusion is client-side). */
export const searchListKeys = {
    all: ['searchList'] as const,
    path: (route: string) => ['searchList', route] as const,
};
