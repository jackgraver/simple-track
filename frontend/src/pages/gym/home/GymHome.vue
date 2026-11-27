<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useQueryClient } from "@tanstack/vue-query";
import { offsetToDate } from "~/api/workout/dates";
import { workoutQueries } from "~/api/workout/queries";
import { useRouter } from "vue-router";
import WorkoutActivityGraph from "~/pages/gym/WorkoutActivityGraph.vue";
import { preloadRouteChunks } from "~/router/preloadChunks";
import { Card } from "~/shared/ui";
import InsightsSection from "./InsightsSection.vue";
import TodayCard from "./TodayCard.vue";
import WeekStrip from "./WeekStrip.vue";
import { useGymWeek } from "./useGymWeek";

// Only the gym home reads the week (programs + 52-week activity), so sub-pages don't fetch it.
const props = defineProps<{ offset: number }>();
const router = useRouter();

const week = useGymWeek(() => props.offset);

// The Today card trails the week strip: it keeps showing the day it has until the newly picked day
// has loaded, then swaps in one step. Otherwise it empties while the day loads, which flashes and
// shifts everything below. ensureQueryData shares the card's own request (same key) and is instant
// for cached days.
const qc = useQueryClient();
const shownOffset = ref(props.offset);
const shownWeek = useGymWeek(() => shownOffset.value);
watch(
    () => props.offset,
    (o) => {
        const show = () => {
            if (props.offset === o) shownOffset.value = o;
        };
        // On failure, swap anyway so the card shows its error state for that day.
        qc.ensureQueryData(workoutQueries.day(offsetToDate(o))).then(show, show);
    },
);

const setOffset = (offset: number) => {
    void router.replace({ name: "gym", query: offset === 0 ? {} : { offset: String(offset) } });
};

// Where the home page leads (start workout -> logging -> exercise, schedule -> plan day): fetch
// their code while idle so those taps don't wait on a chunk download.
onMounted(() =>
    preloadRouteChunks(router, ["logging", "logging-exercise", "gym-plans", "gym-program-day-detail"]),
);
</script>

<template>
    <WeekStrip class="lg:mt-10" :week="week" :offset="offset" @update:offset="setOffset" />
    <TodayCard :offset="shownOffset" :week="shownWeek" />
    <InsightsSection />
    <Card padding="sm">
        <WorkoutActivityGraph />
    </Card>
</template>
