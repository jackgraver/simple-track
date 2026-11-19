<script setup lang="ts">
import axios from "axios";
import { onMounted, ref, useTemplateRef } from "vue";
import { useImportWeightCsv } from "~/api/tracking/queries";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    onResolve?: (v?: unknown) => void;
    onCancel?: () => void;
}>();

const text = ref("");
const textArea = useTemplateRef<HTMLTextAreaElement>("textArea");
const importMutation = useImportWeightCsv();

onMounted(() => textArea.value?.focus());

function errorMessage(err: unknown, fallback: string) {
    if (axios.isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined;
        if (data?.error) return data.error;
    }
    return err instanceof Error ? err.message : fallback;
}

async function submit() {
    if (!text.value.trim() || importMutation.isPending.value) return;
    try {
        const res = await importMutation.mutateAsync(text.value);
        toast.push(
            `Imported ${res.created} new, ${res.updated} updated`,
            "success",
        );
        props.onResolve?.(true);
    } catch (err) {
        toast.push(errorMessage(err, "Import failed"), "error");
    }
}

function onTextKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        void submit();
    }
}
</script>

<template>
    <div class="flex w-full flex-col gap-3 text-left text-zinc-100">
        <p class="m-0 text-xs text-zinc-400">
            Paste the CSV copied from the Renpho app. Days that already exist
            are overwritten.
        </p>
        <textarea
            ref="textArea"
            v-model="text"
            rows="8"
            placeholder="No.,Date,Time,Weight(lb),..."
            class="w-full resize-y rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 font-mono text-base text-zinc-100 outline-none focus:border-zinc-400"
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
                :disabled="!text.trim() || importMutation.isPending.value"
                @click="submit"
            >
                {{ importMutation.isPending.value ? "Importing…" : "Import" }}
            </button>
        </div>
    </div>
</template>
