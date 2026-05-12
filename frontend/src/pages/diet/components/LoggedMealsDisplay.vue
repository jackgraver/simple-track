<script setup lang="ts">
import { ChevronDown, ChevronUp } from "lucide-vue-next";
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import MealCard from "./MealCard.vue";
import { useDietDayMealHandlers } from "./useDietDayMealHandlers";

const props = defineProps<{
    dateOffset: number;
}>();

const { data, logPlannedMeal, logMeal, deleteLoggedMeal } =
    useDietDayMealHandlers(() => props.dateOffset);

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
    const len = (data.value?.day.loggedMeals ?? []).length;
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

const loggedMeals = computed(() => data.value?.day.loggedMeals ?? []);

watch(
    () => loggedMeals.value.length,
    () => {
        recalcMaxVisible();
    },
);

watch(
    () => [loggedMeals.value.length, maxVisible.value],
    ([len, max]) => {
        const maxStart = Math.max(0, len - max);
        if (start.value > maxStart) start.value = maxStart;
    },
);

const visibleItems = computed(() =>
    isMobile.value
        ? loggedMeals.value
        : loggedMeals.value.slice(start.value, start.value + maxVisible.value),
);

watch(visibleItems, () => {
    void nextTick(() => {
        checkOverflow();
    });
});

const showChevrons = computed(
    () => !isMobile.value && loggedMeals.value.length > maxVisible.value,
);

function next() {
    start.value = Math.min(
        start.value + 1,
        Math.max(0, loggedMeals.value.length - maxVisible.value),
    );
}
function prev() {
    start.value = Math.max(start.value - 1, 0);
}
</script>

<template>
    <div ref="containerRef" class="flex min-w-0 flex-1 flex-col gap-2 pt-2">
        <div class="flex flex-row items-center justify-between gap-2">
            <h2 class="mb-0 flex-1 text-lg font-semibold">
                Logged ({{ loggedMeals.length }})
            </h2>
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
                    :disabled="start >= loggedMeals.length - maxVisible"
                    class="shrink-0 disabled:text-zinc-500"
                    aria-label="Next"
                    @click="next"
                >
                    <ChevronDown />
                </button>
            </template>
        </div>
        <template v-if="data">
            <span v-if="loggedMeals.length === 0" class="text-zinc-500"
                >Nothing logged yet.</span
            >
            <MealCard
                v-for="log in visibleItems"
                :key="log.ID"
                :meal="log.meal"
                :on-log-planned="logPlannedMeal"
                :on-log-edited="logMeal"
                :on-delete="deleteLoggedMeal"
                type="logged"
            />
        </template>
    </div>
</template>
