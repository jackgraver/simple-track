<script setup lang="ts">
import { ArrowLeft } from "lucide-vue-next";
import { useSlots } from "vue";

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
    <div
        class="grid w-full grid-cols-[auto_1fr_auto] items-center gap-2 border-b border-zinc-700 pb-2 box-border"
    >
        <button
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded border border-zinc-600 bg-transparent text-inherit hover:bg-zinc-800"
            type="button"
            @click="emit('back')"
        >
            <ArrowLeft :size="20" />
        </button>
        <div
            class="min-w-0 text-center @container"
            :class="{ 'flex h-10 flex-col justify-center': subtitle }"
        >
            <slot v-if="slots.center" name="center" />
            <template v-else-if="title">
                <h2
                    class="m-0 truncate font-medium"
                    :class="subtitle ? 'text-base leading-5' : 'text-lg'"
                >
                    {{ title }}
                </h2>
                <span v-if="subtitle" class="text-xs leading-4 text-zinc-500">
                    {{ subtitle }}
                </span>
            </template>
        </div>
        <div class="min-w-0 shrink-0 text-right text-sm">
            <slot name="right" />
        </div>
    </div>
</template>
