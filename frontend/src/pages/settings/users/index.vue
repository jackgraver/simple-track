<script setup lang="ts">
// Admin: family accounts and their sign-in activity. Accounts are created with a username only;
// the person sets their own password through a one-time link the admin texts them.
import { Ban, CircleCheck, KeyRound, MoreVertical, ShieldAlert, UserPlus } from "lucide-vue-next";
import { computed, ref } from "vue";
import type { AdminUser, SetupLink } from "~/api/account/api";
import { useAdminUsers, useAuthEvents, useCreateSetupLink, useCreateUser, useSetUserDisabled } from "~/api/account/queries";
import { getApiErrorMessage } from "~/api/httpError";
import { username as me } from "~/composables/auth/session";
import { toast } from "~/composables/toast/useToast";
import {
    Avatar,
    Badge,
    Button,
    Card,
    EmptyState,
    ErrorState,
    Field,
    IconButton,
    Input,
    List,
    ListRow,
    Menu,
    PageHeader,
    Section,
    SkeletonList,
    useDelayedFlag,
    type MenuItem,
} from "~/shared/ui";
import SetupLinkOverlay from "./SetupLinkOverlay.vue";
import { eventDetail, eventLabel, failureEvents } from "./eventLabels";

const users = useAdminUsers();
const events = useAuthEvents();
const showUsersSkeleton = useDelayedFlag(users.isPending);
const showEventsSkeleton = useDelayedFlag(events.isPending);
const eventRows = computed(() => events.data.value?.pages.flat() ?? []);

const createUser = useCreateUser();
const createLink = useCreateSetupLink();
const setDisabled = useSetUserDisabled();

const newUsername = ref("");
const linkOpen = ref(false);
const shownLink = ref<SetupLink | null>(null);
const shownFor = ref("");
const busyId = ref<number | null>(null);

function showLink(name: string, link: SetupLink) {
    shownFor.value = name;
    shownLink.value = link;
    linkOpen.value = true;
}

async function create() {
    const name = newUsername.value.trim().toLowerCase();
    if (!name) return;
    try {
        const { link } = await createUser.mutateAsync(name);
        newUsername.value = "";
        showLink(name, link);
    } catch (err) {
        toast.push(getApiErrorMessage(err, "Couldn't create the account"), "error");
    }
}

const dateTime = new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric", hour: "numeric", minute: "2-digit" });
const fmt = (iso: string) => dateTime.format(new Date(iso));

function status(u: AdminUser): string {
    if (u.disabled_at) return `Disabled ${fmt(u.disabled_at)}`;
    if (u.link_expires_at) {
        return `${u.link_purpose === "reset" ? "Reset" : "Setup"} link open until ${fmt(u.link_expires_at)}`;
    }
    if (!u.has_password) return "Not set up: send a new link";
    return u.last_login_at ? `Last sign-in ${fmt(u.last_login_at)}` : "Never signed in";
}

function menuFor(u: AdminUser): MenuItem[] {
    const items: MenuItem[] = [];
    if (!u.disabled_at) {
        items.push({ key: "link", label: u.has_password ? "Password reset link" : "New setup link", icon: KeyRound });
    }
    if (u.username !== me.value) {
        items.push(
            u.disabled_at
                ? { key: "enable", label: "Enable account", icon: CircleCheck }
                : { key: "disable", label: "Disable and sign out", icon: Ban, tone: "danger" },
        );
    }
    return items;
}

async function onMenu(u: AdminUser, key: string) {
    busyId.value = u.id;
    try {
        if (key === "link") {
            showLink(u.username, await createLink.mutateAsync(u.id));
        } else {
            const disabled = key === "disable";
            await setDisabled.mutateAsync({ id: u.id, disabled });
            toast.push(disabled ? `${u.username} is disabled and signed out` : `${u.username} can sign in again`, "success");
        }
    } catch (err) {
        toast.push(getApiErrorMessage(err, "That didn't work"), "error");
    } finally {
        busyId.value = null;
    }
}
</script>

<template>
    <div class="mx-auto flex w-full max-w-2xl flex-col gap-6 pb-8">
        <PageHeader title="Users" />

        <Section title="Add someone">
            <form class="flex flex-col gap-3 lg:flex-row lg:items-end" @submit.prevent="create">
                <Field class="flex-1" label="Username" hint="Lowercase letters, digits, '.', '_' or '-'. They'll choose their own password.">
                    <Input v-model="newUsername" autocomplete="off" required />
                </Field>
                <Button type="submit" variant="primary" :loading="createUser.isPending.value" :disabled="!newUsername.trim()">
                    <template #icon-left><UserPlus class="size-4" /></template>
                    Create and get link
                </Button>
            </form>
        </Section>

        <Section title="People">
            <ErrorState v-if="users.isError.value" :error="users.error.value" title="Couldn't load users" @retry="users.refetch()" />
            <template v-else-if="users.isPending.value">
                <SkeletonList v-if="showUsersSkeleton" :rows="3" />
            </template>
            <EmptyState v-else-if="!users.data.value?.length" title="No accounts yet" description="Add someone above." />
            <Card v-else padding="none">
                <List>
                    <ListRow v-for="u in users.data.value" :key="u.id" :pending="busyId === u.id">
                        <template #leading><Avatar :name="u.username" size="sm" /></template>
                        <span class="font-medium">{{ u.username }}</span>
                        <template #subtitle>{{ status(u) }}</template>
                        <template #end>
                            <Badge v-if="u.is_admin" tone="accent">Admin</Badge>
                            <Badge v-if="u.disabled_at" tone="danger">Disabled</Badge>
                            <Badge v-else-if="!u.has_password" tone="warn">Pending</Badge>
                            <Menu v-if="menuFor(u).length" :items="menuFor(u)" @select="(k) => onMenu(u, k)">
                                <IconButton :label="`Actions for ${u.username}`" :disabled="busyId === u.id">
                                    <MoreVertical />
                                </IconButton>
                            </Menu>
                        </template>
                    </ListRow>
                </List>
            </Card>
        </Section>

        <Section title="Sign-in activity">
            <ErrorState v-if="events.isError.value" :error="events.error.value" title="Couldn't load activity" @retry="events.refetch()" />
            <template v-else-if="events.isPending.value">
                <SkeletonList v-if="showEventsSkeleton" :rows="5" />
            </template>
            <EmptyState v-else-if="!eventRows.length" title="No activity yet" description="Sign-ins, links and account changes show up here." />
            <div v-else class="flex flex-col gap-3">
                <Card padding="none">
                    <List>
                        <ListRow v-for="e in eventRows" :key="e.id">
                            <template #leading>
                                <ShieldAlert v-if="failureEvents.has(e.event)" class="size-4 text-danger" aria-label="Failure" />
                                <CircleCheck v-else class="size-4" aria-hidden="true" />
                            </template>
                            {{ eventLabel(e) }}<template v-if="e.username"> · {{ e.username }}</template>
                            <template #subtitle>
                                {{ fmt(e.created_at) }}
                                <template v-if="eventDetail(e)"> · {{ eventDetail(e) }}</template>
                                <template v-if="e.actor_username && e.actor_username !== e.username"> · by {{ e.actor_username }}</template>
                                <template v-if="e.client_ip"> · {{ e.client_ip }}</template>
                            </template>
                        </ListRow>
                    </List>
                </Card>
                <Button
                    v-if="events.hasNextPage.value"
                    variant="secondary"
                    :loading="events.isFetchingNextPage.value"
                    @click="events.fetchNextPage()"
                >
                    Show older
                </Button>
            </div>
        </Section>

        <SetupLinkOverlay v-model:open="linkOpen" :username="shownFor" :link="shownLink" />
    </div>
</template>
