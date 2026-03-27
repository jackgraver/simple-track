<script setup lang="ts">
import { toast } from "~/composables/toast/useToast";
import { apiClient } from "~/utils/axios";

const dump = async () => {
    try {
        await apiClient.post("db/dump", {});
        toast.push("Dump Successfully!", "success");
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Dump Failed!" + message, "error");
    }
};
const restore = async () => {
    try {
        await apiClient.post("db/restore", {});
        toast.push("Restore Successfully!", "success");
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Restore Failed!" + message, "error");
    }
};
</script>

<template>
    <div class="settings-container">
        <button @click="dump">Dump DB</button>
        <button @click="restore">Restore DB</button>
    </div>
</template>

<style scoped>
.settings-container {
    position: fixed;
    bottom: 0;
    right: 0;
    padding: 1rem;
    display: flex;
    gap: 0.5rem;
    z-index: 1000;
}
</style>
