import { createRouter, createWebHistory, START_LOCATION } from "vue-router";
import type { RouteRecordRaw } from "vue-router";
import { authStatus, isAdmin } from "~/composables/auth/session";
import { resolveAuthSession } from "~/composables/auth/useAuth";
import { isExpiredLoggingResume, loadLastRoute, saveLastRoute } from "~/utils/lastRoute";
import { dietQueries } from "~/api/diet/queries";
import { mediaQueries } from "~/api/media/queries";
import { moneyQueries } from "~/api/money/queries";
import { queryClient } from "~/api/queryClient";
import { trackingQueries } from "~/api/tracking/queries";
import { ROLLING_YEAR_ACTIVITY, workoutQueries } from "~/api/workout/queries";
import { offsetToDate } from "~/api/workout/dates";
import { parseDietDayOffsetQuery } from "~/utils/dateUtil";
import { prefetchRoute, warmNavigation, type RouteQueries } from "./prefetch";

declare module "vue-router" {
    interface RouteMeta {
        breadcrumb?: string;
        /** Show the day picker (pages that read `?offset=`). Inherited by children. */
        dated?: boolean;
        /** Page supplies its own header: hide the mobile top bar (sidebar and drawer stay). */
        focus?: boolean;
        /** Route name of this page's parent. Header Back goes there instead of through history. */
        backTo?: string;
        /** Reachable signed out (the setup link page). */
        public?: boolean;
        /** Admin only; others are sent home (the API refuses them anyway). */
        admin?: boolean;
        /** Multi-column workspace: wider desktop content column (`spacing-content-wide`). */
        wide?: boolean;
        /**
         * Primary queries of this route (`queryOptions` objects built from the target location).
         * Prefetched on nav intent (NavLink) and awaited briefly in `beforeResolve`. Only the matched
         * leaf record's own value is used: child routes do not inherit it.
         */
        queries?: RouteQueries;
    }
}

const routes: RouteRecordRaw[] = [
    {
        path: "/",
        redirect: { name: "gym" },
    },
    {
        // The manifest start_url. Resumes the last route, else the default page.
        path: "/launch",
        name: "launch",
        redirect: () => resolveSavedRoute() ?? { name: "gym" },
    },
    {
        path: "/home",
        name: "home",
        component: () => import("~/pages/home/index.vue"),
        meta: { breadcrumb: "Home" },
    },
    {
        path: "/diet",
        name: "diet",
        component: () => import("~/pages/diet/index.vue"),
        meta: {
            breadcrumb: "Diet",
            dated: true,
            queries: (to) => [dietQueries.day(parseDietDayOffsetQuery(to.query.offset))],
        },
        children: [
            {
                path: "log",
                name: "diet-log",
                component: () => import("~/pages/diet/logmeal/logmeal.vue"),
                meta: { breadcrumb: "Log meal", wide: true },
            },
            {
                path: "edit-planned",
                name: "diet-edit-planned",
                component: () => import("~/pages/diet/edit-planned/index.vue"),
                meta: { breadcrumb: "Plan meals", wide: true, dated: false },
            },
            {
                path: "targets",
                name: "diet-targets",
                component: () => import("~/pages/diet/targets/index.vue"),
                meta: { breadcrumb: "Targets", dated: false },
            },
            {
                path: "water-presets",
                name: "diet-water-presets",
                component: () => import("~/pages/diet/water-presets/index.vue"),
                meta: { breadcrumb: "Water presets", dated: false },
            },
        ],
    },
    {
        path: "/gym",
        name: "gym",
        component: () => import("~/pages/gym/index.vue"),
        meta: {
            breadcrumb: "Gym",
            // No dated: the home page's week strip is its day picker. Children that log or read a
            // day opt back in.
            // The week strip and heatmap read programs + activity: start them with the day so they
            // don't land after first paint and push the Today card down.
            queries: (to) => [
                workoutQueries.day(offsetToDate(parseDietDayOffsetQuery(to.query.offset))),
                workoutQueries.programs(),
                workoutQueries.activity(ROLLING_YEAR_ACTIVITY),
            ],
        },
        children: [
            {
                path: "plans",
                name: "gym-plans",
                component: () => import("~/pages/gym/plans/index.vue"),
                meta: { breadcrumb: "Plans", dated: false },
            },
            {
                path: "plans/:id(\\d+)",
                name: "gym-plan-detail",
                component: () => import("~/pages/gym/plans/[id].vue"),
                meta: { breadcrumb: "Plan", dated: false },
            },
            {
                path: "plans/:programId(\\d+)/days/:id(\\d+)",
                name: "gym-program-day-detail",
                component: () => import("~/pages/gym/plans/[id].vue"),
                meta: { breadcrumb: "Plan day", dated: false },
            },
            {
                path: "logging",
                name: "logging",
                component: () => import("~/pages/gym/logging/index.vue"),
                meta: {
                    breadcrumb: "Logging",
                    dated: false,
                    backTo: "gym",
                },
            },
            {
                path: "logging/:id(\\d+)",
                name: "logging-exercise",
                component: () =>
                    import("~/pages/gym/logging/exercise/[id].vue"),
                meta: {
                    breadcrumb: "Exercise",
                    dated: false,
                    focus: true,
                },
            },
            {
                path: "logging/cardio",
                name: "logging-cardio",
                component: () =>
                    import("~/pages/gym/logging/exercise/[id].vue"),
                meta: { breadcrumb: "Cardio", dated: true, focus: true },
            },
            {
                path: "logging/mobility/:slot",
                name: "logging-mobility",
                component: () =>
                    import("~/pages/gym/logging/exercise/[id].vue"),
                meta: { breadcrumb: "Mobility", dated: true, focus: true },
            },
            {
                path: "progression",
                name: "progression",
                component: () =>
                    import("~/pages/gym/progression/progression.vue"),
                meta: { breadcrumb: "Progression", dated: false },
            },
            {
                path: "weight",
                name: "gym-weight",
                component: () => import("~/pages/gym/weight/index.vue"),
                meta: { breadcrumb: "Weight", dated: true },
            },
            {
                path: "steps",
                name: "gym-steps",
                component: () => import("~/pages/gym/steps/index.vue"),
                meta: { breadcrumb: "Steps", dated: true },
            },
        ],
    },
    {
        path: "/grocery",
        name: "grocery",
        component: () => import("~/pages/grocery/index.vue"),
        meta: { breadcrumb: "Grocery", queries: () => [trackingQueries.groceryItems()] },
    },
    {
        path: "/finances",
        name: "finances",
        component: () => import("~/pages/finances/index.vue"),
        meta: {
            breadcrumb: "Finances",
            queries: () => [moneyQueries.accounts(false), moneyQueries.netWorth()],
        },
    },
    {
        path: "/finances/accounts/:id(\\d+)",
        name: "finance-account",
        component: () => import("~/pages/finances/account.vue"),
        meta: { breadcrumb: "Account" },
    },
    {
        path: "/finances/rules",
        name: "finance-rules",
        component: () => import("~/pages/finances/rules.vue"),
        meta: { breadcrumb: "Contribution room" },
    },
    {
        path: "/finances/config",
        redirect: { name: "finance-rules" },
    },
    {
        path: "/finance",
        redirect: { name: "finances" },
    },
    {
        path: "/media",
        name: "media",
        component: () => import("~/pages/media/index.vue"),
        meta: {
            breadcrumb: "Media",
            // Same default filters as the page, so the keys match.
            queries: () => [mediaQueries.list({ sort: "updated" }), mediaQueries.stats()],
        },
    },
    {
        path: "/media/new",
        name: "media-new",
        component: () => import("~/pages/media/[id].vue"),
        meta: { breadcrumb: "Add media" },
    },
    {
        path: "/media/:id(\\d+)",
        name: "media-detail",
        component: () => import("~/pages/media/[id].vue"),
        meta: { breadcrumb: "Media item" },
    },
    {
        path: "/auth/signin",
        name: "signin",
        component: () => import("~/pages/auth/signin/index.vue"),
        meta: { focus: true },
    },
    {
        // One-time link to set or reset a password. Public; the token stays out of saved routes.
        path: "/setup/:token",
        name: "setup",
        component: () => import("~/pages/setup/index.vue"),
        meta: { focus: true, public: true },
    },
    {
        path: "/settings/account",
        name: "settings-account",
        component: () => import("~/pages/settings/account/index.vue"),
        meta: { breadcrumb: "Account" },
    },
    {
        path: "/settings/users",
        name: "settings-users",
        component: () => import("~/pages/settings/users/index.vue"),
        meta: { breadcrumb: "Users", admin: true },
    },
    {
        path: "/settings/profile",
        name: "settings-profile",
        component: () => import("~/pages/settings/profile/index.vue"),
        meta: { breadcrumb: "Body profile" },
    },
    {
        path: "/:pathMatch(.*)*",
        name: "not-found",
        component: () => import("~/pages/not-found/index.vue"),
    },
];

function isStandalone(): boolean {
    return (
        window.matchMedia("(display-mode: standalone)").matches ||
        (navigator as Navigator & { standalone?: boolean }).standalone === true
    );
}

function resolveSavedRoute(): string | null {
    const entry = loadLastRoute();
    if (!entry) return null;
    const target = router.resolve(entry.path);
    if (!target.name || target.meta.public || target.name === "not-found" || target.name === "signin" || target.name === "launch") {
        return null;
    }
    if (isExpiredLoggingResume(entry)) return router.resolve({ name: "gym" }).fullPath;
    return target.fullPath;
}

export const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach(async (to, from) => {
    // iOS may launch an installed PWA at the URL it was added from rather than
    // the manifest start_url. Resume the last route on the first navigation of
    // any standalone launch, whatever path it lands on. The saved route is the
    // last page visited, so this matches what a plain reload would show.
    if (from === START_LOCATION && to.name !== "launch" && isStandalone()) {
        // The saved route is per user, so who is signed in must be known first.
        if (authStatus.value === "unknown") await resolveAuthSession();
        const saved = resolveSavedRoute();
        if (saved && saved !== to.fullPath) return saved;
    }
    if (to.meta.public) return true;
    if (to.name === "signin") {
        if (authStatus.value === "unknown") await resolveAuthSession();
        if (authStatus.value === "authenticated") {
            return { name: "gym" };
        }
        return true;
    }
    if (authStatus.value === "unknown") await resolveAuthSession();
    if (authStatus.value !== "authenticated") {
        return { name: "signin", query: { redirect: to.fullPath } };
    }
    if (to.meta.admin && !isAdmin.value) return { name: "gym" };
    // App start: begin the landing page's queries while its component chunk loads (redirects are
    // already applied here, so `/` warms the gym day and `/media` warms nothing day-related).
    if (from === START_LOCATION) prefetchRoute(queryClient, to);
    return true;
});

// Warm the target's primary queries, waiting at most NAV_BUDGET_MS so a slow network never blocks
// navigation: the page renders from cache if ready and otherwise shows its own cold-load state.
// Only for authenticated targets (the auth guard above has already redirected otherwise).
router.beforeResolve((to, from) => warmNavigation(queryClient, to, from));

router.afterEach((to, _from, failure) => {
    if (failure || to.meta.public || to.name === "signin" || to.name === "not-found" || to.name === "launch") return;
    saveLastRoute(to.fullPath);
});
