<script setup lang="ts">
import { computed, ref } from "vue";
import { Plus, X } from "lucide-vue-next";
import { Button, Card, IconButton, Input, List, ListRow, Section } from "~/shared/ui";

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
    <Card>
        <Section :title="title">
            <p class="m-0 -mt-2 text-xs text-fg-muted">{{ description }}</p>
            <Card v-if="modelValue.length" padding="none" class="bg-surface-2">
                <List>
                    <ListRow v-for="(item, index) in modelValue" :key="`${item}-${index}`">
                        {{ item }}
                        <template #trailing>
                            <IconButton :label="`Remove ${item}`" size="sm" variant="danger" @click="removeItem(index)">
                                <X />
                            </IconButton>
                        </template>
                    </ListRow>
                </List>
            </Card>
            <p v-else class="m-0 text-sm text-fg-muted">No items configured.</p>
            <div class="flex gap-2">
                <div class="min-w-0 flex-1">
                    <Input
                        v-model="newItem"
                        :aria-label="`${title}: new item`"
                        :placeholder="placeholder"
                        @keydown.enter.prevent="addItem"
                    />
                </div>
                <Button :disabled="!trimmedItem" @click="addItem">
                    <template #icon-left><Plus class="size-4" /></template>
                    Add
                </Button>
            </div>
        </Section>
    </Card>
</template>
