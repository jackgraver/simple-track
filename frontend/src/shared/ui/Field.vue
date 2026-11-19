<script setup lang="ts">
import { AlertCircle } from "lucide-vue-next";
import { computed, useId } from "vue";
import { provideField } from "./fieldContext";

const props = defineProps<{ label: string; hint?: string; error?: string; required?: boolean }>();

const id = useId();
const hintId = `${id}-hint`;
const errorId = `${id}-error`;
const describedBy = computed(() => [props.error ? errorId : "", props.hint ? hintId : ""].filter(Boolean).join(" ") || undefined);

provideField({ id, describedBy: () => describedBy.value, invalid: () => !!props.error });
</script>

<template>
    <div class="flex flex-col gap-1.5">
        <label :for="id" class="text-xs font-medium text-fg-muted">
            {{ label }}<span v-if="required" aria-hidden="true" class="text-fg-subtle"> *</span>
        </label>
        <slot />
        <p v-if="hint && !error" :id="hintId" class="m-0 text-xs text-fg-subtle">{{ hint }}</p>
        <p v-if="error" :id="errorId" class="m-0 flex items-start gap-1.5 text-xs text-danger">
            <AlertCircle aria-hidden="true" class="mt-px size-3.5 shrink-0" />
            <span>{{ error }}</span>
        </p>
    </div>
</template>
