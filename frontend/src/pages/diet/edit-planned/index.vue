<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { parseDietDayOffsetQuery } from "~/utils/dateUtil";
import { PageHeader } from "~/shared/ui";
import PlanCalendar from "./PlanCalendar.vue";
import PlannedDayEditor from "./PlannedDayEditor.vue";
import BulkRangePlanner from "./BulkRangePlanner.vue";

const route = useRoute();
const router = useRouter();

const dateOffset = computed(() => parseDietDayOffsetQuery(route.query.offset));

function setOffset(offset: number) {
    router.replace({ query: { ...route.query, offset: String(offset) } });
}
</script>

<template>
    <div class="mx-auto flex w-full max-w-3xl flex-col gap-6 pt-2 pb-above-tabbar lg:pb-10">
        <PageHeader title="Edit planned meals" />
        <PlanCalendar :date-offset="dateOffset" @select="setOffset" />
        <PlannedDayEditor :date-offset="dateOffset" />
        <BulkRangePlanner />
    </div>
</template>
