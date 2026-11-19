<script setup lang="ts">
import { toast } from "~/composables/toast/useToast";
import { Toaster } from "~/shared/ui";

defineProps<{ tabBar?: boolean; top?: boolean }>();

function runAction(id: number) {
    toast.toasts.value.find((t) => t.id === id)?.action?.onClick();
    toast.dismiss(id);
}
</script>

<template>
    <Toaster
        :toasts="toast.toasts.value"
        :tab-bar="tabBar"
        :top="top"
        @dismiss="toast.dismiss"
        @action="runAction"
    />
</template>
