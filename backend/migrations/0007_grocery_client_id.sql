-- Grocery add idempotency for optimistic UI (optimistic-ui T2).
-- client_id (client-generated uuid) makes POST /tracking/grocery/items safe to retry.
-- NULL for rows created before this migration or by clients that do not send one.

-- migrate:up
ALTER TABLE public.grocery_items ADD COLUMN client_id uuid;
CREATE UNIQUE INDEX uq_grocery_items_client_id ON public.grocery_items USING btree (client_id) WHERE client_id IS NOT NULL;

-- migrate:down
DROP INDEX public.uq_grocery_items_client_id;
ALTER TABLE public.grocery_items DROP COLUMN client_id;
