<script setup lang="ts">
import type { LoggedSetWithStatus } from "../../store/useWorkoutStore";
import { Loader, X, RotateCcw } from "lucide-vue-next";

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
    <div v-if="loggedSets.length > 0" class="sets-logged">
        <h3>Logged Today</h3>
        <ul class="sets-list">
            <li
                v-for="(set, index) in loggedSets"
                :key="set.tempId || index"
                :class="[
                    'set-item',
                    {
                        clickable: set.status === 'success',
                        'set-item-error': set.status === 'error',
                    },
                ]"
                @click="set.status === 'success' && emit('edit', index)"
            >
                <div class="set-info">
                    <span class="set-summary">
                        Set {{ index + 1 }}: {{ set.weight }}lbs ×
                        {{ set.reps }}
                    </span>
                    <div class="set-actions">
                        <div v-if="set.status === 'pending'" class="set-status">
                            <Loader class="spinner" :size="16" />
                        </div>
                        <div
                            v-else-if="set.status === 'error'"
                            class="set-status"
                        >
                            <button
                                class="retry-button"
                                type="button"
                                title="Retry"
                                @click.stop="emit('retry', index)"
                            >
                                <RotateCcw :size="16" />
                            </button>
                        </div>
                        <button
                            v-if="set.status === 'success'"
                            class="delete-button"
                            type="button"
                            title="Delete set"
                            @click.stop="emit('delete', index)"
                        >
                            <X :size="16" />
                        </button>
                    </div>
                </div>
            </li>
        </ul>
    </div>
</template>

<style scoped>
.sets-logged {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.sets-logged h3 {
    margin: 0;
    font-size: 1rem;
    color: rgb(150, 150, 150);
}

.sets-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.sets-list li {
    padding: 0.5rem;
    background: rgb(27, 27, 27);
    border-radius: 3px;
    border: 1px solid rgb(56, 56, 56);
    transition:
        background-color 0.2s,
        border-color 0.2s;
}

.sets-list li.set-item.clickable {
    cursor: pointer;
}

.sets-list li.set-item.clickable:hover {
    background: rgb(35, 35, 35);
    border-color: rgb(100, 100, 100);
}

.sets-list li.set-item-error {
    border-color: rgb(132, 49, 49);
    background: rgb(37, 22, 22);
}

.set-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
}

.set-summary {
    min-width: 0;
    white-space: nowrap;
}

.set-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-shrink: 0;
}

.set-status {
    display: flex;
    align-items: center;
    justify-content: center;
}

.spinner {
    animation: spin 1s linear infinite;
    color: rgb(150, 150, 150);
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(360deg);
    }
}

.retry-button {
    background: transparent;
    border: none;
    color: rgb(255, 100, 100);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25rem;
    border-radius: 3px;
    transition: background-color 0.2s;
}

.retry-button:hover {
    background: rgb(40, 20, 20);
}

.delete-button {
    background: transparent;
    border: none;
    box-shadow: none;
    color: rgb(200, 100, 100);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25rem;
    border-radius: 3px;
    transition:
        background-color 0.2s,
        color 0.2s;
}

.delete-button:hover {
    background: rgb(40, 20, 20);
    color: rgb(255, 100, 100);
}
</style>
