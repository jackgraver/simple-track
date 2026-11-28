<script setup lang="ts">
// Shows a freshly generated setup/reset link once, to copy or share by text. The token isn't stored
// anywhere readable, so closing this means generating a new link to see one again.
import { Copy, Share2 } from "lucide-vue-next";
import { computed } from "vue";
import { setupLinkURL, type SetupLink } from "~/api/account/api";
import { toast } from "~/composables/toast/useToast";
import { Button, Field, Input, Overlay } from "~/shared/ui";

const open = defineModel<boolean>("open", { required: true });
const props = defineProps<{ username: string; link: SetupLink | null }>();

const url = computed(() => (props.link ? setupLinkURL(props.link) : ""));
const canShare = typeof navigator !== "undefined" && typeof navigator.share === "function";
const expires = computed(() =>
    props.link
        ? new Intl.DateTimeFormat(undefined, { weekday: "short", hour: "numeric", minute: "2-digit" }).format(
              new Date(props.link.expires_at),
          )
        : "",
);

async function copy() {
    try {
        await navigator.clipboard.writeText(url.value);
        toast.push("Link copied", "success");
    } catch {
        toast.push("Couldn't copy: select the link and copy it by hand", "error");
    }
}

async function share() {
    try {
        await navigator.share({ title: "simple-track", text: `Your simple-track link (${props.username}):`, url: url.value });
    } catch {
        // Cancelled from the share sheet: nothing to do.
    }
}
</script>

<template>
    <Overlay
        v-model:open="open"
        :title="link?.purpose === 'reset' ? `Password reset link for ${username}` : `Setup link for ${username}`"
        description="Text it to them. It works once and expires in 48 hours; making a new one cancels this one."
    >
        <div class="flex flex-col gap-4 p-4">
            <Field label="Link" :hint="`Expires ${expires}`">
                <Input :model-value="url" readonly aria-label="Setup link" @focus="($event.target as HTMLInputElement).select()" />
            </Field>
            <div class="flex flex-col gap-2 lg:flex-row-reverse">
                <Button v-if="canShare" variant="primary" block @click="share">
                    <template #icon-left><Share2 class="size-4" /></template>
                    Share
                </Button>
                <Button :variant="canShare ? 'secondary' : 'primary'" block @click="copy">
                    <template #icon-left><Copy class="size-4" /></template>
                    Copy link
                </Button>
            </div>
        </div>
    </Overlay>
</template>
