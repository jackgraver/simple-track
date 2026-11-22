<script setup lang="ts">
import axios from "axios";
import { ref } from "vue";
import { useImportWeightCsv } from "~/api/tracking/queries";
import { toast } from "~/composables/toast/useToast";
import { Button, Textarea } from "~/shared/ui";

const props = defineProps<{
    onResolve?: (v?: unknown) => void;
    onCancel?: () => void;
}>();

const text = ref("");
const importMutation = useImportWeightCsv();

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
    <div class="flex w-full flex-col gap-3 text-left">
        <p class="m-0 text-xs text-fg-muted">
            Paste the CSV copied from the Renpho app. Days that already exist
            are overwritten.
        </p>
        <Textarea
            v-model="text"
            :rows="8"
            placeholder="No.,Date,Time,Weight(lb),..."
            class="font-mono"
            autofocus
            @keydown="onTextKeydown"
        />
        <div class="flex justify-end gap-2">
            <Button variant="ghost" @click="props.onCancel?.()">Cancel</Button>
            <Button
                variant="primary"
                :disabled="!text.trim()"
                :loading="importMutation.isPending.value"
                @click="submit"
            >
                {{ importMutation.isPending.value ? "Importing…" : "Import" }}
            </Button>
        </div>
    </div>
</template>
