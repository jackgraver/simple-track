<script setup lang="ts">
import { Button, EmptyState } from "~/shared/ui";
</script>

<template>
    <div class="py-10">
        <EmptyState title="Page not found" description="That page doesn't exist.">
            <template #action>
                <Button variant="primary" :to="{ name: 'gym' }">Back to gym</Button>
            </template>
        </EmptyState>
    </div>
</template>
