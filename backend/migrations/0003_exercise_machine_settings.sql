-- Per-exercise machine settings (seat, pad, pulley...) shown while logging.
-- A JSON array of {"label": "...", "value": "..."}; NULL means none set.

-- migrate:up
ALTER TABLE public.exercises ADD COLUMN machine_settings jsonb;

-- migrate:down
ALTER TABLE public.exercises DROP COLUMN machine_settings;
