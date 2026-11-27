-- The SQL patterns C lanes copy (docs/multi-user/DECISIONS.md, "B3 owner kit"), run against
-- the tables in rls_patterns_before.sql by TestRLSPatterns. Keep DECISIONS and this file in step.
-- app_user_id(), app_is_admin() and catalog_guard() come from migration 0011.

-- ── Owned table (parent) ─────────────────────────────────────────────────────────────────────
-- 1. column, 2. backfill to the admin, 3. NOT NULL + FK + index, 4. per-user uniques,
-- 5. (id, user_id) key for children, 6. RLS.
ALTER TABLE public.example_lists ADD COLUMN user_id bigint;
DO $$
DECLARE
  admins bigint;
  owner_id bigint;
BEGIN
  IF EXISTS (SELECT 1 FROM public.example_lists WHERE user_id IS NULL) THEN
    SELECT count(*), min(id) INTO admins, owner_id FROM public.users WHERE is_admin AND deleted_at IS NULL;
    IF admins <> 1 THEN
      RAISE EXCEPTION 'example_lists: % live admin accounts, expected exactly one to own the existing rows', admins;
    END IF;
    UPDATE public.example_lists SET user_id = owner_id WHERE user_id IS NULL;
  END IF;
END $$;
ALTER TABLE public.example_lists ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.example_lists
  ADD CONSTRAINT fk_example_lists_user FOREIGN KEY (user_id) REFERENCES public.users (id);
CREATE INDEX idx_example_lists_user_id ON public.example_lists USING btree (user_id);
DROP INDEX public.uq_example_lists_name;
CREATE UNIQUE INDEX uq_example_lists_user_name ON public.example_lists USING btree (user_id, name);
DROP INDEX public.uq_example_lists_client_id;
CREATE UNIQUE INDEX uq_example_lists_user_client_id ON public.example_lists USING btree (user_id, client_id)
  WHERE client_id IS NOT NULL;
ALTER TABLE public.example_lists ADD CONSTRAINT uq_example_lists_id_user UNIQUE (id, user_id);
ALTER TABLE public.example_lists ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.example_lists
  USING (user_id = public.app_user_id())
  WITH CHECK (user_id = public.app_user_id());

-- ── Owned child ──────────────────────────────────────────────────────────────────────────────
-- user_id copied from the parent (in the same migration, after the parent), then the parent FK
-- becomes composite so a child can only hang off a parent of the same user. Same ON DELETE as
-- before. Queries filter the child by its own user_id (Owned), never by joining the parent.
ALTER TABLE public.example_list_items ADD COLUMN user_id bigint;
UPDATE public.example_list_items c SET user_id = p.user_id
  FROM public.example_lists p WHERE p.id = c.list_id AND c.user_id IS NULL;
ALTER TABLE public.example_list_items ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.example_list_items
  ADD CONSTRAINT fk_example_list_items_user FOREIGN KEY (user_id) REFERENCES public.users (id);
ALTER TABLE public.example_list_items DROP CONSTRAINT fk_example_lists_items;
ALTER TABLE public.example_list_items
  ADD CONSTRAINT fk_example_lists_items FOREIGN KEY (list_id, user_id)
  REFERENCES public.example_lists (id, user_id) ON DELETE CASCADE;
CREATE INDEX idx_example_list_items_user_list ON public.example_list_items USING btree (user_id, list_id);
ALTER TABLE public.example_list_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.example_list_items
  USING (user_id = public.app_user_id())
  WITH CHECK (user_id = public.app_user_id());

-- ── Shared catalog ───────────────────────────────────────────────────────────────────────────
-- Everyone signed in reads; anyone adds as themselves; creator or admin edits; only admin
-- deletes (hard DELETE by policy, soft delete by catalog_guard).
ALTER TABLE public.example_foods ADD COLUMN created_by bigint;
DO $$
DECLARE
  admins bigint;
  owner_id bigint;
BEGIN
  IF EXISTS (SELECT 1 FROM public.example_foods WHERE created_by IS NULL) THEN
    SELECT count(*), min(id) INTO admins, owner_id FROM public.users WHERE is_admin AND deleted_at IS NULL;
    IF admins <> 1 THEN
      RAISE EXCEPTION 'example_foods: % live admin accounts, expected exactly one to own the existing rows', admins;
    END IF;
    UPDATE public.example_foods SET created_by = owner_id WHERE created_by IS NULL;
  END IF;
END $$;
ALTER TABLE public.example_foods ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE public.example_foods
  ADD CONSTRAINT fk_example_foods_created_by FOREIGN KEY (created_by) REFERENCES public.users (id);
CREATE INDEX idx_example_foods_created_by ON public.example_foods USING btree (created_by);
ALTER TABLE public.example_foods ENABLE ROW LEVEL SECURITY;
CREATE POLICY catalog_read ON public.example_foods FOR SELECT
  USING (public.app_user_id() IS NOT NULL);
CREATE POLICY catalog_add ON public.example_foods FOR INSERT
  WITH CHECK (created_by = public.app_user_id());
CREATE POLICY catalog_edit ON public.example_foods FOR UPDATE
  USING (created_by = public.app_user_id() OR public.app_is_admin())
  WITH CHECK (created_by = public.app_user_id() OR public.app_is_admin());
CREATE POLICY catalog_delete ON public.example_foods FOR DELETE
  USING (public.app_is_admin());
CREATE TRIGGER catalog_guard BEFORE UPDATE ON public.example_foods
  FOR EACH ROW EXECUTE FUNCTION public.catalog_guard();
