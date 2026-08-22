<script setup lang="ts">
import { ref } from "vue";
import { createInvestmentAccountType } from "~/api/money/api";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    onResolve?: (created: boolean) => void;
    onCancel?: () => void;
}>();

const name = ref("");
const contributionStartYear = ref("");
const isSaving = ref(false);

async function submit() {
    if (!name.value.trim() || isSaving.value) return;
    isSaving.value = true;
    try {
        await createInvestmentAccountType({
            name: name.value.trim(),
            contribution_start_year: contributionStartYear.value
                ? Number(contributionStartYear.value)
                : undefined,
        });
        toast.push("Account type created", "success");
        props.onResolve?.(true);
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to create account type"), "error");
    } finally {
        isSaving.value = false;
    }
}
</script>

<template>
    <form class="flex min-w-70 flex-col gap-4" @submit.prevent="submit">
        <input v-model="name" required placeholder="Type name (e.g. TFSA)" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
        <input v-model="contributionStartYear" required min="1900" max="9999" type="number" placeholder="Contribution start year" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
        <div class="flex justify-end gap-2">
            <button type="button" class="rounded-md bg-secondBg px-3 py-2 text-sm hover:bg-thirdBg" @click="props.onCancel?.()">Cancel</button>
            <button type="submit" class="rounded-md bg-primary px-3 py-2 text-sm text-white disabled:opacity-50" :disabled="isSaving">Create type</button>
        </div>
    </form>
</template>
