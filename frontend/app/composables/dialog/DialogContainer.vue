<script setup lang="ts">
import {
    dialogManager,
    type DialogOptions,
} from "~/composables/dialog/useDialog";

function isConfirmDialogType(
    d: typeof dialogManager.dialog.value,
): d is DialogOptions {
    return !!d && "message" in d;
}
</script>

<template>
    <div v-if="dialogManager.dialog.value" class="dialog-backdrop">
        <div class="dialog-container">
            <header class="dialog-header">
                <h2>{{ dialogManager.dialog.value.title }}</h2>
                <button class="close-btn" @click="dialogManager.resolve(false)">
                    ×
                </button>
            </header>
            <template v-if="isConfirmDialogType(dialogManager.dialog.value)">
                <div class="template">
                    <p>{{ dialogManager.dialog.value.message }}</p>
                    <button
                        @click="dialogManager.resolve(true)"
                        class="confirm-btn"
                    >
                        {{
                            dialogManager.dialog.value?.confirmText || "Confirm"
                        }}
                    </button>
                    <button
                        @click="dialogManager.resolve(false)"
                        class="cancel-btn"
                    >
                        {{ dialogManager.dialog.value?.cancelText || "Cancel" }}
                    </button>
                </div>
            </template>
            <template v-else>
                <div class="template">
                    <component :is="dialogManager.dialog.value?.content" />
                </div>
            </template>
        </div>
    </div>
</template>

<style scoped>
.dialog-container {
    position: absolute;
    top: 25%; /* ~1/4 from the top */
    left: 50%; /* center horizontally */
    transform: translateX(-50%); /* shift left by half its width */

    background: rgb(71, 71, 71);
    z-index: 999;
    max-height: 80%; /* optional: prevent going off-screen */
    border-radius: 10px;

    display: flex;
    flex-direction: column;
}

.dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid #ccc;

    /* ensure children don't wrap */
    flex-wrap: nowrap;
}

.dialog-header h2 {
    margin: 0;
    font-size: 1.5rem;

    /* prevent text from wrapping */
    white-space: nowrap;
    overflow: hidden; /* optional: hide overflow */
    text-overflow: ellipsis; /* optional: show "..." if too long */
    max-width: calc(100% - 3rem); /* leave space for close button */
}

.template {
    padding: 1rem; /* your current padding works fine */
    overflow-y: auto; /* optional if content can scroll */
    text-align: center;
}

.close-btn {
    background: transparent;
    border: none;
    font-size: 2rem;
    cursor: pointer;
    margin-left: 1rem; /* optional extra space */
}

.close-btn:hover {
    background-color: #ccc;
}

.confirm-btn {
    background: green;
}

.confirm-btn:hover {
    background: rgb(0, 155, 0);
}

.cancel-btn {
    background: rgb(211, 0, 0);
}

.cancel-btn:hover {
    background: rgb(255, 0, 0);
}
</style>
