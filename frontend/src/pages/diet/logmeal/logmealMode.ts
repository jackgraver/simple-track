/** URL query `?type=` values for `/diet/log`. */
export const LOG_TYPE = "log" as const;
export const CREATE_TYPE = "create" as const;
export const EDIT_TYPE = "edit" as const;
export const EDIT_LOGGED_TYPE = "editlogged" as const;

/** Resolved page mode (three UI states). */
export const PAGE_MODE = {
    create: CREATE_TYPE,
    log: LOG_TYPE,
    edit: EDIT_TYPE,
} as const;

/** When `pageMode === PAGE_MODE.edit`, how the edit flow behaves. */
export const EDIT_VARIANT = {
    logged: "logged",
    planned: "planned",
} as const;

export type LogMealPageMode = (typeof PAGE_MODE)[keyof typeof PAGE_MODE];

export type EditMealVariant =
    (typeof EDIT_VARIANT)[keyof typeof EDIT_VARIANT];

export type LogMealQueryTypeParam =
    | typeof LOG_TYPE
    | typeof CREATE_TYPE
    | typeof EDIT_TYPE
    | typeof EDIT_LOGGED_TYPE;

export function parseLogMealPageMode(
    queryType: string | string[] | null | undefined,
): LogMealPageMode {
    const raw = normalizeQueryString(queryType);
    if (raw === CREATE_TYPE) return PAGE_MODE.create;
    if (raw === EDIT_TYPE || raw === EDIT_LOGGED_TYPE) return PAGE_MODE.edit;
    if (raw === LOG_TYPE) return PAGE_MODE.log;
    return PAGE_MODE.log;
}

export function parseEditMealVariant(
    queryType: string | string[] | null | undefined,
): EditMealVariant {
    return normalizeQueryString(queryType) === EDIT_LOGGED_TYPE
        ? EDIT_VARIANT.planned
        : EDIT_VARIANT.logged;
}

function normalizeQueryString(
    queryType: string | string[] | null | undefined,
): string {
    const v = Array.isArray(queryType) ? queryType[0] : queryType;
    return typeof v === "string" ? v.trim() : "";
}
