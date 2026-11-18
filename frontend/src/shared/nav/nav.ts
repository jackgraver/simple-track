import type { Component } from "vue";
import type { RouteLocationNormalizedLoaded, RouteLocationRaw } from "vue-router";
import { Clapperboard, Dumbbell, Utensils, Wallet } from "lucide-vue-next";

export type NavItem = {
    label: string;
    /** Route name this item links to. */
    to: string;
    /** Carry the `?offset=` day selection across when navigating. */
    keepOffset?: boolean;
    /** Extra route names that should highlight this item (detail pages etc). */
    alsoActiveOn?: string[];
};

export type NavGroup = {
    key: string;
    label: string;
    icon: Component;
    items: NavItem[];
};

export const navGroups: NavGroup[] = [
    {
        key: "fitness",
        label: "Fitness",
        icon: Dumbbell,
        items: [
            {
                label: "Today",
                to: "gym",
                keepOffset: true,
                alsoActiveOn: [
                    "logging",
                    "logging-exercise",
                    "logging-cardio",
                    "logging-mobility",
                ],
            },
            {
                label: "Plans",
                to: "gym-plans",
                alsoActiveOn: ["gym-plan-detail", "gym-program-day-detail"],
            },
            { label: "Progression", to: "progression" },
            { label: "Weight", to: "gym-weight", keepOffset: true },
            { label: "Steps", to: "gym-steps", keepOffset: true },
        ],
    },
    {
        key: "food",
        label: "Food",
        icon: Utensils,
        items: [
            {
                label: "Today",
                to: "diet",
                keepOffset: true,
                alsoActiveOn: ["diet-log", "diet-edit-planned"],
            },
            { label: "Targets", to: "diet-targets" },
            { label: "Water presets", to: "diet-water-presets" },
            { label: "Grocery", to: "grocery" },
        ],
    },
    {
        key: "finance",
        label: "Finance",
        icon: Wallet,
        items: [
            {
                label: "Accounts",
                to: "finances",
                alsoActiveOn: ["finance-account"],
            },
            { label: "Contribution room", to: "finance-rules" },
        ],
    },
    {
        key: "media",
        label: "Media",
        icon: Clapperboard,
        items: [
            {
                label: "Library",
                to: "media",
                alsoActiveOn: ["media-new", "media-detail"],
            },
        ],
    },
];

function routeNameOf(route: RouteLocationNormalizedLoaded): string {
    return typeof route.name === "string" ? route.name : "";
}

export function isItemActive(
    item: NavItem,
    route: RouteLocationNormalizedLoaded,
): boolean {
    const name = routeNameOf(route);
    return item.to === name || (item.alsoActiveOn?.includes(name) ?? false);
}

export function activeItem(
    route: RouteLocationNormalizedLoaded,
): { group: NavGroup; item: NavItem } | null {
    for (const group of navGroups) {
        for (const item of group.items) {
            if (isItemActive(item, route)) return { group, item };
        }
    }
    return null;
}

export function activeGroup(
    route: RouteLocationNormalizedLoaded,
): NavGroup | null {
    return activeItem(route)?.group ?? null;
}

/** True when the route is one of the top-level entries shown in the nav. */
export function isNavRoot(route: RouteLocationNormalizedLoaded): boolean {
    const name = routeNameOf(route);
    return navGroups.some((g) => g.items.some((i) => i.to === name));
}

export function itemLocation(
    item: NavItem,
    route: RouteLocationNormalizedLoaded,
): RouteLocationRaw {
    const offset = route.query.offset;
    if (item.keepOffset && typeof offset === "string" && offset !== "") {
        return { name: item.to, query: { offset } };
    }
    return { name: item.to };
}
