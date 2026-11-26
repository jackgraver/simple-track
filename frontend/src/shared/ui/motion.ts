import type { Transition } from "motion-v";
import { tunable } from "./tune";

// Motion tokens for motion-v, the JS counterpart of the duration utilities in docs/design/TOKENS.md.
// Springs with no bounce: they settle like ease-out but carry a swipe's velocity into the animation.
// Each knob's `value` is the shipped default; `?tune` in dev opens live sliders for them (tune.ts).

/** Week carousel: the slide between weeks and the selected-day outline (owner-picked 350ms), and the
 * swipe distance (share of the strip's width) or speed (px/s) that commits to the next week. */
export const carousel = tunable("Week carousel", {
    duration: { value: 0.35, min: 0.1, max: 1.2, step: 0.01 },
    bounce: { value: 0, min: 0, max: 0.5, step: 0.01 },
    swipeDistance: { value: 0.25, min: 0.05, max: 0.6, step: 0.01 },
    swipeVelocity: { value: 400, min: 50, max: 2000, step: 10 },
});

/** Call at the moment of animating, so a tuned value applies to the next slide. */
export const carouselSpring = (): Transition => ({ type: "spring", visualDuration: carousel.duration, bounce: carousel.bounce });

// Respect the OS reduced-motion setting for JS animations too (the CSS rule in style.css
// only clamps CSS transitions). Mounted once in App.vue.
export { MotionConfig } from "motion-v";
