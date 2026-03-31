<script setup lang="ts">
import { useWorkoutStore } from "./store/useWorkoutStore";
import CardioLoggingView from "./components/CardioLoggingView.vue";
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

const { plannedCardio, loggedCardio, pending, error, saveCardio } =
    useWorkoutStore(offset);

const goBack = () => {
    router.push({
        name: "logging",
        query: offset.value === 0 ? {} : { offset: String(offset.value) },
    });
};

const handleSave = async (minutes: number) => {
    try {
        await saveCardio(minutes);
        toast.push("Cardio saved", "success");
        goBack();
    } catch (err: any) {
        toast.push(err.message || "Failed to save cardio", "error");
    }
};
</script>

<template>
    <div v-if="pending" class="container">
        <div>Loading...</div>
    </div>
    <div v-else-if="error" class="container">
        <div>Error: {{ error.message }}</div>
    </div>
    <div v-else class="container">
        <CardioLoggingView
            :planned-cardio="plannedCardio"
            :logged-cardio="loggedCardio"
            @save="handleSave"
            @back="goBack"
        />
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
    align-self: stretch;
}
@media (max-width: 767px) {
    .container {
        padding: 0.5rem 0;
    }
}
</style>
