<script setup lang="ts">
import { ChevronsUpDown, LogOut, Monitor, Moon, Sun, UserRound } from "lucide-vue-next";
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { useAuth } from "~/composables/auth/useAuth";
import { toast } from "~/composables/toast/useToast";
import { useTheme, type ThemeMode } from "~/composables/useTheme";
import { Avatar, DropdownMenu, MenuAction, MenuLabel, MenuRadioGroup, MenuSeparator } from "~/shared/ui";

const route = useRoute();
const { username, logout } = useAuth();
const { mode, setThemeMode } = useTheme();
const open = ref(false);
const signingOut = ref(false);

const onProfilePage = computed(() => route.name === "settings-profile");
const themeOptions: { label: string; value: ThemeMode; icon: typeof Sun }[] = [
    { label: "Light", value: "light", icon: Sun },
    { label: "Dark", value: "dark", icon: Moon },
    { label: "System", value: "system", icon: Monitor },
];

async function signOut(e: Event) {
    e.preventDefault();
    signingOut.value = true;
    const confirmed = await logout();
    signingOut.value = false;
    open.value = false;
    report(confirmed);
}

function report(confirmed: boolean) {
    if (confirmed) {
        toast.push("Signed out", "success");
        return;
    }
    // The httpOnly cookie may still be valid, so this isn't a real sign-out yet.
    toast.push("Couldn't reach the server to sign out", "error", {
        action: { label: "Retry", onClick: async () => report(await logout()) },
    });
}
</script>

<template>
    <DropdownMenu
        v-model:open="open"
        side="top"
        align="start"
        shape="row"
        :active="onProfilePage"
        width="md"
        :trigger-label="username ? `Account menu for ${username}` : 'Account menu'"
    >
        <template #trigger-content>
            <Avatar :name="username ?? ''" size="sm" />
            <span class="min-w-0 flex-1 truncate text-sm font-medium">{{ username }}</span>
            <ChevronsUpDown aria-hidden="true" class="size-4 shrink-0 text-fg-muted" />
        </template>

        <MenuLabel v-if="username" variant="header">
            <Avatar :name="username" size="md" />
            <span class="flex min-w-0 flex-col">
                <span class="truncate font-medium">{{ username }}</span>
                <span class="text-xs font-normal text-fg-muted">Signed in</span>
            </span>
        </MenuLabel>
        <MenuSeparator v-if="username" />
        <MenuAction :icon="UserRound" :to="{ name: 'settings-profile' }">Body profile</MenuAction>
        <MenuSeparator />
        <MenuLabel>Theme</MenuLabel>
        <MenuRadioGroup label="Theme" :options="themeOptions" :model-value="mode" @update:model-value="setThemeMode" />
        <MenuSeparator />
        <MenuAction :icon="LogOut" :loading="signingOut" @select="signOut">Sign out</MenuAction>
    </DropdownMenu>
</template>
