<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { ChevronDown, ChevronRight, Pencil, Trash2 } from "lucide-vue-next";
import {
    deleteContributionLimit,
    deleteContributionType,
    saveBirthYear,
    upsertContributionLimit,
} from "~/api/money/api";
import { useContributionTypes, useMoneyMutation } from "~/api/money/queries";
import type { ContributionType } from "~/api/money/types";
import { isAdmin } from "~/composables/auth/session";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import ModuleTitleBar, { type ModuleTitleLink } from "~/shared/ModuleTitleBar.vue";
import ContributionTypeDialog from "./components/ContributionTypeDialog.vue";
import PresetDialog from "./components/PresetDialog.vue";
import RoomBreakdown from "./components/RoomBreakdown.vue";
import RoomCard from "./components/RoomCard.vue";
import { formatMoney, parseAmount } from "./format";
import {
    Button,
    Badge,
    Card,
    EmptyState,
    ErrorState,
    Field,
    IconButton,
    Input,
    List,
    ListRow,
    Skeleton,
    useDelayedFlag,
} from "~/shared/ui";

const typesQuery = useContributionTypes();
const types = computed(() => typesQuery.data.value?.contribution_types ?? []);
const birthYear = computed(() => typesQuery.data.value?.birth_year ?? null);
const showSkeleton = useDelayedFlag(typesQuery.isPending);
const needsBirthYear = computed(
    () => birthYear.value === null && types.value.some((t) => t.eligibility_age !== null),
);

const expanded = ref(new Set<number>());
function toggleBreakdown(id: number) {
    const next = new Set(expanded.value);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    expanded.value = next;
}

// ---- Rule summary ----

function ruleChips(type: ContributionType): string[] {
    const chips: string[] = [];
    if (type.opening_room_year !== null && type.opening_room !== null) {
        chips.push(`Opening room ${formatMoney(type.opening_room)} on Jan 1, ${type.opening_room_year}`);
    } else if (type.room) {
        const why = [];
        if (type.start_year !== null) why.push(`start year ${type.start_year}`);
        if (type.eligibility_age !== null && birthYear.value !== null) why.push(`age ${type.eligibility_age} in ${birthYear.value + type.eligibility_age}`);
        chips.push(`Room since ${type.room.start_year}${why.length ? ` (${why.join(", ")})` : ""}`);
    }
    if (type.carry_forward_cap === null) chips.push("Unused room carries forward");
    else if (type.carry_forward_cap === 0) chips.push("Unused room is lost each year");
    else chips.push(`Carry forward up to ${formatMoney(type.carry_forward_cap)}`);
    if (type.lifetime_limit !== null) chips.push(`Lifetime limit ${formatMoney(type.lifetime_limit)}`);
    chips.push(type.withdrawals_restore_room ? "Withdrawals restored next year" : "Withdrawals don't restore room");
    return chips;
}

function limitRange(type: ContributionType, index: number) {
    // Limits are sorted newest first; each applies until the next newer one.
    const limit = type.limits[index];
    const newer = type.limits[index - 1];
    if (!newer) return `${limit.year} onward`;
    if (newer.year - 1 === limit.year) return `${limit.year}`;
    return `${limit.year}–${newer.year - 1}`;
}

// ---- Mutations ----

const limitInputs = reactive<Record<number, { year: string; amount: string }>>({});
function limitInput(id: number) {
    limitInputs[id] ??= { year: String(new Date().getFullYear()), amount: "" };
    return limitInputs[id];
}

const limitMutation = useMoneyMutation(
    (args: { id: number; year: number; amount: number }) =>
        upsertContributionLimit(args.id, args.year, args.amount),
    "Failed to save limit",
);
const removeLimitMutation = useMoneyMutation((args: { id: number; year: number }) =>
    deleteContributionLimit(args.id, args.year),
    "Failed to remove limit",
);
const deleteTypeMutation = useMoneyMutation(deleteContributionType, "Failed to delete");
const birthYearMutation = useMoneyMutation(saveBirthYear, "Failed to save birth year");

async function saveLimit(type: ContributionType) {
    const input = limitInput(type.ID);
    const year = parseAmount(input.year);
    const amount = parseAmount(input.amount);
    if (year === null || amount === null) return;
    try {
        await limitMutation.mutateAsync({ id: type.ID, year: Math.trunc(year), amount });
        input.amount = "";
        toast.push(`${type.name} limit for ${year} saved`, "success");
    } catch {
        // useMoneyMutation already toasted the error.
    }
}

async function removeLimit(type: ContributionType, year: number) {
    const confirmed = await dialogManager.confirm({
        title: "Remove annual limit?",
        tone: "danger",
        message: `The ${type.name} limit starting ${year} will be removed; earlier limits extend to cover those years.`,
        confirmText: "Remove",
    });
    if (!confirmed) return;
    try {
        await removeLimitMutation.mutateAsync({ id: type.ID, year });
    } catch {
        // useMoneyMutation already toasted the error.
    }
}

async function removeType(type: ContributionType) {
    const confirmed = await dialogManager.confirm({
        title: `Delete ${type.name}?`,
        tone: "danger",
        message: type.accounts.length
            ? `Unlink ${type.accounts.map((a) => a.name).join(", ")} first.`
            : "Its rules and limits will be deleted. Accounts and deposits are not affected.",
        confirmText: "Delete",
    });
    if (!confirmed) return;
    try {
        await deleteTypeMutation.mutateAsync(type.ID);
        toast.push(`${type.name} deleted`, "success");
    } catch {
        // useMoneyMutation already toasted the error.
    }
}

function editType(type: ContributionType) {
    void dialogManager.custom<boolean>({
        title: type.can_edit ? `${type.name} rules` : `${type.name}: your room`,
        component: ContributionTypeDialog,
        componentProps: { contributionType: type },
    });
}

function openPreset() {
    void dialogManager.custom<boolean>({
        title: "Add registered account type",
        component: PresetDialog,
        componentProps: { existingNames: types.value.map((t) => t.name) },
    });
}

function openCustom() {
    void dialogManager.custom<boolean>({
        title: "Custom account type",
        component: ContributionTypeDialog,
    });
}

const birthYearInput = ref("");
async function submitBirthYear() {
    const year = parseAmount(birthYearInput.value);
    if (year === null) return;
    try {
        await birthYearMutation.mutateAsync(Math.trunc(year));
        toast.push("Birth year saved", "success");
    } catch {
        // useMoneyMutation already toasted the error.
    }
}

const navLinks: ModuleTitleLink[] = [
    { name: "Add TFSA / RRSP / FHSA", onClick: openPreset },
    { name: "Custom type", onClick: openCustom },
];
</script>

<template>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
        <ModuleTitleBar title="Contribution room" :links="navLinks" />

        <Card v-if="typesQuery.error.value" padding="none">
            <ErrorState
                :error="typesQuery.error.value"
                title="Couldn't load contribution room"
                @retry="typesQuery.refetch()"
            />
        </Card>
        <div v-else-if="typesQuery.isPending.value" class="flex flex-col gap-6" aria-busy="true">
            <template v-if="showSkeleton">
                <Skeleton shape="block" height="14rem" />
            </template>
        </div>

        <template v-else>
            <Card v-if="needsBirthYear">
                <form class="flex flex-col gap-3 lg:flex-row lg:items-end" @submit.prevent="submitBirthYear">
                    <div class="lg:max-w-md lg:flex-1">
                        <Field label="Your birth year is needed to know when room started accruing" required>
                            <Input v-model="birthYearInput" inputmode="numeric" required placeholder="e.g. 1998" />
                        </Field>
                    </div>
                    <Button type="submit" variant="primary" :loading="birthYearMutation.isPending.value">Save</Button>
                </form>
            </Card>

            <Card v-if="types.length === 0" padding="none">
                <EmptyState
                    title="No registered accounts yet"
                    description="Add a TFSA, RRSP or FHSA with CRA's rules and limits filled in, then link your accounts to it. Every deposit and withdrawal in a linked account counts against the shared room."
                >
                    <template #action>
                        <div class="flex flex-wrap items-center justify-center gap-2">
                            <Button variant="primary" @click="openPreset">Add TFSA / RRSP / FHSA</Button>
                            <Button @click="openCustom">Custom type</Button>
                        </div>
                    </template>
                </EmptyState>
            </Card>

            <Card v-for="type in types" :key="type.ID">
              <div class="flex flex-col gap-4">
                <header class="flex flex-col gap-2">
                    <div class="flex items-center justify-between gap-3">
                        <h2 class="m-0 min-w-0 truncate text-lg font-semibold text-fg">{{ type.name }}</h2>
                        <div class="flex shrink-0 items-center gap-1">
                            <Button size="sm" @click="editType(type)">
                                <template #icon-left><Pencil aria-hidden="true" class="size-4" /></template>
                                {{ type.can_edit ? "Rules" : "My room" }}
                            </Button>
                            <IconButton v-if="isAdmin" label="Delete type" variant="danger" size="sm" @click="removeType(type)">
                                <Trash2 />
                            </IconButton>
                        </div>
                    </div>
                    <ul class="m-0 flex list-none flex-wrap gap-1.5 p-0">
                        <li v-for="chip in ruleChips(type)" :key="chip"><Badge>{{ chip }}</Badge></li>
                    </ul>
                </header>

                <div class="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,18rem)_1fr]">
                    <RoomCard :contribution-type="type" bare class="self-start" />

                    <div class="flex flex-col gap-3">
                        <div>
                            <h3 class="m-0 text-sm font-semibold text-fg">Annual limits</h3>
                            <p class="m-0 text-xs text-fg-muted">Each limit applies from its year until a newer one replaces it.</p>
                        </div>
                        <p v-if="type.limits.length === 0" class="m-0 text-sm text-fg-muted">
                            No annual limits{{ type.opening_room !== null ? ", only the opening room counts." : "." }}
                        </p>
                        <List v-else>
                            <ListRow
                                v-for="(limit, index) in type.limits"
                                :key="limit.ID"
                                :pending="removeLimitMutation.isPending.value && removeLimitMutation.variables.value?.year === limit.year && removeLimitMutation.variables.value?.id === type.ID"
                            >
                                <span class="text-fg-muted">{{ limitRange(type, index) }}</span>
                                <template #trailing>
                                    <span class="text-fg">{{ formatMoney(limit.annual_limit) }} / yr</span>
                                    <IconButton
                                        v-if="isAdmin"
                                        :label="`Remove ${limit.year} limit`"
                                        size="sm"
                                        variant="danger"
                                        @click="removeLimit(type, limit.year)"
                                    >
                                        <Trash2 />
                                    </IconButton>
                                </template>
                            </ListRow>
                        </List>
                        <form v-if="type.can_edit" class="grid grid-cols-[6rem_minmax(0,1fr)_auto] items-end gap-2" @submit.prevent="saveLimit(type)">
                            <Field label="Year" required>
                                <Input v-model="limitInput(type.ID).year" inputmode="numeric" required />
                            </Field>
                            <Field label="Annual limit" required>
                                <Input v-model="limitInput(type.ID).amount" inputmode="decimal" required placeholder="0.00" />
                            </Field>
                            <Button type="submit" :loading="limitMutation.isPending.value">Set</Button>
                        </form>

                        <div v-if="type.accounts.length" class="text-sm">
                            <h3 class="m-0 text-sm font-semibold text-fg">Linked accounts</h3>
                            <div class="mt-1 flex flex-wrap gap-2">
                                <Button
                                    v-for="account in type.accounts"
                                    :key="account.id"
                                    size="sm"
                                    :to="{ name: 'finance-account', params: { id: account.id } }"
                                >
                                    {{ account.name }}
                                </Button>
                            </div>
                        </div>
                        <p v-else class="m-0 text-xs text-fg-muted">
                            No accounts linked. Edit an account and choose {{ type.name }} as its registered account.
                        </p>
                    </div>
                </div>

                <div v-if="type.room && type.room.years.length">
                    <Button variant="ghost" size="sm" :aria-expanded="expanded.has(type.ID)" @click="toggleBreakdown(type.ID)">
                        <template #icon-left>
                            <component :is="expanded.has(type.ID) ? ChevronDown : ChevronRight" aria-hidden="true" class="size-4" />
                        </template>
                        Year-by-year breakdown
                    </Button>
                    <div v-if="expanded.has(type.ID)" class="mt-3">
                        <RoomBreakdown :room="type.room" />
                    </div>
                </div>
              </div>
            </Card>
        </template>
    </div>
</template>
