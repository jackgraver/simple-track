import type { AuthEvent } from "~/api/account/api";

const labels: Record<string, string> = {
    login_ok: "Signed in",
    login_failed: "Sign-in failed",
    login_rate_limited: "Sign-in blocked (too many attempts)",
    user_created: "Account created",
    user_disabled: "Account disabled",
    user_enabled: "Account enabled",
    setup_link_created: "Link created",
    setup_link_used: "Password set with link",
    setup_link_rejected: "Dead link used",
    password_changed: "Password changed",
    password_change_failed: "Password change failed",
};

const details: Record<string, string> = {
    wrong_password: "wrong password",
    unknown_user: "unknown username",
    no_password: "account not set up yet",
    disabled: "account disabled",
    setup: "setup",
    reset: "reset",
    used: "already used",
    revoked: "replaced by a newer link",
    expired: "expired",
    unknown: "unknown link",
    malformed: "unknown link",
    user_gone: "account deleted",
};

/** Events that mean something went wrong, shown with a danger marker (and the word, not colour alone). */
export const failureEvents = new Set([
    "login_failed",
    "login_rate_limited",
    "setup_link_rejected",
    "password_change_failed",
]);

export function eventLabel(e: AuthEvent): string {
    return labels[e.event] ?? e.event;
}

export function eventDetail(e: AuthEvent): string {
    return details[e.detail] ?? e.detail;
}
