<script setup lang="ts">
import { computed, ref } from "vue";
import { Plus, X } from "lucide-vue-next";

const props = defineProps<{
    title: string;
    description: string;
    modelValue: string[];
    placeholder: string;
}>();

const emit = defineEmits<{
    (event: "update:modelValue", value: string[]): void;
}>();

const newItem = ref("");
const trimmedItem = computed(() => newItem.value.trim());

const addItem = () => {
    const item = trimmedItem.value;
    if (!item) return;
    const exists = props.modelValue.some(
        (current) => current.toLowerCase() === item.toLowerCase(),
    );
    if (exists) return;
    emit("update:modelValue", [...props.modelValue, item]);
    newItem.value = "";
};

const removeItem = (index: number) => {
    emit(
        "update:modelValue",
        props.modelValue.filter((_, itemIndex) => itemIndex !== index),
    );
};
</script>

<template>
    <section
        class="flex flex-col gap-3 rounded-lg border border-(--color-border) bg-firstBg p-4"
    >
        <div class="flex flex-col gap-1">
            <h2 class="m-0 text-sm font-semibold text-textPrimary">
                {{ title }}
            </h2>
            <p class="m-0 text-xs text-textSecondary">{{ description }}</p>
        </div>
        <div v-if="modelValue.length" class="flex flex-col gap-1">
            <div
                v-for="(item, index) in modelValue"
                :key="`${item}-${index}`"
                class="flex items-center gap-3 rounded-md bg-secondBg px-3 py-2 text-sm"
            >
                <span class="min-w-0 flex-1 text-textPrimary">{{ item }}</span>
                <button
                    type="button"
                    class="shrink-0 rounded p-1 text-textSecondary transition-colors hover:bg-firstBg hover:text-(--color-cf-red)"
                    :aria-label="`Remove ${item}`"
                    @click="removeItem(index)"
                >
                    <X class="size-4" />
                </button>
            </div>
        </div>
        <p v-else class="m-0 text-sm italic text-textSecondary">
            No items configured.
        </p>
        <div class="flex gap-2">
            <input
                v-model="newItem"
                type="text"
                class="min-w-0 flex-1 rounded-md border border-(--color-border) bg-firstBg px-3 py-2 text-sm text-textPrimary placeholder:text-textSecondary/60 focus:outline-none focus:ring-2 focus:ring-(--color-cf-red)/40"
                :placeholder="placeholder"
                @keydown.enter.prevent="addItem"
            />
            <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-md bg-secondBg px-3 py-2 text-sm font-medium text-textPrimary transition-colors hover:bg-thirdBg disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="!trimmedItem"
                @click="addItem"
            >
                <Plus class="size-4" />
                Add
            </button>
        </div>
    </section>
</template>
