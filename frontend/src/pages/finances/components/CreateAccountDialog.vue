<script setup lang="ts">
import { ref } from "vue";
import {
    createInvestmentAccount,
    type InvestmentAccountType,
} from "~/api/money/api";
import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    accountTypes: InvestmentAccountType[];
    onResolve?: (created: boolean) => void;
    onCancel?: () => void;
}>();

const name = ref("");
const accountTypeId = ref("");
const currentBalance = ref("");
const isSaving = ref(false);

async function submit() {
    if (!name.value.trim() || currentBalance.value === "" || isSaving.value) return;
    isSaving.value = true;
    try {
        await createInvestmentAccount({
            name: name.value.trim(),
            current_balance: Number(currentBalance.value),
            investment_account_type_id: accountTypeId.value
                ? Number(accountTypeId.value)
                : undefined,
        });
        toast.push("Investment account created", "success");
        props.onResolve?.(true);
    } catch (error) {
        toast.push(getApiErrorMessage(error, "Failed to create account"), "error");
    } finally {
        isSaving.value = false;
    }
}
</script>

<template>
    <form class="flex min-w-70 flex-col gap-4" @submit.prevent="submit">
        <input v-model="name" required placeholder="Account name" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
        <select v-model="accountTypeId" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none">
            <option value="">No account type</option>
            <option v-for="type in accountTypes" :key="type.ID" :value="type.ID">{{ type.name }}</option>
        </select>
        <input v-model="currentBalance" required min="0" step="0.01" type="number" placeholder="Current balance" class="rounded-md bg-thirdBg px-3 py-2 text-sm outline-none" />
        <div class="flex justify-end gap-2">
            <button type="button" class="rounded-md bg-secondBg px-3 py-2 text-sm hover:bg-thirdBg" @click="props.onCancel?.()">Cancel</button>
            <button type="submit" class="rounded-md bg-primary px-3 py-2 text-sm text-white disabled:opacity-50" :disabled="isSaving">Create account</button>
        </div>
    </form>
</template>
