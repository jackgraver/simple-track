<script setup lang="ts">
import { computed, shallowRef, watch } from "vue";
import { dialogManager } from "~/composables/dialog/useDialog";
import { Button, Overlay } from "~/shared/ui";

// Renders the one dialog `dialogManager` holds. The API is unchanged; this is only the presentation.
// `shown` keeps the last dialog mounted while Overlay plays its exit animation.
const shown = shallowRef(dialogManager.dialog.value);
watch(
    () => dialogManager.dialog.value,
    (d) => {
        if (d) shown.value = d;
    },
);

const open = computed({
    get: () => !!dialogManager.dialog.value,
    set: (v) => {
        if (!v) dialogManager.cancel();
    },
});
</script>

<template>
    <Overlay
        v-if="shown"
        v-model:open="open"
        :title="shown.options.title || ''"
        :description="shown.type === 'confirm' ? shown.options.message : undefined"
        :size="shown.type === 'custom' && shown.options.wide ? 'lg' : 'md'"
    >
        <component
            :is="shown.options.component"
            v-if="shown.type === 'custom'"
            v-bind="{
                ...(shown.options.componentProps ?? {}),
                onResolve: dialogManager.resolve,
                onCancel: dialogManager.cancel,
            }"
        />
        <template v-if="shown.type === 'confirm'" #footer>
            <Button variant="secondary" @click="dialogManager.resolve(false)">
                {{ shown.options.cancelText || "Cancel" }}
            </Button>
            <Button :variant="shown.options.tone === 'danger' ? 'danger-solid' : 'primary'" @click="dialogManager.resolve(true)">
                {{ shown.options.confirmText || "Confirm" }}
            </Button>
        </template>
    </Overlay>
</template>
