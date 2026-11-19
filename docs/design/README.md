# Design system

The permanent home of how simple-track's UI should look and behave. Migration work (the
ui-overhaul project) is tracked separately in `docs/projects/ui-overhaul/`.

| File | Read it when |
|---|---|
| [PRINCIPLES.md](PRINCIPLES.md) | Always, once: what the app is, the feel, density, colour, motion, mobile vs desktop, states, a11y |
| [RULES.md](RULES.md) | Before any UI change: the hard rules, the lint and how to verify visually |
| [TOKENS.md](TOKENS.md) | When you pick a colour, size, radius, shadow, z-index or duration |
| [PRIMITIVES.md](PRIMITIVES.md) | When you need a button, input, card, row, dialog, toast or state component |

Status (2026-11-19): principles and rules are in force, and the lint enforces them for new code.
Tokens and the `shared/ui/` primitives (Reka UI, headless, styled with our tokens) are built,
and every page uses them. Only the global element CSS in `App.vue` is left (phase 6).

Agent entry points: `frontend/CLAUDE.md` (loaded automatically when working under `frontend/`)
and the `ui` skill (`.claude/skills/ui/`).
