<script setup lang="ts">
import { ref } from "vue";

const data = ref("");

function addList() {
    const splitList = data.value.split("\n");
    
}
</script>

<template>
    <h1>Enter Text List</h1>
    <textarea 
        v-model="data"
        type="text"
        cols="40"
        rows="4"
        placeholder="List.."
    />
        <button
            @click="addList"
            >Add List</button>
</template>
