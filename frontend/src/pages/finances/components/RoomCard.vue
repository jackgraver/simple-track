<script setup lang="ts">
import { computed } from "vue";
import { AlertTriangle } from "lucide-vue-next";
import type { ContributionType } from "~/api/money/types";
import { Badge, Card } from "~/shared/ui";
import { formatMoney, yearAllowance } from "../format";

const props = defineProps<{
    contributionType: ContributionType;
    compact?: boolean;
    /** Hide the name/accounts header when the page already shows them. */
    bare?: boolean;
}>();

const room = computed(() => props.contributionType.room);
const currentYear = computed(() => room.value?.years[0]);
/** This year's own limit; null for types without annual limits. */
const allowance = computed(() => (currentYear.value ? yearAllowance(currentYear.value) : null));

function percent(used: number, total: number) {
    if (total <= 0) return used > 0 ? 100 : 0;
    return Math.min(100, Math.max(0, (used / total) * 100));
}

/** Share of this year's limit (or, without one, of all available room) used. */
const usedPercent = computed(() => {
    const year = currentYear.value;
    if (!year) return 0;
    if (allowance.value) return percent(allowance.value.contributed, allowance.value.allowance);
    return percent(year.contributed, year.available);
});
const barColor = computed(() => {
    if (room.value?.over_contributed) return "bg-danger";
    if (allowance.value?.over) return "bg-warn";
    return "bg-success";
});
</script>

<template>
    <component :is="bare ? 'div' : Card">
        <div class="flex flex-col gap-3">
        <div v-if="!bare || room?.over_contributed || allowance?.over" class="flex items-start justify-between gap-2">
            <div v-if="!bare" class="min-w-0">
                <h3 class="m-0 text-sm font-semibold text-fg">{{ contributionType.name }}</h3>
                <p class="m-0 truncate text-xs text-fg-muted">
                    {{
                        contributionType.accounts.length
                            ? contributionType.accounts.map((a) => a.name).join(", ")
                            : "No accounts linked"
                    }}
                </p>
            </div>
            <Badge v-if="room?.over_contributed" tone="danger" title="Contributions exceed all available room">
                <AlertTriangle aria-hidden="true" class="size-3" /> Over
            </Badge>
            <Badge
                v-else-if="allowance?.over && currentYear"
                tone="warn"
                title="Contributed more than this year's limit, using room carried forward from earlier years"
            >
                <AlertTriangle aria-hidden="true" class="size-3" /> Over {{ currentYear.year }} limit
            </Badge>
        </div>

        <template v-if="room && currentYear">
            <template v-if="allowance">
                <div>
                    <p class="m-0 text-xs text-fg-muted">Left to contribute in {{ currentYear.year }}</p>
                    <p class="m-0 text-2xl font-semibold tabular-nums" :class="allowance.left < 0 ? 'text-danger' : 'text-fg'">
                        {{ formatMoney(allowance.left) }}
                    </p>
                </div>
                <div class="flex flex-col gap-1">
                    <div class="h-1.5 w-full overflow-hidden rounded-full bg-surface-3">
                        <div class="h-full rounded-full" :class="barColor" :style="{ width: `${usedPercent}%` }" />
                    </div>
                    <p class="m-0 text-xs text-fg-muted">
                        {{ formatMoney(allowance.contributed) }} of {{ formatMoney(allowance.allowance) }} this year
                        <template v-if="allowance.restored > 0">
                            ({{ formatMoney(allowance.limit) }} limit + {{ formatMoney(allowance.restored) }} restored)
                        </template>
                        <template v-else>({{ currentYear.year }} limit)</template>
                    </p>
                </div>
                <dl class="m-0 grid grid-cols-2 gap-2 text-xs">
                    <div>
                        <dt class="text-fg-muted">Total room incl. past years</dt>
                        <dd class="m-0 font-medium tabular-nums" :class="room.available < 0 ? 'text-danger' : ''">
                            {{ formatMoney(room.available) }}
                        </dd>
                    </div>
                    <div v-if="room.lifetime_remaining !== null">
                        <dt class="text-fg-muted">Lifetime left</dt>
                        <dd class="m-0 font-medium tabular-nums">{{ formatMoney(room.lifetime_remaining) }}</dd>
                    </div>
                    <template v-if="!compact">
                        <div v-if="currentYear.carried_in !== 0">
                            <dt class="text-fg-muted">Carried in from {{ currentYear.year - 1 }}</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(currentYear.carried_in) }}</dd>
                        </div>
                        <div v-if="currentYear.withdrawn > 0">
                            <dt class="text-fg-muted">Withdrawn {{ currentYear.year }}</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(currentYear.withdrawn) }}</dd>
                        </div>
                    </template>
                </dl>
            </template>

            <template v-else>
                <div>
                    <p class="m-0 text-xs text-fg-muted">Available to contribute</p>
                    <p class="m-0 text-2xl font-semibold tabular-nums" :class="room.available < 0 ? 'text-danger' : 'text-fg'">
                        {{ formatMoney(room.available) }}
                    </p>
                </div>
                <div class="flex flex-col gap-1">
                    <div class="h-1.5 w-full overflow-hidden rounded-full bg-surface-3">
                        <div class="h-full rounded-full" :class="barColor" :style="{ width: `${usedPercent}%` }" />
                    </div>
                    <p class="m-0 text-xs text-fg-muted">
                        {{ formatMoney(currentYear.contributed) }} of {{ formatMoney(currentYear.available) }} used in {{ currentYear.year }}
                    </p>
                </div>
                <dl v-if="!compact || room.lifetime_remaining !== null" class="m-0 grid grid-cols-2 gap-2 text-xs">
                    <div v-if="room.lifetime_remaining !== null">
                        <dt class="text-fg-muted">Lifetime left</dt>
                        <dd class="m-0 font-medium tabular-nums">{{ formatMoney(room.lifetime_remaining) }}</dd>
                    </div>
                    <template v-if="!compact">
                        <div>
                            <dt class="text-fg-muted">New room {{ currentYear.year }}</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(currentYear.new_room) }}</dd>
                        </div>
                        <div v-if="currentYear.restored > 0">
                            <dt class="text-fg-muted">Restored withdrawals</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(currentYear.restored) }}</dd>
                        </div>
                        <div v-if="currentYear.withdrawn > 0">
                            <dt class="text-fg-muted">Withdrawn {{ currentYear.year }}</dt>
                            <dd class="m-0 font-medium tabular-nums">{{ formatMoney(currentYear.withdrawn) }}</dd>
                        </div>
                    </template>
                </dl>
            </template>
        </template>
        <p v-else class="m-0 text-xs text-warn">
            {{ contributionType.room_issue || "Room can't be calculated yet." }}
        </p>
        </div>
    </component>
</template>
