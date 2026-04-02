<script setup lang="ts">
import { useWorkoutStore } from "./store/useWorkoutStore";
import MobilityLoggingView from "./components/MobilityLoggingView.vue";
import { useRoute, useRouter } from "vue-router";
import { computed } from "vue";
import { toast } from "~/composables/toast/useToast";

const router = useRouter();
const route = useRoute();
const offset = computed(() => {
    const raw = route.query.offset;
    const value = typeof raw === "string" ? Number.parseInt(raw, 10) : 0;
    return Number.isNaN(value) ? 0 : value;
});

const { loggedPreMobility, pending, error, savePreMobility } = useWorkoutStore(offset);

const goBack = () => {
    router.push({
        name: "logging",
        query: offset.value === 0 ? {} : { offset: String(offset.value) },
    });
};

const handleSave = async (checked: string[]) => {
    try {
        await savePreMobility(checked);
    } catch (err: any) {
        toast.push(err.message || "Failed to save", "error");
    }
};
</script>

<template>
    <div v-if="pending" class="flex w-full flex-col gap-4 self-stretch">
        <div>Loading...</div>
    </div>
    <div v-else-if="error" class="flex w-full flex-col gap-4 self-stretch">
        <div>Error: {{ error.message }}</div>
    </div>
    <div v-else-if="!loggedPreMobility" class="flex w-full flex-col gap-4 self-stretch">
        <div>No pre-workout mobility for this day.</div>
        <button
            class="rounded border border-zinc-600 px-4 py-2 hover:bg-zinc-800"
            type="button"
            @click="goBack"
        >
            Back
        </button>
    </div>
    <div v-else class="container flex w-full flex-col gap-4 self-stretch">
        <MobilityLoggingView :logged-mobility="loggedPreMobility" @back="goBack" @save="handleSave" />
    </div>
</template>

<style scoped>
.container {
    width: 100%;
    max-width: 100%;
}
@media (max-width: 767px) {
    .container {
        padding: 0.5rem 0;
    }
}
</style>
