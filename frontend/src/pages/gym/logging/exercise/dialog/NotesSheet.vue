<script setup lang="ts">
import { ref } from "vue";
import { Button, Field, Textarea } from "~/shared/ui";

// Notes live off the main screen; they're saved with the exercise on the next log or finish.
const props = defineProps<{
    notes: string;
    onResolve?: (value: string) => void;
}>();

const draft = ref(props.notes);
</script>

<template>
    <div class="flex w-full min-w-0 flex-col gap-4">
        <Field label="Notes">
            <Textarea v-model="draft" :rows="4" placeholder="Left shoulder tight, try seat 5 next time" />
        </Field>
        <Button variant="primary" size="lg" block @click="onResolve?.(draft)">Save notes</Button>
    </div>
</template>
