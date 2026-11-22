<script setup lang="ts">
import axios from "axios";
import { computed, nextTick, onMounted, ref, useTemplateRef } from "vue";
import {
    useCreateGroceryItems,
    useParseGroceryList,
} from "~/api/tracking/queries";
import { toast } from "~/composables/toast/useToast";
import { Badge, Button, Checkbox, Input, List, ListRow, Textarea } from "~/shared/ui";

const props = defineProps<{
    onResolve?: (v?: unknown) => void;
    onCancel?: () => void;
}>();

type Row = {
    input: string;
    name: string;
    matched: boolean;
    onList: boolean;
    include: boolean;
};

const text = ref("");
const rows = ref<Row[] | null>(null);
const textWrap = useTemplateRef<HTMLElement>("textWrap");
const focusText = () => textWrap.value?.querySelector("textarea")?.focus();

const parseMutation = useParseGroceryList();
const createMutation = useCreateGroceryItems();

const selectedNames = computed(() =>
    (rows.value ?? [])
        .filter((r) => r.include && r.name.trim())
        .map((r) => r.name.trim()),
);

onMounted(focusText);

function errorMessage(err: unknown, fallback: string) {
    if (axios.isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined;
        if (data?.error) return data.error;
    }
    return err instanceof Error ? err.message : fallback;
}

async function review() {
    if (!text.value.trim() || parseMutation.isPending.value) return;
    try {
        const lines = await parseMutation.mutateAsync(text.value);
        if (lines.length === 0) {
            toast.push("No items found in that text", "error");
            return;
        }
        rows.value = lines.map((l) => ({
            input: l.input,
            name: l.name,
            matched: l.matched,
            onList: l.on_list,
            include: !l.on_list,
        }));
    } catch (err) {
        toast.push(errorMessage(err, "Failed to read list"), "error");
    }
}

async function back() {
    rows.value = null;
    await nextTick();
    focusText();
}

async function addSelected() {
    const names = selectedNames.value;
    if (names.length === 0 || createMutation.isPending.value) return;
    try {
        const { items, skipped } = await createMutation.mutateAsync(names);
        const added = `Added ${items.length} item${items.length === 1 ? "" : "s"}`;
        toast.push(
            skipped.length ? `${added} (${skipped.length} already on list)` : added,
            "success",
        );
        props.onResolve?.(true);
    } catch (err) {
        toast.push(errorMessage(err, "Failed to add items"), "error");
    }
}

function onTextKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        void review();
    }
}
</script>

<template>
    <div class="flex w-full flex-col gap-3 text-left">
        <template v-if="rows === null">
            <p class="m-0 text-sm text-fg-muted">
                Paste a list: one item per line, or separated by commas.
                Items are matched to things you've bought before.
            </p>
            <div ref="textWrap" @keydown="onTextKeydown">
                <Textarea
                    v-model="text"
                    :rows="8"
                    aria-label="List to add"
                    placeholder="milk&#10;eggs&#10;5 apples, bread"
                />
            </div>
            <div class="flex flex-col-reverse gap-2 lg:flex-row lg:justify-end">
                <Button variant="ghost" @click="props.onCancel?.()">Cancel</Button>
                <Button
                    variant="primary"
                    :disabled="!text.trim()"
                    :loading="parseMutation.isPending.value"
                    @click="review"
                >
                    {{ parseMutation.isPending.value ? "Reading…" : "Review" }}
                </Button>
            </div>
        </template>

        <template v-else>
            <p class="m-0 text-sm text-fg-muted">
                Check the names, untick anything you don't want.
            </p>
            <List class="max-h-[50vh] overflow-y-auto">
                <ListRow v-for="(row, i) in rows" :key="i" :disabled="!row.include">
                    <div class="flex items-center gap-3 py-1">
                        <Checkbox v-model="row.include" :label="`Include ${row.name}`">
                            <span class="sr-only">Include {{ row.name }}</span>
                        </Checkbox>
                        <div class="flex min-w-0 flex-1 flex-col gap-1">
                            <Input v-model="row.name" autocomplete="off" :aria-label="`Name for ${row.input}`" />
                            <Badge v-if="row.onList" tone="warn" class="self-start">Already on your list</Badge>
                            <span v-else-if="row.matched && row.name !== row.input" class="truncate text-xs text-fg-muted">
                                from “{{ row.input }}”
                            </span>
                            <span v-else-if="!row.matched" class="text-xs text-fg-muted">New item</span>
                        </div>
                    </div>
                </ListRow>
            </List>
            <div class="flex flex-col-reverse gap-2 lg:flex-row lg:justify-end">
                <Button variant="ghost" @click="back">Back</Button>
                <Button
                    variant="primary"
                    :disabled="selectedNames.length === 0"
                    :loading="createMutation.isPending.value"
                    @click="addSelected"
                >
                    {{ createMutation.isPending.value ? "Adding…" : `Add ${selectedNames.length} item${selectedNames.length === 1 ? "" : "s"}` }}
                </Button>
            </div>
        </template>
    </div>
</template>
