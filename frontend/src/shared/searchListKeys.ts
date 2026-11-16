/** Query keys used by `SearchList.vue` (keyed by request path). */
export const searchListKeys = {
    all: ['searchList'] as const,
    path: (requestPath: string) => ['searchList', requestPath] as const,
};
