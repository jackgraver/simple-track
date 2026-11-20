// MOCK DATA: design preview for the /gym Insights section. Nothing here comes from the API yet.
// Notes on what each one needs, for when it's built for real:
// - volume / PRs / e1RM: aggregate logged sets per week/month (backend endpoint, or client-side from
//   a date-range logs query). e1RM = Epley (w × (1 + reps/30)) on each session's top set.
// - consistency: computable today from useGymWeek's schedule + activity (trained vs scheduled days).
// - callouts: rules over the same aggregates (days since a plan was last trained, top-set trend).

export type InsightTone = "up" | "down" | "flat";

export type InsightStat = {
    key: string;
    label: string;
    value: string;
    unit?: string;
    note: string;
    tone: InsightTone;
};

export type InsightCallout = {
    key: string;
    kind: "trend" | "attention";
    title: string;
    body: string;
};

export const mockInsightStats: InsightStat[] = [
    { key: "volume", label: "Volume this week", value: "18.2k", unit: "lb", note: "+6% vs last week", tone: "up" },
    { key: "prs", label: "PRs this month", value: "3", note: "Bench · Row · Squat", tone: "flat" },
    { key: "consistency", label: "Consistency", value: "92", unit: "%", note: "11 of 12 planned", tone: "flat" },
    { key: "e1rm", label: "Bench e1RM", value: "205", unit: "lb", note: "+10 lb in 4 wks", tone: "up" },
];

export const mockInsightCallouts: InsightCallout[] = [
    {
        key: "legs-gap",
        kind: "attention",
        title: "Legs last trained 9 days ago",
        body: "Friday is your most-missed day this month.",
    },
    {
        key: "ohp-trend",
        kind: "trend",
        title: "Overhead Press is climbing",
        body: "Top set up 15 lb over the last 6 sessions.",
    },
];
