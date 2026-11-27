-- Owner kit (lane B3, docs/multi-user/DECISIONS.md "B3 owner kit").
--
-- 1. app_user_id() / app_is_admin(): what row-level security policies compare against. They read
--    the request identity the server sets per transaction (set_config('app.user_id', ..., true)).
--    Unset reads as NULL, so a query without an identity matches no owned rows. The NULLIF
--    matters: once a session has set app.user_id in some earlier transaction, the setting reads
--    as '' (not NULL) afterwards, and ''::bigint would be an error instead of "no rows".
-- 2. catalog_guard(): trigger function for shared-catalog tables (below).
-- 3. drink_size_presets becomes an owned table, the reference conversion every C lane copies:
--    user_id (backfilled to the admin), FK, index, UNIQUE (id, user_id) for children's
--    composite FKs, RLS policy. It has no per-name unique to make per-user.

-- migrate:up
CREATE FUNCTION public.app_user_id() RETURNS bigint
  LANGUAGE sql STABLE PARALLEL SAFE
  AS $$ SELECT NULLIF(current_setting('app.user_id', true), '')::bigint $$;

CREATE FUNCTION public.app_is_admin() RETURNS boolean
  LANGUAGE sql STABLE PARALLEL SAFE
  AS $$ SELECT COALESCE(NULLIF(current_setting('app.is_admin', true), '')::boolean, false) $$;

-- Shared-catalog rows (foods, exercises, ...): BEFORE UPDATE trigger for what policies can't
-- express. gorm's soft delete is an UPDATE of deleted_at, which the DELETE policy never sees, so
-- only an admin may set or clear deleted_at; and created_by never changes. Applies only where
-- row-level security does (the app role), so migrations and seeds on the owner role are free.
-- Needs created_by and deleted_at columns. Lanes attach it per table (DECISIONS.md, B3 owner kit).
CREATE FUNCTION public.catalog_guard() RETURNS trigger
  LANGUAGE plpgsql
  AS $$
BEGIN
  IF NOT row_security_active(TG_RELID) THEN
    RETURN NEW;
  END IF;
  IF NEW.created_by IS DISTINCT FROM OLD.created_by THEN
    RAISE EXCEPTION '%: created_by can''t change', TG_TABLE_NAME USING ERRCODE = 'insufficient_privilege';
  END IF;
  IF NEW.deleted_at IS DISTINCT FROM OLD.deleted_at AND NOT public.app_is_admin() THEN
    RAISE EXCEPTION '%: only an admin deletes catalog rows', TG_TABLE_NAME USING ERRCODE = 'insufficient_privilege';
  END IF;
  RETURN NEW;
END $$;

ALTER TABLE public.drink_size_presets ADD COLUMN user_id bigint;

-- Backfill existing rows (soft-deleted ones too) to the admin: the owner, the only account
-- before D1. Rows without exactly one admin to give them to stop the migration.
DO $$
DECLARE
  admins bigint;
  owner_id bigint;
BEGIN
  IF EXISTS (SELECT 1 FROM public.drink_size_presets WHERE user_id IS NULL) THEN
    SELECT count(*), min(id) INTO admins, owner_id FROM public.users
      WHERE is_admin AND deleted_at IS NULL;
    IF admins <> 1 THEN
      RAISE EXCEPTION 'drink_size_presets: % live admin accounts, expected exactly one to own the existing rows (docs/multi-user/DECISIONS.md, B3)', admins;
    END IF;
    UPDATE public.drink_size_presets SET user_id = owner_id WHERE user_id IS NULL;
  END IF;
END $$;

ALTER TABLE public.drink_size_presets ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.drink_size_presets
  ADD CONSTRAINT fk_drink_size_presets_user FOREIGN KEY (user_id) REFERENCES public.users (id);
CREATE INDEX idx_drink_size_presets_user_id ON public.drink_size_presets USING btree (user_id);
-- Target for composite FKs from rows that point at a preset (water_logs.preset_id, lane C1), so a
-- row can only reference a preset of the same user.
ALTER TABLE public.drink_size_presets
  ADD CONSTRAINT uq_drink_size_presets_id_user UNIQUE (id, user_id);

ALTER TABLE public.drink_size_presets ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.drink_size_presets
  USING (user_id = public.app_user_id())
  WITH CHECK (user_id = public.app_user_id());

-- migrate:down
DROP POLICY owner ON public.drink_size_presets;
ALTER TABLE public.drink_size_presets DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.drink_size_presets DROP CONSTRAINT uq_drink_size_presets_id_user;
DROP INDEX public.idx_drink_size_presets_user_id;
ALTER TABLE public.drink_size_presets DROP CONSTRAINT fk_drink_size_presets_user;
ALTER TABLE public.drink_size_presets DROP COLUMN user_id;
DROP FUNCTION public.catalog_guard();
DROP FUNCTION public.app_is_admin();
DROP FUNCTION public.app_user_id();
