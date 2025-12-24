<script setup lang="ts">
import {
    dialogManager,
    type DialogOptions,
} from "~/composables/dialog/useDialog";
import { X } from "lucide-vue-next";

function isConfirmDialog(
    d: typeof dialogManager.dialog.value,
): d is { type: "confirm"; options: DialogOptions } {
    return d?.type === "confirm";
}
</script>

<template>
    <div
        v-if="dialogManager.dialog.value"
        class="dialog-backdrop"
        :class="{
            'top-dialog': isConfirmDialog(dialogManager.dialog.value),
        }"
    >
        <div class="dialog-container">
            <header class="dialog-header">
                <h2>
                    {{
                        isConfirmDialog(dialogManager.dialog.value)
                            ? dialogManager.dialog.value.options.title
                            : dialogManager.dialog.value.options.title || ""
                    }}
                </h2>
                <button
                    class="close-btn"
                    @click="dialogManager.cancel()"
                    v-if="!isConfirmDialog(dialogManager.dialog.value)"
                >
                    <X />
                </button>
            </header>
            <template v-if="isConfirmDialog(dialogManager.dialog.value)">
                <div class="confirm-template">
                    <h3>{{ dialogManager.dialog.value.options.message }}</h3>
                    <div class="confirm-buttons">
                        <button
                            @click="dialogManager.resolve(true)"
                            class="confirm-btn"
                        >
                            {{
                                dialogManager.dialog.value.options
                                    .confirmText || "Confirm"
                            }}
                        </button>
                        <button
                            @click="dialogManager.resolve(false)"
                            class="cancel-btn"
                        >
                            {{
                                dialogManager.dialog.value.options
                                    .cancelText || "Cancel"
                            }}
                        </button>
                    </div>
                </div>
            </template>
            <template v-else>
                <div class="template">
                    <component
                        :is="dialogManager.dialog.value.options.component"
                        v-bind="{
                            ...dialogManager.dialog.value.options.props,
                            onResolve: dialogManager.resolve,
                            onCancel: dialogManager.cancel,
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
    max-width: 90vw;
    background: rgb(26, 26, 26);
    border: 1px solid #3d3d3d;
    border-radius: 10px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
    padding: 0;
    max-height: 85vh;
    overflow: hidden;
}

.dialog-container > * {
    width: auto;
}

.dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.25rem;
    border-bottom: 1px solid #3d3d3d;
    flex-shrink: 0;
}

.dialog-header h2 {
    margin: 0;
    font-size: 1.25rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
}

.close-btn {
    background: transparent;
    border: none;
    color: #ccc;
    cursor: pointer;
    padding: 0.25rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: background 0.2s;
}

.close-btn:hover {
    background: rgba(255, 255, 255, 0.1);
}

.template {
    flex: 1;
    padding: 1.25rem;
    overflow-y: auto;
    text-align: left;
}

.confirm-template {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    padding: 1.25rem;
}

.confirm-template h3 {
    margin: 0;
    font-size: 1rem;
    font-weight: normal;
    color: #ccc;
}

.confirm-buttons {
    display: flex;
    flex-direction: row;
    gap: 0.75rem;
    justify-content: flex-end;
}

button {
    padding: 0.625rem 1.25rem;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 500;
    transition: all 0.2s;
}

.confirm-btn {
    background: rgb(34, 139, 34);
    color: white;
}

.confirm-btn:hover {
    background: rgb(40, 160, 40);
}

.cancel-btn {
    background: rgb(60, 60, 60);
    color: white;
}

.cancel-btn:hover {
    background: rgb(80, 80, 80);
}
</style>
