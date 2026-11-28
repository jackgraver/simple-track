-- Media (lane C5, docs/multi-user/DECISIONS.md "C5 media").
--
-- media_items becomes the shared catalog: title, type, creator, year, genres, cover, synopsis,
-- length, source + external_id (TMDB etc.) and created_by, with the catalog policies and
-- catalog_guard. Everything personal (status, rating, review, favorite, progress, dates, times
-- consumed) moves to the new owned table user_media, one row per (user, catalog item).
--
-- Existing rows: created_by and the user_media rows go to the one live admin. user_media.id is the
-- old media_items.id, so /media/:id URLs keep working. Live items that share a (source,
-- external_id) would break the new "one catalog row per external ID" unique index: all but the
-- oldest lose their source/external_id (they stay as hand-entered rows, nothing is deleted).
-- ROLLOUT.md has the pre-check that lists them.

-- migrate:up
ALTER TABLE public.media_items ADD COLUMN created_by bigint;

DO $$
DECLARE
  admins bigint;
  owner_id bigint;
BEGIN
  IF EXISTS (SELECT 1 FROM public.media_items) THEN
    SELECT count(*), min(id) INTO admins, owner_id FROM public.users
      WHERE is_admin AND deleted_at IS NULL;
    IF admins <> 1 THEN
      RAISE EXCEPTION 'media_items: % live admin accounts, expected exactly one to own the existing rows (docs/multi-user/DECISIONS.md, B3)', admins;
    END IF;
    UPDATE public.media_items SET created_by = owner_id WHERE created_by IS NULL;
  END IF;
END $$;

ALTER TABLE public.media_items ALTER COLUMN created_by SET NOT NULL;
ALTER TABLE public.media_items
  ADD CONSTRAINT fk_media_items_created_by FOREIGN KEY (created_by) REFERENCES public.users (id);
CREATE INDEX idx_media_items_created_by ON public.media_items USING btree (created_by);

-- Normalise "no source" to NULL-free empty strings first, then free duplicates.
UPDATE public.media_items SET source = '' WHERE source IS NULL;
UPDATE public.media_items SET external_id = '' WHERE external_id IS NULL;
UPDATE public.media_items m SET source = '', external_id = ''
  WHERE m.deleted_at IS NULL AND m.source <> '' AND m.external_id <> ''
    AND m.id <> (SELECT min(o.id) FROM public.media_items o
                 WHERE o.deleted_at IS NULL AND o.source = m.source AND o.external_id = m.external_id);
CREATE UNIQUE INDEX uq_media_items_source_external_id ON public.media_items USING btree (source, external_id)
  WHERE deleted_at IS NULL AND source <> '' AND external_id <> '';

CREATE TABLE public.user_media (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    media_item_id bigint NOT NULL,
    status character varying(16) NOT NULL,
    rating numeric(2,1),
    review text,
    favorite boolean DEFAULT false NOT NULL,
    progress bigint,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    times_consumed bigint DEFAULT 0 NOT NULL,
    CONSTRAINT fk_user_media_user FOREIGN KEY (user_id) REFERENCES public.users (id),
    CONSTRAINT fk_user_media_item FOREIGN KEY (media_item_id) REFERENCES public.media_items (id)
);

-- Existing per-item fields -> the admin's user_media rows (ids carried over).
INSERT INTO public.user_media (id, created_at, updated_at, deleted_at, user_id, media_item_id, status,
                               rating, review, favorite, progress, started_at, finished_at, times_consumed)
  SELECT id, created_at, updated_at, deleted_at, created_by, id, status,
         rating, review, favorite, progress, started_at, finished_at, times_consumed
  FROM public.media_items;
SELECT setval(pg_get_serial_sequence('public.user_media', 'id'),
              COALESCE((SELECT max(id) FROM public.user_media), 1),
              EXISTS (SELECT 1 FROM public.user_media));

CREATE INDEX idx_user_media_deleted_at ON public.user_media USING btree (deleted_at);
CREATE INDEX idx_user_media_user_id ON public.user_media USING btree (user_id);
CREATE INDEX idx_user_media_media_item_id ON public.user_media USING btree (media_item_id);
CREATE INDEX idx_user_media_status ON public.user_media USING btree (user_id, status);
-- One save per user per catalog item. Soft-deleted saves are excluded so a removed item can be saved again.
CREATE UNIQUE INDEX uq_user_media_user_item ON public.user_media USING btree (user_id, media_item_id)
  WHERE deleted_at IS NULL;

ALTER TABLE public.user_media ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON public.user_media
  USING (user_id = public.app_user_id())
  WITH CHECK (user_id = public.app_user_id());

DROP INDEX public.idx_media_items_status;
ALTER TABLE public.media_items
  DROP COLUMN status, DROP COLUMN rating, DROP COLUMN review, DROP COLUMN favorite,
  DROP COLUMN progress, DROP COLUMN started_at, DROP COLUMN finished_at, DROP COLUMN times_consumed;

ALTER TABLE public.media_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY catalog_read ON public.media_items FOR SELECT
  USING (public.app_user_id() IS NOT NULL);
CREATE POLICY catalog_add ON public.media_items FOR INSERT
  WITH CHECK (created_by = public.app_user_id());
CREATE POLICY catalog_edit ON public.media_items FOR UPDATE
  USING (created_by = public.app_user_id() OR public.app_is_admin())
  WITH CHECK (created_by = public.app_user_id() OR public.app_is_admin());
CREATE POLICY catalog_delete ON public.media_items FOR DELETE
  USING (public.app_is_admin());
CREATE TRIGGER catalog_guard BEFORE UPDATE ON public.media_items
  FOR EACH ROW EXECUTE FUNCTION public.catalog_guard();

-- migrate:down
DROP TRIGGER catalog_guard ON public.media_items;
DROP POLICY catalog_delete ON public.media_items;
DROP POLICY catalog_edit ON public.media_items;
DROP POLICY catalog_add ON public.media_items;
DROP POLICY catalog_read ON public.media_items;
ALTER TABLE public.media_items DISABLE ROW LEVEL SECURITY;

ALTER TABLE public.media_items
  ADD COLUMN status character varying(16),
  ADD COLUMN rating numeric(2,1),
  ADD COLUMN review text,
  ADD COLUMN favorite boolean DEFAULT false NOT NULL,
  ADD COLUMN progress bigint,
  ADD COLUMN started_at timestamp with time zone,
  ADD COLUMN finished_at timestamp with time zone,
  ADD COLUMN times_consumed bigint DEFAULT 0 NOT NULL;
-- Lossy by nature: with several users, an item takes the personal fields of the lowest user_id.
UPDATE public.media_items m SET status = u.status, rating = u.rating, review = u.review,
    favorite = u.favorite, progress = u.progress, started_at = u.started_at,
    finished_at = u.finished_at, times_consumed = u.times_consumed
  FROM (SELECT DISTINCT ON (media_item_id) * FROM public.user_media ORDER BY media_item_id, user_id, id) u
  WHERE u.media_item_id = m.id;
UPDATE public.media_items SET status = 'planned' WHERE status IS NULL;
ALTER TABLE public.media_items ALTER COLUMN status SET NOT NULL;
CREATE INDEX idx_media_items_status ON public.media_items USING btree (status);

DROP TABLE public.user_media;
DROP INDEX public.uq_media_items_source_external_id;
DROP INDEX public.idx_media_items_created_by;
ALTER TABLE public.media_items DROP CONSTRAINT fk_media_items_created_by;
ALTER TABLE public.media_items DROP COLUMN created_by;
