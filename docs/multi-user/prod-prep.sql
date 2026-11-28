-- D1 prod prep: run ONCE on the prod database, as the owner, before the multi-user deploy
-- (ROLLOUT 3.0). Found by the D1 rehearsal on the 2026-11-27 backup; not a migration because it
-- only applies to prod's drifted schema and migrations 0009+ are already applied on every dev DB.
--
-- 1. Prod's tables still carry a legacy `uuid user_id` column from an abandoned earlier
--    multi-user design (baseline 0001 never had them). Every C lane adds a `bigint user_id` with
--    the same name, so their migrations fail on prod. Nothing reads the legacy columns (the app's
--    models don't have the field; a fresh database doesn't have them). Dropping them also drops
--    their indexes. Only columns of type uuid are dropped, so it is safe to re-run.
-- 2. Soft-delete the unused second account ("gen"), so migration 0009 sees exactly one live
--    account (the owner). Edit the username below if it isn't the one you mean; the owner
--    (username `jack`) is never touched.
BEGIN;

DO $$
DECLARE r record;
BEGIN
  FOR r IN
    SELECT c.table_name FROM information_schema.columns c
    WHERE c.table_schema = 'public' AND c.column_name = 'user_id' AND c.data_type = 'uuid'
      AND c.table_name IN ('body_weight_logs', 'days', 'drink_size_presets', 'meals', 'plans',
                           'saved_meals', 'step_logs', 'water_logs', 'workout_logs', 'workout_plans')
  LOOP
    EXECUTE format('ALTER TABLE public.%I DROP COLUMN user_id', r.table_name);
    RAISE NOTICE 'dropped legacy uuid user_id from %', r.table_name;
  END LOOP;
END $$;

UPDATE public.users SET deleted_at = now() WHERE username = 'gen' AND deleted_at IS NULL;

-- Expect exactly 1 live account (jack) and 0 remaining uuid user_id columns on these tables.
SELECT (SELECT count(*) FROM public.users WHERE deleted_at IS NULL) AS live_accounts,
       (SELECT count(*) FROM information_schema.columns
         WHERE table_schema = 'public' AND column_name = 'user_id' AND data_type = 'uuid'
           AND table_name <> 'users' AND table_name <> 'group_members') AS legacy_columns_left;
COMMIT;
