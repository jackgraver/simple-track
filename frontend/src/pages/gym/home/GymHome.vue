<script setup lang="ts">
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import WorkoutActivityGraph from "~/pages/gym/WorkoutActivityGraph.vue";
import { preloadRouteChunks } from "~/router/preloadChunks";
import ModuleTitleBar from "~/shared/ModuleTitleBar.vue";
import { Card } from "~/shared/ui";
import InsightsSection from "./InsightsSection.vue";
import TodayCard from "./TodayCard.vue";
import WeekStrip from "./WeekStrip.vue";
import { useGymWeek } from "./useGymWeek";

// Only the gym home reads the week (programs + 52-week activity), so sub-pages don't fetch it.
const props = defineProps<{ offset: number }>();
const router = useRouter();

const week = useGymWeek(() => props.offset);

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
    <ModuleTitleBar title="Gym" :links="[]" />
    <WeekStrip :week="week" :offset="offset" @update:offset="setOffset" />
    <TodayCard :offset="offset" :week="week" />
    <InsightsSection />
    <Card padding="sm">
        <WorkoutActivityGraph />
    </Card>
</template>
