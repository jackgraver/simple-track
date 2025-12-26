<script setup lang="ts">
import { onMounted } from 'vue';
import ToastContainer from "./composables/toast/ToastContainer.vue";
import DialogContainer from "./composables/dialog/DialogContainer.vue";
import SideBar from "./shared/SideBar.vue";

onMounted(() => {
    document.documentElement.classList.add('dark-mode');
});
</script>

<template>
    <div class="page dark-mode">
        <SideBar />
        <RouterView />
    </div>

    <ToastContainer />
    <DialogContainer />
</template>

<style>
html {
    color: rgb(218, 218, 218);
    background-color: black;
    font-family:
        system-ui,
        -apple-system,
        BlinkMacSystemFont,
        "Segoe UI",
        Roboto,
        "Helvetica Neue",
        Arial,
        sans-serif;
}

@media (max-width: 767px) {
    .page {
        display: flex;
        flex-direction: column;
    }
}

@media (min-width: 768px) {
    .page {
        height: 100vh;
        display: flex;
        flex-direction: row;
    }
}

a,
.router-link {
    all: unset;
    cursor: pointer;
}

a,
.router-link {
    text-decoration: none;
    color: inherit;
    font: inherit;
}

button {
    background: #2c2c2c;
    padding: 6px 12px;
    margin: 4px 0px;
    border-radius: 4px;
    border: 2px solid #a7a7a7;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    color: white;
    border: none;
    cursor: pointer;
}

button:disabled {
    background: #ccc;
}

button:disabled:hover {
    cursor: not-allowed;
}

button:hover:not(:disabled) {
    background: #525252;
}

button svg {
    vertical-align: middle;
    pointer-events: none;
}

.delete-button {
    background: #ff2b2b;
}

.delete-button:hover:not(:disabled) {
    background: #c91e1e;
}

.confirm-button {
    background: #55db2c;
}

.confirm-button:hover:not(:disabled) {
    background: #38a716;
}

input,
select {
    background: rgb(71, 71, 71);
    color: white;
    border: none;
    padding: 6px 12px;
    margin: 4px 0 4px 0;
    border-radius: 4px;
}

input:focus {
    outline: none;
}
</style>

