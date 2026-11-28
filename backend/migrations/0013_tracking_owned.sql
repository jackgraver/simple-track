-- Tracking becomes per-user (lane C1, docs/multi-user/DECISIONS.md "C1 tracking").
--
-- body_weight_logs, step_logs, water_logs, grocery_items and user_profiles are owned tables:
-- user_id NOT NULL (existing rows go to the admin, the only account before D1), FK, per-user
-- uniques, row-level security. Follows the B3 patterns (internal/database/testdata/rls_patterns.sql).
--
-- * water_logs.preset_id becomes the composite FK (preset_id, user_id) -> drink_size_presets
--   (id, user_id), which 0011 prepared: FK checks bypass RLS, so a plain FK would let one user
--   log water against another user's preset.
-- * user_profiles is one row per user: unique on user_id among live rows. The app only ever read
--   the oldest live row, so if the table somehow holds more than one live row, the migration
--   soft-deletes the newer ones (kept, not dropped) before the unique index.
-- * The unique indexes on (date) and (client_id) become (user_id, date) / (user_id, client_id).

-- migrate:up
ALTER TABLE public.body_weight_logs ADD COLUMN user_id bigint;
ALTER TABLE public.step_logs        ADD COLUMN user_id bigint;
ALTER TABLE public.water_logs       ADD COLUMN user_id bigint;
ALTER TABLE public.grocery_items    ADD COLUMN user_id bigint;
ALTER TABLE public.user_profiles    ADD COLUMN user_id bigint;

-- Backfill every row (soft-deleted ones too) to the admin. A table with rows to give away and not
-- exactly one live admin stops the migration instead of guessing (README "Migrating existing data").
DO $$
DECLARE
  admins bigint;
  owner_id bigint;
  tbl text;
  pending boolean;
BEGIN
  SELECT count(*), min(id) INTO admins, owner_id FROM public.users
    WHERE is_admin AND deleted_at IS NULL;
  FOREACH tbl IN ARRAY ARRAY['body_weight_logs', 'step_logs', 'water_logs', 'grocery_items', 'user_profiles'] LOOP
    EXECUTE format('SELECT EXISTS (SELECT 1 FROM public.%I WHERE user_id IS NULL)', tbl) INTO pending;
    IF pending THEN
      IF admins <> 1 THEN
        RAISE EXCEPTION '%: % live admin accounts, expected exactly one to own the existing rows (docs/multi-user/DECISIONS.md, B3)', tbl, admins;
      END IF;
      EXECUTE format('UPDATE public.%I SET user_id = $1 WHERE user_id IS NULL', tbl) USING owner_id;
    END IF;
  END LOOP;
END $$;

-- One live profile per user. Keep the oldest live row (the one GET /tracking/profile returned).
UPDATE public.user_profiles p SET deleted_at = now()
  WHERE p.deleted_at IS NULL
    AND p.id <> (SELECT min(q.id) FROM public.user_profiles q WHERE q.user_id = p.user_id AND q.deleted_at IS NULL);

ALTER TABLE public.body_weight_logs ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.step_logs        ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.water_logs       ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.grocery_items    ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.user_profiles    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE public.body_weight_logs
  ADD CONSTRAINT fk_body_weight_logs_user FOREIGN KEY (user_id) REFERENCES public.users (id);
ALTER TABLE public.step_logs
  ADD CONSTRAINT fk_step_logs_user FOREIGN KEY (user_id) REFERENCES public.users (id);
ALTER TABLE public.water_logs
  ADD CONSTRAINT fk_water_logs_user FOREIGN KEY (user_id) REFERENCES public.users (id);
ALTER TABLE public.grocery_items
  ADD CONSTRAINT fk_grocery_items_user FOREIGN KEY (user_id) REFERENCES public.users (id);
ALTER TABLE public.user_profiles
  ADD CONSTRAINT fk_user_profiles_user FOREIGN KEY (user_id) REFERENCES public.users (id);

-- Per-user uniques. (user_id, ...) also serves queries by user, so the date tables need no
-- separate user_id index.
DROP INDEX public.idx_body_weight_logs_date;
CREATE UNIQUE INDEX uq_body_weight_logs_user_date ON public.body_weight_logs USING btree (user_id, date);
DROP INDEX public.idx_step_logs_date;
CREATE UNIQUE INDEX uq_step_logs_user_date ON public.step_logs USING btree (user_id, date);

DROP INDEX public.uq_water_logs_client_id;
CREATE UNIQUE INDEX uq_water_logs_user_client_id ON public.water_logs USING btree (user_id, client_id)
  WHERE client_id IS NOT NULL;
CREATE INDEX idx_water_logs_user_date ON public.water_logs USING btree (user_id, date);

DROP INDEX public.uq_grocery_items_client_id;
CREATE UNIQUE INDEX uq_grocery_items_user_client_id ON public.grocery_items USING btree (user_id, client_id)
  WHERE client_id IS NOT NULL;
CREATE INDEX idx_grocery_items_user_id ON public.grocery_items USING btree (user_id);

CREATE UNIQUE INDEX uq_user_profiles_user_id ON public.user_profiles USING btree (user_id)
  WHERE deleted_at IS NULL;

-- A water log can only point at a preset of the same user (MATCH SIMPLE: no preset, no check).
ALTER TABLE public.water_logs DROP CONSTRAINT fk_water_logs_preset;
ALTER TABLE public.water_logs
  ADD CONSTRAINT fk_water_logs_preset FOREIGN KEY (preset_id, user_id)
  REFERENCES public.drink_size_presets (id, user_id);

ALTER TABLE public.body_weight_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.body_weight_logs
  USING (user_id = public.app_user_id())
  WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.step_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.step_logs
  USING (user_id = public.app_user_id())
  WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.water_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.water_logs
  USING (user_id = public.app_user_id())
  WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.grocery_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.grocery_items
  USING (user_id = public.app_user_id())
  WITH CHECK (user_id = public.app_user_id());
ALTER TABLE public.user_profiles ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.user_profiles
  USING (user_id = public.app_user_id())
  WITH CHECK (user_id = public.app_user_id());

-- migrate:down
DROP POLICY owner ON public.user_profiles;
ALTER TABLE public.user_profiles DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.grocery_items;
ALTER TABLE public.grocery_items DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.water_logs;
ALTER TABLE public.water_logs DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.step_logs;
ALTER TABLE public.step_logs DISABLE ROW LEVEL SECURITY;
DROP POLICY owner ON public.body_weight_logs;
ALTER TABLE public.body_weight_logs DISABLE ROW LEVEL SECURITY;

ALTER TABLE public.water_logs DROP CONSTRAINT fk_water_logs_preset;
ALTER TABLE public.water_logs
  ADD CONSTRAINT fk_water_logs_preset FOREIGN KEY (preset_id) REFERENCES public.drink_size_presets (id);

DROP INDEX public.uq_user_profiles_user_id;
DROP INDEX public.idx_grocery_items_user_id;
DROP INDEX public.uq_grocery_items_user_client_id;
CREATE UNIQUE INDEX uq_grocery_items_client_id ON public.grocery_items USING btree (client_id)
  WHERE client_id IS NOT NULL;
DROP INDEX public.idx_water_logs_user_date;
DROP INDEX public.uq_water_logs_user_client_id;
CREATE UNIQUE INDEX uq_water_logs_client_id ON public.water_logs USING btree (client_id)
  WHERE client_id IS NOT NULL;
DROP INDEX public.uq_step_logs_user_date;
CREATE UNIQUE INDEX idx_step_logs_date ON public.step_logs USING btree (date);
DROP INDEX public.uq_body_weight_logs_user_date;
CREATE UNIQUE INDEX idx_body_weight_logs_date ON public.body_weight_logs USING btree (date);

ALTER TABLE public.user_profiles    DROP CONSTRAINT fk_user_profiles_user;
ALTER TABLE public.grocery_items    DROP CONSTRAINT fk_grocery_items_user;
ALTER TABLE public.water_logs       DROP CONSTRAINT fk_water_logs_user;
ALTER TABLE public.step_logs        DROP CONSTRAINT fk_step_logs_user;
ALTER TABLE public.body_weight_logs DROP CONSTRAINT fk_body_weight_logs_user;

ALTER TABLE public.user_profiles    DROP COLUMN user_id;
ALTER TABLE public.grocery_items    DROP COLUMN user_id;
ALTER TABLE public.water_logs       DROP COLUMN user_id;
ALTER TABLE public.step_logs        DROP COLUMN user_id;
ALTER TABLE public.body_weight_logs DROP COLUMN user_id;
