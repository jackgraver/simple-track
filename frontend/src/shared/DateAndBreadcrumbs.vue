<script lang="ts" setup>  
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-vue-next";
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { formatDateLong } from "~/utils/dateUtil";

const route = useRoute();
const router = useRouter();

const dayOffset = computed(() => {
    const raw = route.query.offset;
    const value = typeof raw === "string" ? Number.parseInt(raw, 10) : 0;
    return Number.isNaN(value) ? 0 : value;
});

const updateOffset = (nextOffset: number) => {
    const nextQuery = { ...route.query };
    if (nextOffset === 0) {
        delete nextQuery.offset;
    } else {
        nextQuery.offset = String(nextOffset);
    }

    router.push({
        path: route.path,
        query: nextQuery,
        hash: route.hash,
    });
};


const goToPreviousDay = () => {
    updateOffset(dayOffset.value + 1);
};

const goToNextDay = () => {
    updateOffset(dayOffset.value - 1);
};

const isToday = computed(() => dayOffset.value === 0);

const goToToday = () => {
    if (!isToday.value) updateOffset(0);
};

/** Mirrors backend `utils.ZerodTime(offset)` (local calendar day minus offset). */
const selectedCalendarDate = computed(() => {
    const now = new Date();
    return new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate() - dayOffset.value,
        12,
        0,
        0,
        0,
    );
});

const dateLabel = computed(() => {
    const d = selectedCalendarDate.value;
    const ymd = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T12:00:00`;
    return formatDateLong(ymd);
});

const selectedYmd = computed(() => {
    const d = selectedCalendarDate.value;
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
});

const nativeDateInputEl = ref<HTMLInputElement | null>(null);

function openNativeDatePicker() {
    const el = nativeDateInputEl.value;
    if (!el) return;
    if (typeof el.showPicker === "function") {
        try {
            void el.showPicker();
        } catch {
            el.click();
        }
    } else {
        el.click();
    }
}

function onNativeDateChange(ev: Event) {
    const raw = (ev.target as HTMLInputElement).value;
    if (!raw) return;
    const parts = raw.split("-").map((s) => Number.parseInt(s, 10));
    if (
        parts.length !== 3 ||
        parts.some((n) => Number.isNaN(n))
    )
        return;
    const [y, m0, dom] = parts;
    const picked = new Date(y, m0 - 1, dom);
    const now = new Date();
    const todayMid = new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate(),
    );
    const pickedMid = new Date(
        picked.getFullYear(),
        picked.getMonth(),
        picked.getDate(),
    );
    const nextOffset = Math.round(
        (todayMid.getTime() - pickedMid.getTime()) / (24 * 60 * 60 * 1000),
    );
    updateOffset(nextOffset);
}
</script>

<template>
    <section class="flex flex-col gap-2">
        <div class="flex items-center justify-between gap-3">
            <button
                type="button"
                aria-label="Previous day"
                class="shrink-0 rounded-md border border-(--color-border) bg-firstBg p-2 text-textPrimary transition-colors hover:bg-secondBg"
                @click="goToPreviousDay"
            >
                <ChevronLeftIcon class="size-4" />
            </button>
            <div
                class="flex min-w-0 flex-1 flex-col items-center justify-center gap-0"
            >
                <div
                    class="grid w-full max-w-[min(100%,22rem)] place-items-center text-center text-pretty"
                >
                    <span
                        class="invisible col-start-1 row-start-1 m-0 max-w-full px-1 text-base font-medium leading-tight text-textPrimary"
                        aria-hidden="true"
                    >
                        Wednesday, December&nbsp;31st,&nbsp;2026
                    </span>
                    <button
                        type="button"
                        class="col-start-1 row-start-1 m-0 max-w-full cursor-pointer rounded border-none bg-transparent px-1 text-base font-medium leading-tight text-textPrimary underline-offset-2 transition-colors hover:text-textPrimary hover:underline focus:outline-none focus-visible:ring-2 focus-visible:ring-(--color-border)"
                        :aria-label="`Chosen day: ${dateLabel}. Open calendar`"
                        @click="openNativeDatePicker"
                    >
                        {{ dateLabel }}
                    </button>
                </div>
                <input
                    ref="nativeDateInputEl"
                    type="date"
                    class="sr-only"
                    tabindex="-1"
                    aria-hidden="true"
                    :value="selectedYmd"
                    @change="onNativeDateChange"
                />
                <div class="flex h-4 shrink-0 items-center justify-center leading-none">
                    <button
                        type="button"
                        class="m-0! py-0! text-xs leading-none text-textSecondary underline-offset-2 transition-colors hover:text-textPrimary hover:underline disabled:pointer-events-none"
                        :class="{ invisible: isToday }"
                        :disabled="isToday"
                        :tabindex="isToday ? -1 : 0"
                        :aria-hidden="isToday ? 'true' : undefined"
                        @click="goToToday"
                    >
                        Jump to today
                    </button>
                </div>
            </div>
            <button
                type="button"
                aria-label="Next day"
                class="shrink-0 rounded-md border border-(--color-border) bg-firstBg p-2 text-textPrimary transition-colors hover:bg-secondBg"
                @click="goToNextDay"
            >
                <ChevronRightIcon class="size-4" />
            </button>
        </div>
    </section>
</template>