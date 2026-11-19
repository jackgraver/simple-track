<script setup lang="ts">
import { computed, ref } from "vue";
import Chart from "~/shared/chart/Chart.vue";
import { AlertTriangle, ChevronRight, Landmark, Plus, RefreshCw } from "lucide-vue-next";
import {
    useAccounts,
    useContributionTypes,
    useNetWorthHistory,
} from "~/api/money/queries";
import type { Account, AccountKind } from "~/api/money/types";
import { dialogManager } from "~/composables/dialog/useDialog";
import ModuleTitleBar, { type ModuleTitleLink } from "~/shared/ModuleTitleBar.vue";
import AccountFormDialog from "./components/AccountFormDialog.vue";
import UpdateBalancesDialog from "./components/UpdateBalancesDialog.vue";
import RoomCard from "./components/RoomCard.vue";
import { buildMoneyLineChart } from "./chart";
import {
    STALE_BALANCE_DAYS,
    accountKindLabels,
    dayKey,
    daysSince,
    formatAge,
    formatMoney,
    formatSignedMoney,
    changeClass,
} from "./format";
import {
    Badge,
    Button,
    Card,
    EmptyState,
    ErrorState,
    List,
    ListRow,
    Section,
    SegmentedControl,
    Skeleton,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";
import { useRouter } from "vue-router";

const router = useRouter();
const showArchived = ref(false);

const accountsQuery = useAccounts(showArchived);
const historyQuery = useNetWorthHistory();
const typesQuery = useContributionTypes();

const accounts = computed(() => accountsQuery.data.value ?? []);
const activeAccounts = computed(() => accounts.value.filter((a) => !a.archived_at));
const archivedAccounts = computed(() => accounts.value.filter((a) => a.archived_at));
const contributionTypes = computed(() => typesQuery.data.value?.contribution_types ?? []);

const assets = computed(() =>
    activeAccounts.value.filter((a) => !a.is_liability).reduce((sum, a) => sum + a.current_balance, 0),
);
const liabilities = computed(() =>
    activeAccounts.value.filter((a) => a.is_liability).reduce((sum, a) => sum + a.current_balance, 0),
);
const netWorth = computed(() => assets.value - liabilities.value);

const staleAccounts = computed(() =>
    activeAccounts.value.filter(
        (a) => !a.balance_date || daysSince(a.balance_date) > STALE_BALANCE_DAYS,
    ),
);

// ---- Net worth history ----

type Range = "3m" | "1y" | "all";
const range = ref<Range>("1y");
const rangeOptions: { value: Range; label: string }[] = [
    { value: "3m", label: "3M" },
    { value: "1y", label: "1Y" },
    { value: "all", label: "All" },
];

function cutoffFor(value: Range) {
    if (value === "all") return "";
    const d = new Date();
    d.setMonth(d.getMonth() - (value === "3m" ? 3 : 12));
    return d.toISOString().slice(0, 10);
}

const history = computed(() => historyQuery.data.value ?? []);
const visibleHistory = computed(() => {
    const cutoff = cutoffFor(range.value);
    const inRange = history.value.filter((p) => dayKey(p.date) >= cutoff);
    // Anchor the start of the range with the last point before it so the
    // line doesn't start mid-air.
    const earlier = history.value.filter((p) => dayKey(p.date) < cutoff);
    const before = earlier[earlier.length - 1];
    return before ? [before, ...inRange] : inRange;
});
const rangeChange = computed(() => {
    const points = visibleHistory.value;
    if (points.length < 2) return null;
    return points[points.length - 1].net_worth - points[0].net_worth;
});
const chart = computed(() =>
    buildMoneyLineChart(
        visibleHistory.value.map((p) => ({ date: p.date, value: p.net_worth })),
        "Net worth",
    ),
);

// ---- Account groups ----

const groups: { title: string; kinds: AccountKind[] }[] = [
    { title: "Cash & banking", kinds: ["chequing", "savings", "cash"] },
    { title: "Investments", kinds: ["investment"] },
    { title: "Property & other", kinds: ["property", "other"] },
    { title: "Liabilities", kinds: ["credit_card", "loan"] },
];

const groupedAccounts = computed(() =>
    groups
        .map((group) => {
            const list = activeAccounts.value.filter((a) => group.kinds.includes(a.kind));
            return {
                ...group,
                accounts: list,
                total: list.reduce((sum, a) => sum + a.current_balance, 0),
            };
        })
        .filter((group) => group.accounts.length > 0),
);

function balanceChange(account: Account) {
    if (account.previous_balance === null) return null;
    const diff = Math.round((account.current_balance - account.previous_balance) * 100) / 100;
    return diff === 0 ? null : diff;
}

function accountChangeClass(account: Account, diff: number) {
    return changeClass(diff > 0 !== account.is_liability);
}

function accountHref(id: number) {
    return router.resolve({ name: "finance-account", params: { id } }).href;
}

function isStale(account: Account) {
    return !account.balance_date || daysSince(account.balance_date) > STALE_BALANCE_DAYS;
}

// ---- Actions ----

async function openAddAccount() {
    const id = await dialogManager.custom<number>({
        title: "Add account",
        component: AccountFormDialog,
        componentProps: { contributionTypes: contributionTypes.value },
    });
    if (typeof id === "number") {
        void router.push({ name: "finance-account", params: { id } });
    }
}

function openUpdateBalances() {
    void dialogManager.custom<boolean>({
        title: "Update balances",
        component: UpdateBalancesDialog,
        componentProps: { accounts: activeAccounts.value },
    });
}

const navLinks = computed<ModuleTitleLink[]>(() => [
    { name: "Update balances", onClick: openUpdateBalances },
    { name: "Add account", onClick: openAddAccount },
]);

const isLoading = computed(() => accountsQuery.isPending.value);
const showSkeleton = useDelayedFlag(isLoading);
const loadError = computed(() => accountsQuery.error.value);
</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <ModuleTitleBar title="Finances" :links="navLinks" />

        <Card v-if="loadError" padding="none">
            <ErrorState :error="loadError" title="Couldn't load accounts" @retry="accountsQuery.refetch()" />
        </Card>
        <div v-else-if="isLoading" class="flex flex-col gap-6" :aria-busy="true">
            <template v-if="showSkeleton">
                <Skeleton shape="block" height="16rem" />
                <Card padding="none"><SkeletonList :rows="4" /></Card>
            </template>
        </div>

        <Card v-else-if="activeAccounts.length === 0 && !showArchived" padding="none">
            <EmptyState
                :icon="Landmark"
                title="Track every account in one place"
                description="Add your chequing, savings, investments, credit cards and loans, then update balances whenever you check them to build a net worth history. Link TFSA / RRSP / FHSA accounts to see your contribution room."
            >
                <template #action>
                    <div class="flex flex-wrap items-center justify-center gap-2">
                        <Button variant="primary" @click="openAddAccount">Add your first account</Button>
                        <Button :to="{ name: 'finance-rules' }">Set up contribution room</Button>
                        <Button variant="ghost" @click="showArchived = true">Show archived</Button>
                    </div>
                </template>
            </EmptyState>
        </Card>

        <template v-else>
            <!-- Net worth -->
            <Card>
              <div class="flex flex-col gap-4">
                <div class="flex flex-wrap items-end justify-between gap-4">
                    <div>
                        <p class="m-0 text-sm text-fg-muted">Net worth</p>
                        <p class="m-0 text-3xl font-semibold text-fg tabular-nums">{{ formatMoney(netWorth) }}</p>
                        <p v-if="rangeChange !== null" class="m-0 mt-1 text-sm tabular-nums" :class="changeClass(rangeChange >= 0)">
                            {{ formatSignedMoney(rangeChange) }}
                            <span class="text-fg-muted">over {{ range === "all" ? "all time" : range === "1y" ? "the past year" : "3 months" }}</span>
                        </p>
                    </div>
                    <dl class="m-0 flex gap-6 text-sm">
                        <div>
                            <dt class="text-fg-muted">Assets</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(assets) }}</dd>
                        </div>
                        <div>
                            <dt class="text-fg-muted">Liabilities</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(liabilities) }}</dd>
                        </div>
                    </dl>
                </div>
                <div v-if="history.length > 1" class="flex flex-col gap-3">
                    <div class="lg:w-56 lg:self-end">
                        <SegmentedControl v-model="range" :options="rangeOptions" label="Net worth range" />
                    </div>
                    <div class="h-56 w-full min-w-0">
                        <Chart type="line" :data="chart.data" :options="chart.options" class="h-full w-full" />
                    </div>
                </div>
                <p v-else class="m-0 text-xs text-fg-muted">
                    Record balances on different days to build a net worth history.
                </p>
              </div>
            </Card>

            <!-- Stale balances -->
            <Card
                v-if="staleAccounts.length"
                as="button"
                interactive
                padding="sm"
                class="text-sm"
                @click="openUpdateBalances"
            >
                <span class="flex min-h-touch items-center gap-3 lg:min-h-row">
                    <AlertTriangle aria-hidden="true" class="size-5 shrink-0 text-warn" />
                    <span class="flex-1">
                        {{ staleAccounts.length }} balance{{ staleAccounts.length === 1 ? " hasn't" : "s haven't" }} been updated in over a month
                    </span>
                    <span class="flex items-center gap-1 font-medium">Update <RefreshCw aria-hidden="true" class="size-4" /></span>
                </span>
            </Card>

            <!-- Contribution room -->
            <Section v-if="contributionTypes.length" title="Contribution room">
                <template #actions>
                    <Button variant="ghost" size="sm" :to="{ name: 'finance-rules' }">
                        Details
                        <template #icon-right><ChevronRight aria-hidden="true" class="size-4" /></template>
                    </Button>
                </template>
                <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
                    <RoomCard v-for="type in contributionTypes" :key="type.ID" :contribution-type="type" compact />
                </div>
            </Section>

            <!-- Accounts -->
            <Section v-for="group in groupedAccounts" :key="group.title" :title="group.title">
                <template #actions>
                    <span class="text-sm text-fg-muted tabular-nums">{{ formatMoney(group.total) }}</span>
                </template>
                <Card padding="none">
                    <List>
                        <ListRow v-for="account in group.accounts" :key="account.ID" :to="accountHref(account.ID)">
                            <span class="flex items-center gap-2">
                                <span class="truncate">{{ account.name }}</span>
                                <Badge v-if="account.contribution_type">{{ account.contribution_type.name }}</Badge>
                            </span>
                            <template #subtitle>
                                {{ account.institution || accountKindLabels[account.kind] }}
                                ·
                                <span :class="isStale(account) ? 'text-warn' : ''">
                                    {{ account.balance_date ? `updated ${formatAge(account.balance_date)}` : "no balance yet" }}
                                </span>
                            </template>
                            <template #trailing>
                                <span class="flex flex-col items-end">
                                    <span class="font-medium text-fg">{{ formatMoney(account.current_balance) }}</span>
                                    <span v-if="balanceChange(account) !== null" class="text-xs" :class="accountChangeClass(account, balanceChange(account)!)">
                                        {{ formatSignedMoney(balanceChange(account)!) }}
                                    </span>
                                </span>
                                <ChevronRight aria-hidden="true" class="size-4 text-fg-subtle" />
                            </template>
                        </ListRow>
                    </List>
                </Card>
            </Section>

            <Section v-if="showArchived && archivedAccounts.length" title="Archived">
                <Card padding="none" class="opacity-70">
                    <List>
                        <ListRow v-for="account in archivedAccounts" :key="account.ID" :to="accountHref(account.ID)">
                            {{ account.name }}
                            <template #trailing>{{ formatMoney(account.current_balance) }}</template>
                        </ListRow>
                    </List>
                </Card>
            </Section>

            <div class="flex flex-wrap items-center justify-between gap-2">
                <Button variant="ghost" @click="showArchived = !showArchived">
                    {{ showArchived ? "Hide archived accounts" : "Show archived accounts" }}
                </Button>
                <Button @click="openAddAccount">
                    <template #icon-left><Plus aria-hidden="true" class="size-4" /></template>
                    Add account
                </Button>
            </div>
        </template>
    </div>
</template>
