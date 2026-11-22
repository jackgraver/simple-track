<script setup lang="ts">
import type { MobilityLogged } from "~/types/workout";
import { ListChecks } from "lucide-vue-next";
import { Button, Card, Checkbox, EmptyState, List } from "~/shared/ui";
import LoggingHeader from "./LoggingHeader.vue";
import { ref, watch, computed } from "vue";
import { useLoggingRouteContext } from "../../composables/useLoggingRouteContext";
import { useRoute } from "vue-router";

const props = defineProps<{
    loggedPreMobility: MobilityLogged | null;
    loggedPostMobility: MobilityLogged | null;
    savePreMobility: (checked: string[]) => void;
    savePostMobility: (checked: string[]) => void;
}>();

const route = useRoute();
const { goBackToLogging } = useLoggingRouteContext();

const mobilitySlot = computed<"pre" | "post" | null>(() => {
    const raw = route.params.slot;
    const value = typeof raw === "string" ? raw : "";
    if (value === "pre" || value === "post") return value;
    return null;
});

const effectiveLogged = computed(() =>
    mobilitySlot.value === "post"
        ? props.loggedPostMobility
        : props.loggedPreMobility,
);

const localChecked = ref<Set<string>>(new Set());

watch(
    effectiveLogged,
    (m) => {
        localChecked.value = new Set(m?.checked ?? []);
    },
    { immediate: true },
);

const items = computed(() => effectiveLogged.value?.items ?? []);
const title = computed(() => effectiveLogged.value?.title ?? "");

const doneCount = computed(
    () => items.value.filter((name) => localChecked.value.has(name)).length,
);

const headerProgress = computed(() => {
    const n = items.value.length;
    return n ? `${doneCount.value}/${n}` : "";
});

const toggle = (name: string) => {
    const next = new Set(localChecked.value);
    if (next.has(name)) {
        next.delete(name);
    } else {
        next.add(name);
    }
    localChecked.value = next;
    // The cache updates at once; a failed save rolls it back and `effectiveLogged` resets the ticks.
    if (mobilitySlot.value === "post") props.savePostMobility([...next]);
    else props.savePreMobility([...next]);
};

const isChecked = (name: string) => localChecked.value.has(name);
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <LoggingHeader :title="title" @back="goBackToLogging">
            <template #right>
                <span v-if="headerProgress" class="text-sm font-medium text-fg-muted tabular-nums">{{ headerProgress }}</span>
            </template>
        </LoggingHeader>
        <EmptyState v-if="items.length === 0" :icon="ListChecks" title="No items in this routine" />
        <Card v-else padding="none">
            <List>
                <li v-for="item in items" :key="item" class="px-3 lg:px-4">
                    <Checkbox
                        :model-value="isChecked(item)"
                        :label="item"
                        @update:model-value="toggle(item)"
                    />
                </li>
            </List>
        </Card>
        <Button variant="primary" size="lg" block @click="goBackToLogging">Done</Button>
    </div>
</template>
