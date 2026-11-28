import { apiClient, apiGET, apiPOST } from "~/api/client";
import type { AuthResponse } from "~/composables/auth/useAuth";

/** A row of the admin's user list (GET /admin/users). */
export interface AdminUser {
    id: number;
    username: string;
    is_admin: boolean;
    has_password: boolean;
    created_at: string;
    disabled_at: string | null;
    /** Expiry of the user's live setup/reset link, if one is outstanding. */
    link_expires_at: string | null;
    link_purpose: SetupPurpose | null;
    last_login_at: string | null;
}

export type SetupPurpose = "setup" | "reset";

/** A freshly generated one-time link. `path` is `/setup/<token>`; the token is never shown again. */
export interface SetupLink {
    path: string;
    purpose: SetupPurpose;
    expires_at: string;
}

export interface AuthEvent {
    id: number;
    created_at: string;
    event: string;
    user_id: number | null;
    username: string | null;
    actor_id: number | null;
    actor_username: string | null;
    client_ip: string;
    detail: string;
}

export interface SetupInfo {
    username: string;
    purpose: SetupPurpose;
    expires_at: string;
}

export async function fetchUsers(): Promise<AdminUser[]> {
    return (await apiGET<{ users: AdminUser[] }>("/admin/users")).users;
}

export async function createUser(username: string): Promise<{ link: SetupLink }> {
    return apiPOST<{ link: SetupLink }>("/admin/users", { username });
}

export async function createSetupLink(userId: number): Promise<SetupLink> {
    return (await apiPOST<{ link: SetupLink }>(`/admin/users/${userId}/setup-link`)).link;
}

export async function setUserDisabled(userId: number, disabled: boolean): Promise<void> {
    await apiPOST(`/admin/users/${userId}/${disabled ? "disable" : "enable"}`);
}

export async function fetchAuthEvents(before?: number): Promise<AuthEvent[]> {
    const params = before ? { before, limit: 50 } : { limit: 50 };
    return (await apiGET<{ events: AuthEvent[] }>("/admin/auth-events", { params })).events;
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await apiPOST("/auth/password", { current_password: currentPassword, new_password: newPassword });
}

/** Tokens go in the body, never in an API URL, so they stay out of request logs. */
export async function inspectSetupLink(token: string): Promise<SetupInfo> {
    return apiPOST<SetupInfo>("/auth/setup/inspect", { token });
}

export async function completeSetup(token: string, password: string): Promise<AuthResponse> {
    return (await apiClient.post<AuthResponse>("/auth/setup", { token, password })).data;
}

/** The full link to text: the server only knows the path, the client knows its own origin. */
export function setupLinkURL(link: SetupLink): string {
    return new URL(link.path, window.location.origin).toString();
}

/** Same rules as the server (services.ValidatePassword): 10+ characters, at most 72 bytes. */
export function passwordProblem(password: string): string | null {
    if ([...password].length < 10) return "Use at least 10 characters.";
    if (new TextEncoder().encode(password).length > 72) return "Too long: at most 72 bytes (fewer with accents or emoji).";
    return null;
}
