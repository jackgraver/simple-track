<script setup lang="ts">
import { ref, watch } from "vue";
import { Button, Field, Input, Overlay } from "~/shared/ui";

// A one-field naming dialog (new program, rename program, new routine). Replaces window.prompt.
const open = defineModel<boolean>("open", { required: true });
const props = withDefaults(
    defineProps<{ title: string; label: string; initial?: string; confirmText?: string; loading?: boolean; placeholder?: string }>(),
    { initial: "", confirmText: "Save", loading: false, placeholder: "" },
);
const emit = defineEmits<{ submit: [name: string] }>();

const name = ref(props.initial);
const touched = ref(false);
watch(open, (isOpen) => {
    if (!isOpen) return;
    name.value = props.initial;
    touched.value = false;
});

const submit = () => {
    touched.value = true;
    const trimmed = name.value.trim();
    if (!trimmed || props.loading) return;
    emit("submit", trimmed);
};
</script>

<template>
    <Overlay v-model:open="open" :title="title" size="sm">
        <form id="name-dialog-form" class="flex flex-col gap-4" @submit.prevent="submit">
            <Field :label="label" :error="touched && !name.trim() ? 'Enter a name' : undefined" required>
                <Input v-model="name" :placeholder="placeholder" autocomplete="off" />
            </Field>
        </form>
        <template #footer>
            <Button variant="ghost" @click="open = false">Cancel</Button>
            <Button variant="primary" type="submit" form="name-dialog-form" :loading="loading">{{ confirmText }}</Button>
        </template>
    </Overlay>
</template>
