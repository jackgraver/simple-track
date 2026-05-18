<script setup lang="ts">
import axios from "axios";
import { computed, onMounted, ref } from "vue";
import { apiGET, apiPOST } from "~/api/client";
import { toast } from "~/composables/toast/useToast";
import { formatDateLong } from "~/utils/dateUtil";

type Account = {
    id: number;
    name: string;
    balance: number;
};
type TransactionCategory = { id: number; name: string };
type Transaction = {
    id: number;
    account_id: number;
    amount: number;
    date: string;
    category_id: number;
    account?: Account;
    category?: TransactionCategory;
};

function todayYmd(): string {
    const d = new Date();
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
}

const accounts = ref<Account[]>([]);
const transactions = ref<Transaction[]>([]);
const categories = ref<TransactionCategory[]>([]);
const loading = ref(true);
const accountName = ref("");
const accountBalance = ref("");
const txAccountId = ref<number | "">("");
const txAmount = ref("");
const txCategoryId = ref<number | "">("");
const txDate = ref(todayYmd());
const errMsg = (err: unknown): string => {
    if (
        axios.isAxiosError(err) &&
        err.response?.data &&
        typeof err.response.data === "object" &&
        "error" in err.response.data
    ) {
        const e0 = (err.response.data as { error?: string }).error;
        if (e0) return e0;
    }
    if (err instanceof Error) return err.message;
    return "Request failed";
};
const loadAccounts = async () => {
    const data = await apiGET<{ accounts: Account[] }>("/finance/accounts");
    accounts.value = data.accounts ?? [];
};
const loadTransactions = async () => {
    const data = await apiGET<{ transactions: Transaction[] }>(
        "/finance/transactions",
    );
    transactions.value = data.transactions ?? [];
};
const loadCategories = async () => {
    const data = await apiGET<{ categories: TransactionCategory[] }>(
        "/finance/categories",
    );
    categories.value = data.categories ?? [];
    if (txCategoryId.value === "") {
        const unc = categories.value.find((c) => c.name === "Uncategorized");
        const first = categories.value[0];
        const pick = unc ?? first;
        txCategoryId.value = pick ? pick.id : "";
    }
};
const refresh = async () => {
    loading.value = true;
    try {
        await Promise.all([
            loadAccounts(),
            loadTransactions(),
            loadCategories(),
        ]);
    } catch (err: unknown) {
        toast.push(errMsg(err), "error");
    } finally {
        loading.value = false;
    }
};
onMounted(refresh);
const createAccount = async () => {
    const name = accountName.value.trim();
    if (!name) {
        toast.push("Account name is required", "error");
        return;
    }
    const bal = Number.parseFloat(accountBalance.value.replace(",", ".")) || 0;
    try {
        await apiPOST("/finance/accounts", { name, balance: bal });
        toast.push("Account created", "success");
        accountName.value = "";
        accountBalance.value = "";
        await loadAccounts();
    } catch (err: unknown) {
        toast.push(errMsg(err), "error");
    }
};
const createTransaction = async () => {
    const aid =
        typeof txAccountId.value === "number"
            ? txAccountId.value
            : Number(txAccountId.value);
    const amt = Number.parseFloat(txAmount.value.replace(",", "."));
    const cid =
        typeof txCategoryId.value === "number"
            ? txCategoryId.value
            : Number(txCategoryId.value);

    console.log(aid, amt, cid, txDate.value);
    if (!aid || Number.isNaN(amt)) {
        toast.push("Pick an account and enter a valid amount", "error");
        return;
    }
    if (!cid || Number.isNaN(cid)) {
        toast.push("Pick a category", "error");
        return;
    }
    const dateStr = txDate.value.trim();
    if (!dateStr) {
        toast.push("Date is required", "error");
        return;
    }
    try {
        await apiPOST("/finance/transactions", {
            account_id: aid,
            amount: amt,
            date: dateStr,
            category_id: cid,
        });
        toast.push("Transaction created", "success");
        txAmount.value = "";
        txDate.value = todayYmd();
        const unc = categories.value.find((c) => c.name === "Uncategorized");
        txCategoryId.value = unc?.id ?? categories.value[0]?.id ?? "";
        await loadTransactions();
        await loadAccounts();
    } catch (err: unknown) {
        toast.push(errMsg(err), "error");
    }
};
const hasAccounts = computed(() => accounts.value.length > 0);
const hasCategories = computed(() => categories.value.length > 0);
const canSubmitTransaction = computed(
    () => hasAccounts.value && hasCategories.value,
);
</script>
<template>
    <div class="flex w-full min-w-0 flex-col gap-8">
        <div
            class="flex items-center justify-between gap-4 border-b border-(--color-border) pb-3"
        >
            <h1 class="m-0 text-lg font-semibold text-textPrimary">Finance</h1>
        </div>
        <p v-if="loading" class="m-0 text-sm text-textSecondary">Loading…</p>
        <template v-else>
            <section class="flex flex-col gap-3">
                <h2 class="m-0 text-base font-semibold text-textPrimary">
                    Accounts
                </h2>
                <div
                    class="overflow-x-auto rounded-md border border-(--color-border)"
                >
                    <table class="w-full border-collapse text-left text-sm">
                        <thead>
                            <tr
                                class="border-b border-(--color-border) bg-secondBg"
                            >
                                <th
                                    class="px-3 py-2 font-medium text-textSecondary"
                                >
                                    Name
                                </th>
                                <th
                                    class="px-3 py-2 font-medium text-textSecondary"
                                >
                                    Balance
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr
                                v-for="a in accounts"
                                :key="a.id"
                                class="border-b border-(--color-border)"
                            >
                                <td class="px-3 py-2 text-textPrimary">
                                    {{ a.name }}
                                </td>
                                <td class="px-3 py-2 text-textPrimary">
                                    {{ a.balance }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                    <p
                        v-if="!accounts.length"
                        class="m-0 px-3 py-4 text-sm text-textSecondary"
                    >
                        No accounts yet.
                    </p>
                </div>
            </section>
            <section class="flex flex-col gap-3">
                <h2 class="m-0 text-base font-semibold text-textPrimary">
                    New account
                </h2>
                <form
                    class="flex max-w-md flex-col gap-3 rounded-md border border-(--color-border) bg-firstBg p-4"
                    @submit.prevent="createAccount"
                >
                    <label
                        class="flex flex-col gap-1 text-xs text-textSecondary"
                        >Name
                        <input
                            v-model="accountName"
                            type="text"
                            autocomplete="off"
                            class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                        />
                    </label>
                    <label
                        class="flex flex-col gap-1 text-xs text-textSecondary"
                        >Starting balance
                        <input
                            v-model="accountBalance"
                            type="text"
                            inputmode="decimal"
                            autocomplete="off"
                            placeholder="0"
                            class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                        />
                    </label>
                    <button
                        type="submit"
                        class="rounded-md bg-secondBg px-4 py-2 text-sm font-semibold text-textPrimary transition-colors hover:bg-thirdBg"
                    >
                        Create account
                    </button>
                </form>
            </section>
            <section class="flex flex-col gap-3">
                <h2 class="m-0 text-base font-semibold text-textPrimary">
                    Transactions
                </h2>
                <div
                    class="overflow-x-auto rounded-md border border-(--color-border)"
                >
                    <table class="w-full border-collapse text-left text-sm">
                        <thead>
                            <tr
                                class="border-b border-(--color-border) bg-secondBg"
                            >
                                <th
                                    class="px-3 py-2 font-medium text-textSecondary"
                                >
                                    Date
                                </th>
                                <th
                                    class="px-3 py-2 font-medium text-textSecondary"
                                >
                                    Account
                                </th>
                                <th
                                    class="px-3 py-2 font-medium text-textSecondary"
                                >
                                    Category
                                </th>
                                <th
                                    class="px-3 py-2 font-medium text-textSecondary"
                                >
                                    Amount
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr
                                v-for="t in transactions"
                                :key="t.id"
                                class="border-b border-(--color-border)"
                            >
                                <td class="px-3 py-2 text-textPrimary">
                                    {{ formatDateLong(t.date) }}
                                </td>
                                <td class="px-3 py-2 text-textPrimary">
                                    {{ t.account?.name ?? "—" }}
                                </td>
                                <td class="px-3 py-2 text-textPrimary">
                                    {{ t.category?.name ?? "—" }}
                                </td>
                                <td class="px-3 py-2 text-textPrimary">
                                    {{ t.amount }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                    <p
                        v-if="!transactions.length"
                        class="m-0 px-3 py-4 text-sm text-textSecondary"
                    >
                        No transactions yet.
                    </p>
                </div>
            </section>
            <section class="flex flex-col gap-3">
                <h2 class="m-0 text-base font-semibold text-textPrimary">
                    New transaction
                </h2>
                <form
                    class="flex max-w-md flex-col gap-3 rounded-md border border-(--color-border) bg-firstBg p-4"
                    @submit.prevent="createTransaction"
                >
                    <label
                        class="flex flex-col gap-1 text-xs text-textSecondary"
                        >Account
                        <select
                            v-model="txAccountId"
                            class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                            :disabled="!canSubmitTransaction"
                        >
                            <option disabled value="">Select…</option>
                            <option
                                v-for="a in accounts"
                                :key="a.id"
                                :value="a.id"
                            >
                                {{ a.name }}
                            </option>
                        </select>
                    </label>
                    <label
                        class="flex flex-col gap-1 text-xs text-textSecondary"
                        >Category
                        <select
                            v-model="txCategoryId"
                            required
                            class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                            :disabled="!canSubmitTransaction"
                        >
                            <option disabled value="">Select…</option>
                            <option
                                v-for="cat in categories"
                                :key="cat.id"
                                :value="cat.id"
                            >
                                {{ cat.name }}
                            </option>
                        </select>
                    </label>
                    <label
                        class="flex flex-col gap-1 text-xs text-textSecondary"
                        >Amount (negative for expense)
                        <input
                            v-model="txAmount"
                            type="text"
                            inputmode="decimal"
                            autocomplete="off"
                            class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                        />
                    </label>
                    <label
                        class="flex flex-col gap-1 text-xs text-textSecondary"
                        >Date
                        <input
                            v-model="txDate"
                            type="date"
                            required
                            class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary"
                        />
                    </label>
                    <button
                        type="submit"
                        class="rounded-md bg-secondBg px-4 py-2 text-sm font-semibold text-textPrimary transition-colors hover:bg-thirdBg disabled:opacity-50"
                        :disabled="!canSubmitTransaction"
                    >
                        Create transaction
                    </button>
                    <p
                        v-if="!hasAccounts"
                        class="m-0 text-xs text-textSecondary"
                    >
                        Add an account first.
                    </p>
                    <p
                        v-else-if="!hasCategories"
                        class="m-0 text-xs text-textSecondary"
                    >
                        No categories available.
                    </p>
                </form>
            </section>
        </template>
    </div>
</template>
