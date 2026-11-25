<script setup lang="ts">
// The signed-in account: avatar trigger + a menu with the body profile link, theme and sign out.
// `row` is the side nav footer (avatar + name); `compact` is the avatar alone (mobile header).
import { ChevronsUpDown, LogOut, Monitor, Moon, Sun, UserRound } from "lucide-vue-next";
import { computed } from "vue";
import { useAuth } from "~/composables/auth/useAuth";
import { toast } from "~/composables/toast/useToast";
import { useTheme, type ThemeMode } from "~/composables/useTheme";
import { Avatar, DropdownMenu, MenuAction, MenuLabel, MenuRadioGroup, MenuSeparator } from "~/shared/ui";

defineProps<{ variant: "row" | "compact" }>();

const { username, logout } = useAuth();
const name = computed(() => username.value || "Account");
const { mode, setThemeMode } = useTheme();

const themeOptions: { label: string; value: ThemeMode; icon: typeof Sun }[] = [
    { label: "Light", value: "light", icon: Sun },
    { label: "Dark", value: "dark", icon: Moon },
    { label: "System", value: "system", icon: Monitor },
];

// logout() never throws: a failed request still clears the session locally and goes to sign-in,
// so the user is signed out either way. The Toaster is app-level, so the toast survives the redirect.
async function signOut() {
    await logout();
    toast.push("Signed out", "success");
}

</script>

<template>
    <DropdownMenu
        :side="variant === 'row' ? 'top' : 'bottom'"
        :align="variant === 'row' ? 'start' : 'end'"
        width="md"
        :trigger-label="`Account menu for ${name}`"
        :trigger-class="
            variant === 'row'
                ? 'w-full gap-2.5 px-2 py-1.5'
                : 'size-touch justify-center rounded-full'
        "
    >
        <template #trigger-content>
            <Avatar :name="name" size="sm" />
            <template v-if="variant === 'row'">
                <span class="min-w-0 flex-1 truncate text-sm font-medium">{{ name }}</span>
                <ChevronsUpDown aria-hidden="true" class="size-4 shrink-0 text-fg-muted" />
            </template>
        </template>

        <MenuLabel variant="header">
            <Avatar :name="name" size="md" />
            <span class="flex min-w-0 flex-col">
                <span class="truncate font-medium">{{ name }}</span>
                <span class="text-xs font-normal text-fg-muted">Signed in</span>
            </span>
        </MenuLabel>
        <MenuSeparator />
        <MenuAction :icon="UserRound" :to="{ name: 'settings-profile' }">Body profile</MenuAction>
        <MenuSeparator />
        <MenuLabel>Theme</MenuLabel>
        <MenuRadioGroup label="Theme" :options="themeOptions" :model-value="mode" @update:model-value="setThemeMode" />
        <MenuSeparator />
        <MenuAction :icon="LogOut" @select="void signOut()">Sign out</MenuAction>
    </DropdownMenu>
</template>
