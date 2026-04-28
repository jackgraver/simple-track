<script setup lang="ts">
import { Bell } from "lucide-vue-next";
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useMissedTracking } from "~/api/tracking/queries";
import { isAuthenticated } from "~/composables/auth/session";

const { data, isPending, isError } = useMissedTracking();
const open = ref(false);
const root = ref<HTMLElement | null>(null);
function onDocClick(ev: MouseEvent) {
    const el = root.value;
    if (!el || el.contains(ev.target as Node)) return;
    open.value = false;
}
onMounted(() => document.addEventListener("click", onDocClick));
onUnmounted(() => document.removeEventListener("click", onDocClick));
const missedItems = computed(() => {
    const d = data.value;
    if (!d) return [];
    const q = { date: d.date };
    const items: {
        key: string;
        label: string;
        to: { name: string; query: { date: string } };
    }[] = [];
    if (d.steps)
        items.push({
            key: "steps",
            label: "Log steps",
            to: { name: "gym-steps", query: q },
        });
    if (d.weight)
        items.push({
            key: "weight",
            label: "Log weight",
            to: { name: "gym-weight", query: q },
        });
    return items;
});
const show = computed(
    () =>
        isAuthenticated.value &&
        !isPending.value &&
        !isError.value &&
        missedItems.value.length > 0,
);
const count = computed(() => missedItems.value.length);
function close() {
    open.value = false;
}
</script>
<template>
    <div v-if="show" ref="root" class="relative shrink-0">
        <button
            type="button"
            class="relative flex h-10 w-10 items-center justify-center rounded-md text-textSecondary transition-colors hover:bg-secondBg hover:text-textPrimary"
            :aria-expanded="open"
            aria-label="Tracking reminders"
            @click.stop="open = !open"
        >
            <Bell :size="32" :stroke-width="2" />
            <span
                class="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-(--color-cf-red) px-1 text-[10px] font-semibold leading-none text-white"
            >
                {{ count }}
            </span>
        </button>
        <div
            v-if="open"
            class="absolute right-0 top-full z-30 mt-1 min-w-56 rounded-md border border-thirdBg bg-secondBg p-0 shadow-lg"
            role="region"
            aria-label="Missed daily logs"
            @click.stop
        >
            <p
                class="m-0 border-b border-(--color-border) px-3 py-2 text-xs font-medium text-textSecondary"
            >
                Missed yesterday
            </p>
            <ul class="m-0 list-none p-0">
                <li
                    v-for="it in missedItems"
                    :key="it.key"
                    class="border-b border-(--color-border) last:border-b-0"
                >
                    <router-link
                        :to="it.to"
                        class="block px-3 py-2.5 text-sm text-textPrimary transition-colors hover:bg-thirdBg"
                        @click="close"
                    >
                        {{ it.label }}
                    </router-link>
                </li>
            </ul>
        </div>
    </div>
</template>
