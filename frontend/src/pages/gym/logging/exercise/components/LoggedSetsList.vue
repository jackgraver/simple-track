<script setup lang="ts">
import type { LoggedSetWithStatus } from "../../store/useWorkoutStore";
import { RotateCcw, X } from "lucide-vue-next";
import { IconButton, List, ListRow, Section, Spinner } from "~/shared/ui";

defineProps<{
    loggedSets: LoggedSetWithStatus[];
}>();

const emit = defineEmits<{
    retry: [index: number];
    delete: [index: number];
    edit: [index: number];
}>();
</script>

<template>
    <Section v-if="loggedSets.length > 0" title="Logged today">
        <List class="rounded-card border border-line bg-surface">
            <ListRow
                v-for="(set, index) in loggedSets"
                :key="set.tempId || index"
                :pending="set.status === 'pending'"
                v-bind="set.status === 'success' ? { onClick: () => emit('edit', index) } : {}"
            >
                Set {{ index + 1 }}: {{ set.weight }}lbs × {{ set.reps }}
                <template v-if="set.weight_setup?.trim()" #subtitle>{{ set.weight_setup }}</template>
                <template v-if="set.status === 'error'" #trailing>
                    <span class="text-xs text-danger">Failed to save</span>
                </template>
                <template v-else-if="set.status === 'pending'" #trailing><Spinner size="sm" /></template>
                <template v-if="set.status === 'error'" #end>
                    <IconButton label="Retry saving set" variant="danger" @click="emit('retry', index)">
                        <RotateCcw />
                    </IconButton>
                </template>
                <template v-else-if="set.status === 'success'" #end>
                    <IconButton label="Delete set" variant="danger" @click="emit('delete', index)"><X /></IconButton>
                </template>
            </ListRow>
        </List>
    </Section>
</template>
