<script setup lang="ts">
import { computed, ref } from "vue";
import { Plus, X } from "lucide-vue-next";
import { Button, Card, IconButton, Input, List, ListRow } from "~/shared/ui";
import { addListItem } from "../domain/routine";

// Warmup moves or stretches: a short list with remove buttons and an add field under it.
// Changes are emitted straight away; the page saves them (no Save button).
const props = defineProps<{ items: string[]; label: string; placeholder: string; disabled?: boolean }>();
const emit = defineEmits<{ change: [items: string[]] }>();

const draft = ref("");
const canAdd = computed(() => addListItem(props.items, draft.value) !== props.items);

const add = () => {
    const next = addListItem(props.items, draft.value);
    draft.value = "";
    if (next !== props.items) emit("change", next);
};
const remove = (index: number) => emit("change", props.items.filter((_, i) => i !== index));
</script>

<template>
    <div class="flex flex-col gap-2">
        <Card v-if="items.length" padding="none">
            <List>
                <ListRow v-for="(item, index) in items" :key="item">
                    {{ item }}
                    <template #end>
                        <IconButton :label="`Remove ${item}`" size="sm" variant="danger" :disabled="disabled" @click="remove(index)">
                            <X />
                        </IconButton>
                    </template>
                </ListRow>
            </List>
        </Card>
        <form class="flex gap-2" @submit.prevent="add">
            <div class="min-w-0 flex-1">
                <Input v-model="draft" :aria-label="`Add to ${label}`" :placeholder="placeholder" autocomplete="off" />
            </div>
            <Button type="submit" :disabled="!canAdd || disabled">
                <template #icon-left><Plus class="size-4" /></template>
                Add
            </Button>
        </form>
    </div>
</template>
