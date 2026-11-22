<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { parseDietDayOffsetQuery } from "~/utils/dateUtil";
import WorkoutActivityGraph from "~/pages/gym/WorkoutActivityGraph.vue";
import { gymDayOptions, useAdjacentDayPrefetch } from "~/api/prefetchDays";
import ModuleTitleBar from "~/shared/ModuleTitleBar.vue";
import { Card } from "~/shared/ui";
import InsightsSection from "./home/InsightsSection.vue";
import TodayCard from "./home/TodayCard.vue";
import WeekStrip from "./home/WeekStrip.vue";
import { useGymWeek } from "./home/useGymWeek";

const route = useRoute();
const router = useRouter();
const isGymHome = computed(() => route.name === "gym");
const dayOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));

useAdjacentDayPrefetch(dayOffset, gymDayOptions, isGymHome);

const week = useGymWeek(dayOffset);

const setOffset = (offset: number) => {
    void router.replace({ name: "gym", query: offset === 0 ? {} : { offset: String(offset) } });
};
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <template v-if="isGymHome">
            <ModuleTitleBar title="Gym" :links="[]" />
            <WeekStrip :week="week" :offset="dayOffset" @update:offset="setOffset" />
            <TodayCard :offset="dayOffset" :week="week" />
            <InsightsSection />
            <Card padding="sm">
                <WorkoutActivityGraph />
            </Card>
        </template>
        <router-view v-else />
    </div>
</template>
