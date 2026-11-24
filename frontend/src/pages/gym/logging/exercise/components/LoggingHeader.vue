<script setup lang="ts">
import { ArrowLeft, Menu } from "lucide-vue-next";
import { useSlots } from "vue";
import { useNavDrawer } from "~/shared/nav/mobile/useNavDrawer";
import { IconButton } from "~/shared/ui";

const { openDrawer } = useNavDrawer();

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
        <!-- Phone: the nav drawer (the tab bar is gone). Desktop: the sidebar has no link to the list, so Back stays. -->
        <IconButton label="Open menu" variant="secondary" class="lg:hidden" @click="openDrawer"><Menu /></IconButton>
        <IconButton label="Back" variant="secondary" class="max-lg:hidden" @click="emit('back')"><ArrowLeft /></IconButton>
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
