<script setup lang="ts">
import { toast } from "~/composables/toast/useToast";

const dump = async () => {
    const { response, error } = await useAPIPost<{ dump: string }>(
        "db/dump",
        "POST",
        {},
    );
    if (error) {
        toast.push("Dump Failed!" + error.message, "error");
    } else if (response) {
        toast.push("Dump Successfully!", "success");
    }
};
const restore = async () => {
    const { response, error } = await useAPIPost<{ restore: string }>(
        "db/restore",
        "POST",
        {},
    );
    if (error) {
        toast.push("Restore Failed!" + error.message, "error");
    } else if (response) {
        toast.push("Restore Successfully!", "success");
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
    top: 15%;
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    padding: 2rem;
    max-width: 75%;
    width: 100%;
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 0.5rem;
}
</style>
