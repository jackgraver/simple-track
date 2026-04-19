<script setup lang="ts">
import { ChevronDown, ChevronUp, Pencil } from "lucide-vue-next";
import { useRouter } from "vue-router";
import { computed, ref } from "vue";
import MealCard from "./MealCard.vue";
import { useDietDayMealHandlers } from "./useDietDayMealHandlers";

const props = defineProps<{
    dateOffset: number;
}>();

const router = useRouter();

const { data, logPlannedMeal, logMeal, deleteLoggedMeal, editLogMeal } =
    useDietDayMealHandlers(() => props.dateOffset);

const editPlannedMeal = () => {
    router.push({
        name: "diet-edit-planned",
        query: { offset: String(props.dateOffset) },
    });
};

const start = ref(0);

const visibleItems = computed(() =>
    data.value?.day.plannedMeals.slice(start.value, start.value + 2),
);

function next() {
    start.value = Math.min(
        start.value + 1,
        (data?.value?.day?.plannedMeals?.length ?? 0) - 2,
    );
}
function prev() {
    start.value = Math.max(start.value - 1, 0);
}
</script>

<template>
    <div class="flex min-w-0 flex-1 flex-col gap-2 pt-2">
        <div class="flex flex-row items-center justify-between gap-2">
            <div class="flex w-full items-center gap-2">
                <h2 class="mb-0 flex-1 text-lg font-semibold">Planned</h2>
                <button
                    type="button"
                    class="flex items-center gap-1 rounded-md border px-2 py-1 text-sm font-medium text-zinc-200"
                    @click="editPlannedMeal"
                >
                    <Pencil :size="15" />
                    Edit
                </button>
            </div>
            <template v-if="data && data.day.plannedMeals.length > 2">
                <button type="button" aria-label="Previous" @click="prev">
                    <ChevronUp />
                </button>
                <button type="button" aria-label="Next" @click="next">
                    <ChevronDown />
                </button>
            </template>
        </div>
        <template v-if="data">
            <span
                v-if="visibleItems?.length === 0"
                class="text-zinc-500"
                >Nothing else planned.</span
            >
            <MealCard
                v-for="log in visibleItems"
                :key="log.ID"
                :meal="log.meal"
                :on-log-planned="logPlannedMeal"
                :on-log-edited="logMeal"
                :on-delete="deleteLoggedMeal"
                :on-edit="editLogMeal"
                type="planned"
            />
        </template>
    </div>
</template>
