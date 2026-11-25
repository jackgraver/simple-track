import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import { ChevronsUpDown, Copy, LogOut, Monitor, Moon, MoreVertical, Pencil, Sun, Trash2, UserRound } from "lucide-vue-next";
import Avatar from "./Avatar.vue";
import DropdownMenu from "./DropdownMenu.vue";
import IconButton from "./IconButton.vue";
import MenuAction from "./MenuAction.vue";
import MenuLabel from "./MenuLabel.vue";
import MenuRadioGroup from "./MenuRadioGroup.vue";
import MenuSeparator from "./MenuSeparator.vue";

const meta = {
    title: "UI/DropdownMenu",
    component: DropdownMenu,
    tags: ["autodocs"],
} satisfies Meta<typeof DropdownMenu>;
export default meta;
type Story = StoryObj<typeof meta>;

const parts = { Avatar, DropdownMenu, IconButton, MenuAction, MenuLabel, MenuRadioGroup, MenuSeparator };
const icons = { ChevronsUpDown, Copy, LogOut, MoreVertical, Pencil, Trash2, UserRound };

/** Every part: header and section labels, a link, a radio row, loading, disabled and danger items. */
export const AccountRow: Story = {
    render: () => ({
        components: parts,
        setup: () => ({
            ...icons,
            theme: ref("dark"),
            isOpen: ref(true),
            themes: [
                { label: "Light", value: "light", icon: Sun },
                { label: "Dark", value: "dark", icon: Moon },
                { label: "System", value: "system", icon: Monitor },
            ],
        }),
        template: `
            <div class="flex h-96 w-56 items-end p-3">
                <DropdownMenu side="top" align="start" width="md" shape="row" trigger-label="Account menu for dev" v-model:open="isOpen">
                    <template #trigger-content>
                        <Avatar name="dev" size="sm" />
                        <span class="min-w-0 flex-1 truncate text-sm font-medium">dev</span>
                        <ChevronsUpDown aria-hidden="true" class="size-4 shrink-0 text-fg-muted" />
                    </template>
                    <MenuLabel variant="header">
                        <Avatar name="dev" size="md" />
                        <span class="flex flex-col"><span class="font-medium">dev</span><span class="text-xs font-normal text-fg-muted">Signed in</span></span>
                    </MenuLabel>
                    <MenuSeparator />
                    <MenuAction :icon="UserRound" to="/settings/profile">Body profile</MenuAction>
                    <MenuAction :icon="Copy" disabled>Disabled item</MenuAction>
                    <MenuAction :icon="Pencil">With trailing<template #trailing>⌘E</template></MenuAction>
                    <MenuSeparator />
                    <MenuLabel>Theme</MenuLabel>
                    <MenuRadioGroup v-model="theme" label="Theme" :options="themes" />
                    <MenuSeparator />
                    <MenuAction :icon="LogOut" loading>Signing out (loading)</MenuAction>
                    <MenuAction :icon="Trash2" tone="danger">Delete account</MenuAction>
                </DropdownMenu>
            </div>`,
    }),
};

/** `shape="circle"` with `active`: the mobile header avatar on the page the menu links to. */
export const CircleTrigger: Story = {
    render: () => ({
        components: parts,
        setup: () => icons,
        template: `
            <div class="flex justify-end p-4">
                <DropdownMenu shape="circle" active trigger-label="Account menu for dev">
                    <template #trigger-content><Avatar name="dev" size="sm" /></template>
                    <MenuAction :icon="UserRound">Body profile</MenuAction>
                </DropdownMenu>
            </div>`,
    }),
};

export const AsChildTrigger: Story = {
    render: () => ({
        components: parts,
        setup: () => icons,
        template: `
            <div class="flex justify-end p-4">
                <DropdownMenu>
                    <template #trigger><IconButton label="Actions"><MoreVertical /></IconButton></template>
                    <MenuAction :icon="Pencil">Edit</MenuAction>
                    <MenuAction :icon="Trash2" tone="danger">Delete</MenuAction>
                </DropdownMenu>
            </div>`,
    }),
};
