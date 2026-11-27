import type { Transition } from "motion-v";
import { tunable } from "./tune";

// Motion tokens for motion-v, the JS counterpart of the duration utilities in docs/design/TOKENS.md.
// Springs with no bounce: they settle like ease-out but carry a swipe's velocity into the animation.
// Each knob's `value` is the shipped default; `?tune` in dev opens live sliders for them (tune.ts).

/** Week carousel: the slide between weeks and the selected-day outline (owner-picked 350ms), and the
 * swipe distance (share of the strip's width) or speed (px/s) that commits to the next week. */
export const carousel = tunable("Week carousel", {
    duration: {
        value: 0.35, min: 0.1, max: 1.2, step: 0.01,
        hint: "Roughly how long a week change takes to slide, in seconds. The selected-day outline moves at the same speed.",
    },
    bounce: {
        value: 0, min: 0, max: 0.5, step: 0.01,
        hint: "How much the slide overshoots and springs back at the end. 0 stops dead on target.",
    },
    swipeDistance: {
        value: 0.25, min: 0.05, max: 0.6, step: 0.01,
        hint: "How far you have to drag before letting go changes the week, as a share of the strip's width (0.25 = a quarter).",
    },
    swipeVelocity: {
        value: 400, min: 50, max: 2000, step: 10,
        hint: "How fast a flick (px per second) has to be to change the week even if it was short. Lower = easier to flick.",
    },
});

// Imperative animate() calls don't see <MotionConfig>, so springs check the OS setting themselves.
const reducedQuery = typeof matchMedia === "function" ? matchMedia("(prefers-reduced-motion: reduce)") : null;
export const reduceMotion = (): boolean => reducedQuery?.matches ?? false;

/** Build springs at the moment of animating, so a tuned value applies to the next animation. */
const spring = (duration: number, bounce: number, velocity = 0): Transition =>
    reduceMotion() ? { duration: 0 } : { type: "spring", visualDuration: duration, bounce, velocity };

export const carouselSpring = (): Transition => spring(carousel.duration, carousel.bounce);

/** Mobile nav drawer: the slide in/out (carrying a release's velocity, px/s), and when a released
 * swipe opens it: a fling faster than `flingVelocity` px/s goes its way, otherwise it opens if more
 * than `openShare` of the panel is showing. */
export const drawer = tunable("Nav drawer", {
    duration: {
        value: 0.3, min: 0.1, max: 1, step: 0.01,
        hint: "Roughly how long the drawer takes to slide open or closed (menu button, Esc, or after you let go of a swipe), in seconds.",
    },
    flingVelocity: {
        value: 400, min: 100, max: 2000, step: 10,
        hint: "Release speed (px per second) above which the drawer goes the way you flicked, however far it was pulled. Lower = easier to flick.",
    },
    openShare: {
        value: 0.5, min: 0.2, max: 0.8, step: 0.01,
        hint: "For a slow release: how much of the drawer has to be showing for it to stay open (0.5 = half), otherwise it closes.",
    },
});
// No bounce knob: the open edge is held (an overshoot would show a gap) and a closing one is off screen.
export const drawerSpring = (velocity = 0): Transition => spring(drawer.duration, 0, velocity);

/** Sliding highlight (desktop nav): how fast the hover and active pills move, and the press shrink. */
export const highlight = tunable("Nav highlight", {
    duration: {
        value: 0.1, min: 0.05, max: 0.6, step: 0.01,
        hint: "Roughly how long the hover highlight and the current-page highlight take to slide to another row, in seconds.",
    },
    pressScale: {
        value: 0.97, min: 0.85, max: 1, step: 0.005,
        hint: "How small a row shrinks while the mouse button is held on it (1 = no shrink, 0.9 = 10% smaller).",
    },
    pressDuration: {
        value: 0.1, min: 0.03, max: 0.4, step: 0.01,
        hint: "How long the press shrink and the spring back on release take, in seconds.",
    },
});
export const highlightSpring = (): Transition => spring(highlight.duration, 0);
export const pressTransition = (): Transition =>
    reduceMotion() ? { duration: 0 } : { duration: highlight.pressDuration, ease: "easeOut" };

// Respect the OS reduced-motion setting for JS animations too (the CSS rule in style.css
// only clamps CSS transitions). Mounted once in App.vue.
export { MotionConfig } from "motion-v";
