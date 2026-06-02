import { computed } from "vue";
import { useRoute } from "vue-router";
import { parseDietDayOffsetQuery } from "~/utils/dateUtil";
import {
    EDIT_VARIANT,
    PAGE_MODE,
    parseEditMealVariant,
    parseLogMealPageMode,
    type LogMealPageMode,
} from "./logmealMode";

export function useLogMealMode() {
    const route = useRoute();

    const dayOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));

    const queryType = computed(() => {
        const t = route.query.type;
        return Array.isArray(t) ? t[0] : t;
    });

    const pageMode = computed(
        (): LogMealPageMode => parseLogMealPageMode(queryType.value),
    );

    const editVariant = computed(() =>
        pageMode.value === PAGE_MODE.edit
            ? parseEditMealVariant(queryType.value)
            : null,
    );

    const id = computed(() => Number(route.query.id ?? 0));

    const mealLogDayId = computed(() => {
        const d = route.query.dayId;
        const v = Array.isArray(d) ? d[0] : d;
        const n = Number(v ?? 0);
        return Number.isFinite(n) && n > 0 ? n : undefined;
    });

    const mealId = computed(() => {
        if (pageMode.value !== PAGE_MODE.edit) return null;
        if (editVariant.value === EDIT_VARIANT.saved) return null;
        return id.value !== 0 ? id.value : null;
    });

    const savedMealEditId = computed(() => {
        if (pageMode.value !== PAGE_MODE.edit) return null;
        if (editVariant.value !== EDIT_VARIANT.saved) return null;
        return id.value !== 0 ? id.value : null;
    });

    const editMissingId = computed(
        () => pageMode.value === PAGE_MODE.edit && id.value === 0,
    );

    const pageTitle = computed(() => {
        switch (pageMode.value) {
            case PAGE_MODE.create:
                return "Create New Meal";
            case PAGE_MODE.log:
                return "Log Meal";
            case PAGE_MODE.edit:
                if (editVariant.value === EDIT_VARIANT.saved) {
                    return "Edit Saved Meal";
                }
                return editVariant.value === EDIT_VARIANT.planned
                    ? "Log Meal"
                    : "Edit Logged Meal";
            default:
                return "Log Meal";
        }
    });

    const showDietDayMacroBars = computed(
        () =>
            pageMode.value !== PAGE_MODE.create &&
            !(
                pageMode.value === PAGE_MODE.edit &&
                editVariant.value === EDIT_VARIANT.saved
            ),
    );

    return {
        dayOffset,
        queryType,
        pageMode,
        editVariant,
        id,
        mealLogDayId,
        mealId,
        savedMealEditId,
        editMissingId,
        pageTitle,
        showDietDayMacroBars,
    };
}
