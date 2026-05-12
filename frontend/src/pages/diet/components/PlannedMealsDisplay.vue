<script setup lang="ts">
import { ChevronDown, ChevronUp, Pencil } from "lucide-vue-next";
import { useRouter } from "vue-router";
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import MealCard from "./MealCard.vue";
import { useDietDayMealHandlers } from "./useDietDayMealHandlers";

const props = defineProps<{
    dateOffset: number;
}>();

const router = useRouter();

const { data, logPlannedMeal, logMeal, deleteLoggedMeal } =
    useDietDayMealHandlers(() => props.dateOffset);

const editPlannedMeal = () => {
    router.push({
        name: "diet-edit-planned",
        query: { offset: String(props.dateOffset) },
    });
};

const CARD_HEIGHT_ESTIMATE = 140;
const BOTTOM_PADDING = 24;

const isMobile = ref(false);
const containerRef = ref<HTMLElement | null>(null);
const maxVisible = ref(2);
const start = ref(0);

let mediaQuery: MediaQueryList | null = null;

function onMediaChange(e: MediaQueryListEvent | MediaQueryList) {
    isMobile.value = e.matches;
}

function checkOverflow() {
    if (isMobile.value || !containerRef.value) return;
    const bottom = containerRef.value.getBoundingClientRect().bottom;
    if (
        bottom > window.innerHeight - BOTTOM_PADDING &&
        maxVisible.value > 1
    ) {
        maxVisible.value--;
    }
}

function recalcMaxVisible() {
    if (isMobile.value || !containerRef.value) return;
    const top = containerRef.value.getBoundingClientRect().top;
    const available = window.innerHeight - top - BOTTOM_PADDING;
    const estimated = Math.max(1, Math.floor(available / CARD_HEIGHT_ESTIMATE));
    const len = (data.value?.day.plannedMeals ?? []).length;
    maxVisible.value =
        len > 0 ? Math.min(estimated, len) : estimated;
    void nextTick(() => {
        checkOverflow();
    });
}

onMounted(() => {
    mediaQuery = window.matchMedia("(max-width: 767px)");
    onMediaChange(mediaQuery);
    mediaQuery.addEventListener("change", onMediaChange);
    window.addEventListener("resize", recalcMaxVisible);
    recalcMaxVisible();
});

onUnmounted(() => {
    mediaQuery?.removeEventListener("change", onMediaChange);
    window.removeEventListener("resize", recalcMaxVisible);
});

const sortedPlannedMeals = computed(() => {
    const list = data.value?.day.plannedMeals ?? [];
    return [...list].sort(
        (a, b) =>
            (a.display_order ?? 0) - (b.display_order ?? 0) || a.ID - b.ID,
    );
});

watch(
    () => sortedPlannedMeals.value.length,
    () => {
        recalcMaxVisible();
    },
);

watch(
    () => [sortedPlannedMeals.value.length, maxVisible.value],
    ([len, max]) => {
        const maxStart = Math.max(0, len - max);
        if (start.value > maxStart) start.value = maxStart;
    },
);

const visibleItems = computed(() =>
    isMobile.value
        ? sortedPlannedMeals.value
        : sortedPlannedMeals.value.slice(start.value, start.value + maxVisible.value),
);

watch(visibleItems, () => {
    void nextTick(() => {
        checkOverflow();
    });
});

const showChevrons = computed(
    () => !isMobile.value && sortedPlannedMeals.value.length > maxVisible.value,
);

function next() {
    start.value = Math.min(
        start.value + 1,
        Math.max(
            0,
            sortedPlannedMeals.value.length - maxVisible.value,
        ),
    );
}
function prev() {
    start.value = Math.max(start.value - 1, 0);
}
</script>

<template>
    <div ref="containerRef" class="flex min-w-0 flex-1 flex-col gap-2 pt-2">
        <div class="flex flex-row items-center justify-between gap-2">
            <div class="flex min-w-0 flex-1 items-center gap-2">
                <h2 class="mb-0 flex-1 text-lg font-semibold">
                    Planned ({{ sortedPlannedMeals.length }})
                </h2>
                <button
                    type="button"
                    class="flex items-center gap-1 rounded-md border px-2 py-1 m-0! text-sm font-medium text-zinc-200"
                    @click="editPlannedMeal"
                >
                    <Pencil :size="15" />
                    Edit
                </button>
            </div>
            <template v-if="data && showChevrons">
                <button
                    type="button"
                    :disabled="start === 0"
                    class="shrink-0 disabled:text-zinc-500"
                    aria-label="Previous"
                    @click="prev"
                >
                    <ChevronUp />
                </button>
                <button
                    type="button"
                    :disabled="start >= sortedPlannedMeals.length - maxVisible"
                    class="shrink-0 disabled:text-zinc-500"
                    aria-label="Next"
                    @click="next"
                >
                    <ChevronDown />
                </button>
            </template>
        </div>
        <template v-if="data">
            <span v-if="visibleItems?.length === 0" class="text-zinc-500"
                >Nothing else planned.</span
            >
            <MealCard
                v-for="log in visibleItems"
                :key="log.ID"
                :meal="log.meal"
                :on-log-planned="logPlannedMeal"
                :on-log-edited="logMeal"
                :on-delete="deleteLoggedMeal"
                type="planned"
            />
        </template>
    </div>
</template>
