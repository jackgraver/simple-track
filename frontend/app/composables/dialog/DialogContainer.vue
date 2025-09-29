<script setup lang="ts">
import { dialogManager } from "~/composables/dialog/useDialog";
console.log(dialogManager.dialog);
</script>

<template>
    <div v-if="dialogManager.dialog.value" class="dialog-container">
        <div>
            <h2>
                {{ dialogManager.dialog.value?.title }}
            </h2>
            <p>{{ dialogManager.dialog.value?.message }}</p>

            <div>
                <button @click="dialogManager.resolve(false)">
                    {{ dialogManager.dialog.value?.cancelText || "Cancel" }}
                </button>
                <button @click="dialogManager.resolve(true)">
                    {{ dialogManager.dialog.value?.confirmText || "OK" }}
                </button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.dialog-container {
    position: absolute;
    background: rgb(71, 71, 71);
    z-index: 999;
    top: 10%;
    left: 30%;
    width: 50%;
    height: 50%;
    border-radius: 10px;
}
</style>
