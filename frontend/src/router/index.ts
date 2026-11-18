import { createRouter, createWebHistory, START_LOCATION } from "vue-router";
import type { RouteRecordRaw } from "vue-router";
import { authStatus } from "~/composables/auth/session";
import { resolveAuthSession } from "~/composables/auth/useAuth";
import { loadLastRoute, saveLastRoute } from "~/utils/lastRoute";

declare module "vue-router" {
    interface RouteMeta {
        breadcrumb?: string;
        /** Show the day picker (pages that read `?offset=`). Inherited by children. */
        dated?: boolean;
        /** Full-screen flow: hide the app nav (sidebar, mobile header, tab bar). */
        focus?: boolean;
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
        meta: { breadcrumb: "Diet", dated: true },
        children: [
            {
                path: "log",
                name: "diet-log",
                component: () => import("~/pages/diet/logmeal/logmeal.vue"),
                meta: { breadcrumb: "Log meal" },
            },
            {
                path: "edit-planned",
                name: "diet-edit-planned",
                component: () => import("~/pages/diet/edit-planned/index.vue"),
                meta: { breadcrumb: "Edit planned" },
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
        meta: { breadcrumb: "Gym", dated: true },
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
                meta: { breadcrumb: "Cardio" },
            },
            {
                path: "logging/mobility/:slot",
                name: "logging-mobility",
                component: () =>
                    import("~/pages/gym/logging/exercise/[id].vue"),
                meta: { breadcrumb: "Mobility" },
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
                meta: { breadcrumb: "Weight" },
            },
            {
                path: "steps",
                name: "gym-steps",
                component: () => import("~/pages/gym/steps/index.vue"),
                meta: { breadcrumb: "Steps" },
            },
        ],
    },
    {
        path: "/grocery",
        name: "grocery",
        component: () => import("~/pages/grocery/index.vue"),
        meta: { breadcrumb: "Grocery" },
    },
    {
        path: "/finances",
        name: "finances",
        component: () => import("~/pages/finances/index.vue"),
        meta: { breadcrumb: "Finances" },
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
        meta: { breadcrumb: "Media" },
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
        path: "/settings/profile",
        name: "settings-profile",
        component: () => import("~/pages/settings/profile/index.vue"),
        meta: { breadcrumb: "Profile" },
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
    const saved = loadLastRoute();
    if (!saved) return null;
    const target = router.resolve(saved);
    if (!target.name || target.name === "not-found" || target.name === "signin" || target.name === "launch") {
        return null;
    }
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
        const saved = resolveSavedRoute();
        if (saved && saved !== to.fullPath) return saved;
    }
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
    return true;
});

router.afterEach((to, _from, failure) => {
    if (failure || to.name === "signin" || to.name === "not-found" || to.name === "launch") return;
    saveLastRoute(to.fullPath);
});
