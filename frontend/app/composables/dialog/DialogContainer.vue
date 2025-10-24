<script setup lang="ts">
import {
    dialogManager,
    type DialogOptions,
} from "~/composables/dialog/useDialog";
import { X } from "lucide-vue-next";

function isConfirmDialogType(
    d: typeof dialogManager.dialog.value,
): d is DialogOptions {
    return !!d && "message" in d;
}
</script>

<template>
    <div
        v-if="dialogManager.dialog.value"
        class="dialog-backdrop"
        :class="{
            'top-dialog': isConfirmDialogType(dialogManager.dialog.value),
        }"
    >
        <div class="dialog-container">
            <header class="dialog-header">
                <h2>{{ dialogManager.dialog.value.title }}</h2>
                <button
                    class="close-btn"
                    @click="dialogManager.resolve(false)"
                    v-if="!isConfirmDialogType(dialogManager.dialog.value)"
                >
                    <X />
                </button>
            </header>
            <template v-if="isConfirmDialogType(dialogManager.dialog.value)">
                <div class="confirm-template">
                    <h3>{{ dialogManager.dialog.value.message }}</h3>
                    <div class="confirm-buttons">
                        <button
                            @click="dialogManager.resolve(true)"
                            class="confirm-btn"
                        >
                            {{
                                dialogManager.dialog.value?.confirmText ||
                                "Confirm"
                            }}
                        </button>
                        <button
                            @click="dialogManager.resolve(false)"
                            class="cancel-btn"
                        >
                            {{
                                dialogManager.dialog.value?.cancelText ||
                                "Cancel"
                            }}
                        </button>
                    </div>
                </div>
            </template>
            <template v-else>
                <div class="template">
                    <component
                        :is="dialogManager.dialog.value.component"
                        v-bind="{
                            ...dialogManager.dialog.value.props,
                            onResolve: dialogManager.resolve,
                        }"
                    />
                </div>
            </template>
        </div>
    </div>
</template>

<style scoped>
.dialog-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding-top: 10vh;
    z-index: 998;
}

.top-dialog {
    align-items: flex-start;
    padding-top: 10vh;
}

.dialog-container {
    position: relative;
    display: flex;
    flex-direction: column;
    min-width: 30%;
    background: rgb(26, 26, 26);
    border: 1px solid #3d3d3d;
    border-radius: 10px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
    padding: 0;
}
.dialog-container > * {
    width: auto;
}

.dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 0.8rem;
    border-bottom: 1px solid #ccc;
}

.dialog-header h2 {
    margin: 0;
    font-size: 1.25rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
}

.template {
    flex: 1;
    padding: 0 1rem;
    overflow: visible;
    text-align: left;
}

.confirm-template {
    display: flex;
    flex-direction: column;
    /* gap: 0.5rem; */
    padding: 0rem 1rem 0.5rem 1rem;
}

.confirm-buttons {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    justify-content: center;
}

button {
    padding: 8px 16px;
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
