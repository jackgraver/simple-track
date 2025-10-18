<script setup lang="ts">
import ToastContainer from "./composables/toast/ToastContainer.vue";
import DialogContainer from "./composables/dialog/DialogContainer.vue";
</script>

<template>
    <!-- <Header /> -->
    <NuxtPage />
    <ToastContainer />
    <DialogContainer />
</template>

<style>
html {
    color: white;
    background-color: black;
}

button {
    background: #3a88c7;
    padding: 6px 12px;
    margin: 4px;
    border-radius: 4px;
    color: white;
    border: none;
}

.delete-button {
    background: #ff2b2b;
}

.confirm-button {
    background: #62ff33;
}

button:disabled {
    background: #ccc;
}

button:disabled:hover {
    cursor: not-allowed;
}

button:hover:not(:disabled) {
    background: #58b4ff;
}

input,
select {
    background: rgb(71, 71, 71);
    color: white;
    border: none;
    padding: 6px 12px;
    margin: 4px;
    border-radius: 4px;
}
</style>
