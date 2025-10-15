<script setup lang="ts">
import ToastContainer from "./composables/toast/ToastContainer.vue";
import DialogContainer from "./composables/dialog/DialogContainer.vue";
</script>

<template>
    <div>
        <!-- <Header /> -->
        <NuxtPage />
        <ToastContainer />
        <DialogContainer />
    </div>
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

button:disabled {
    background: #ccc;
}

button:disabled:hover {
    cursor: not-allowed;
}

button:hover && button:not(:disabled) {
    background: #4baeff;
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
