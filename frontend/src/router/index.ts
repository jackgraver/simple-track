import { createRouter, createWebHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";
import { authStatus } from "~/composables/auth/session";
import { resolveAuthSession } from "~/composables/auth/useAuth";

declare module "vue-router" {
    interface RouteMeta {
        breadcrumb?: string;
        hideDateAndBreadcrumbs?: boolean;
    }
}

const routes: RouteRecordRaw[] = [
    {
        path: "/",
        redirect: { name: "gym" },
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
        meta: { breadcrumb: "Diet" },
        children: [
            {
                path: "log",
                name: "diet-log",
                component: () => import("~/pages/diet/logmeal/logmeal.vue"),
                meta: { breadcrumb: "Log meal", hideDateAndBreadcrumbs: false },
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
                meta: { breadcrumb: "Targets" },
            },
            {
                path: "water-presets",
                name: "diet-water-presets",
                component: () => import("~/pages/diet/water-presets/index.vue"),
                meta: { breadcrumb: "Water presets" },
            },
        ],
    },
    {
        path: "/gym",
        name: "gym",
        component: () => import("~/pages/gym/index.vue"),
        meta: { breadcrumb: "Gym" },
        children: [
            {
                path: "plans",
                name: "gym-plans",
                component: () => import("~/pages/gym/plans/index.vue"),
                meta: { breadcrumb: "Plans" },
            },
            {
                path: "plans/:id(\\d+)",
                name: "gym-plan-detail",
                component: () => import("~/pages/gym/plans/[id].vue"),
                meta: { breadcrumb: "Plan" },
            },
            {
                path: "plans/:programId(\\d+)/days/:id(\\d+)",
                name: "gym-program-day-detail",
                component: () => import("~/pages/gym/plans/[id].vue"),
                meta: { breadcrumb: "Plan day" },
            },
            {
                path: "logging",
                name: "logging",
                component: () => import("~/pages/gym/logging/index.vue"),
                meta: {
                    breadcrumb: "Logging",
                    hideDateAndBreadcrumbs: true,
                },
            },
            {
                path: "logging/:id(\\d+)",
                name: "logging-exercise",
                component: () =>
                    import("~/pages/gym/logging/exercise/[id].vue"),
                meta: {
                    breadcrumb: "Exercise",
                    hideDateAndBreadcrumbs: true,
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
                meta: { breadcrumb: "Progression" },
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
        path: "/auth/signin",
        name: "signin",
        component: () => import("~/pages/auth/signin/index.vue"),
    },
    {
        path: "/settings/profile",
        name: "settings-profile",
        component: () => import("~/pages/settings/profile/index.vue"),
        meta: { breadcrumb: "Profile" },
    },
];

export const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach(async (to) => {
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
