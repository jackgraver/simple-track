<script setup lang="ts">
import { ArrowLeft } from "lucide-vue-next";
import { useSlots } from "vue";
import { IconButton } from "~/shared/ui";

defineProps<{
    title?: string;
    subtitle?: string;
}>();

const emit = defineEmits<{
    (e: "back"): void;
}>();

const slots = useSlots();
</script>

<template>
    <div class="grid w-full grid-cols-[auto_1fr_auto] items-center gap-2 border-b border-line pb-2">
        <IconButton label="Back" variant="secondary" @click="emit('back')"><ArrowLeft /></IconButton>
        <div class="min-w-0 text-center" :class="{ 'flex min-h-10 flex-col justify-center': subtitle }">
            <slot v-if="slots.center" name="center" />
            <template v-else-if="title">
                <h2 class="m-0 truncate font-medium text-fg" :class="subtitle ? 'text-base leading-5' : 'text-lg'">
                    {{ title }}
                </h2>
                <span v-if="subtitle" class="text-xs leading-4 text-fg-muted">{{ subtitle }}</span>
            </template>
        </div>
        <div class="min-w-0 shrink-0 text-right text-sm text-fg"><slot name="right" /></div>
    </div>
</template>
