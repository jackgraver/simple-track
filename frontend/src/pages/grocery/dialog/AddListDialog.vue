<script setup lang="ts">
import axios from "axios";
import { computed, nextTick, onMounted, ref, useTemplateRef } from "vue";
import {
    useCreateGroceryItems,
    useParseGroceryList,
} from "~/api/tracking/queries";
import { toast } from "~/composables/toast/useToast";

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
const textArea = useTemplateRef<HTMLTextAreaElement>("textArea");

const parseMutation = useParseGroceryList();
const createMutation = useCreateGroceryItems();

const selectedNames = computed(() =>
    (rows.value ?? [])
        .filter((r) => r.include && r.name.trim())
        .map((r) => r.name.trim()),
);

onMounted(() => textArea.value?.focus());

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
    textArea.value?.focus();
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
    <div class="flex w-full flex-col gap-3 text-left text-zinc-100">
        <template v-if="rows === null">
            <p class="m-0 text-xs text-zinc-400">
                Paste a list — one item per line, or separated by commas.
                Items are matched to things you've bought before.
            </p>
            <textarea
                ref="textArea"
                v-model="text"
                rows="8"
                placeholder="milk&#10;2 dozen eggs&#10;bananas, bread"
                class="w-full resize-y rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-base text-zinc-100 outline-none focus:border-zinc-400"
                @keydown="onTextKeydown"
            />
            <div class="flex justify-end gap-2">
                <button
                    type="button"
                    class="rounded-md bg-zinc-800 px-4 py-2 text-sm text-zinc-200 hover:bg-zinc-700"
                    @click="props.onCancel?.()"
                >
                    Cancel
                </button>
                <button
                    type="button"
                    class="rounded-md bg-zinc-600 px-4 py-2 text-sm font-semibold text-zinc-100 hover:bg-zinc-500 disabled:opacity-50"
                    :disabled="!text.trim() || parseMutation.isPending.value"
                    @click="review"
                >
                    {{ parseMutation.isPending.value ? "Reading…" : "Review" }}
                </button>
            </div>
        </template>

        <template v-else>
            <p class="m-0 text-xs text-zinc-400">
                Check the names, untick anything you don't want.
            </p>
            <ul class="m-0 flex max-h-[50vh] list-none flex-col gap-1 overflow-y-auto p-0">
                <li
                    v-for="(row, i) in rows"
                    :key="i"
                    class="flex items-start gap-3 py-1"
                    :class="{ 'opacity-50': !row.include }"
                >
                    <input
                        v-model="row.include"
                        type="checkbox"
                        class="mt-1.5 size-4 shrink-0 cursor-pointer"
                        :aria-label="`Include ${row.name}`"
                    />
                    <div class="flex min-w-0 flex-1 flex-col">
                        <input
                            v-model="row.name"
                            type="text"
                            autocomplete="off"
                            class="w-full min-w-0 rounded border border-transparent bg-transparent px-1 py-0.5 text-base text-zinc-100 outline-none hover:border-zinc-700 focus:border-zinc-500"
                            :aria-label="`Name for ${row.input}`"
                        />
                        <span
                            v-if="row.onList"
                            class="px-1 text-xs text-amber-400/80"
                        >
                            Already on your list
                        </span>
                        <span
                            v-else-if="row.matched && row.name !== row.input"
                            class="truncate px-1 text-xs text-zinc-500"
                        >
                            from “{{ row.input }}”
                        </span>
                        <span
                            v-else-if="!row.matched"
                            class="px-1 text-xs text-zinc-500"
                        >
                            New item
                        </span>
                    </div>
                </li>
            </ul>
            <div class="flex justify-end gap-2">
                <button
                    type="button"
                    class="rounded-md bg-zinc-800 px-4 py-2 text-sm text-zinc-200 hover:bg-zinc-700"
                    @click="back"
                >
                    Back
                </button>
                <button
                    type="button"
                    class="rounded-md bg-zinc-600 px-4 py-2 text-sm font-semibold text-zinc-100 hover:bg-zinc-500 disabled:opacity-50"
                    :disabled="selectedNames.length === 0 || createMutation.isPending.value"
                    @click="addSelected"
                >
                    {{
                        createMutation.isPending.value
                            ? "Adding…"
                            : `Add ${selectedNames.length} item${selectedNames.length === 1 ? "" : "s"}`
                    }}
                </button>
            </div>
        </template>
    </div>
</template>
