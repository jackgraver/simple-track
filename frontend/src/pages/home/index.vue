<script setup lang="ts">
import { PageHeader, EmptyState } from "~/shared/ui";
</script>

<template>
    <div class="mx-auto flex w-full max-w-3xl flex-col gap-6">
        <PageHeader title="Home" />
        <EmptyState title="Nothing here yet" description="The home page is still a work in progress." />
    </div>
</template>
