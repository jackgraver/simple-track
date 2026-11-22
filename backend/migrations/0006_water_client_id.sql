-- Water log writes safe to retry for optimistic UI (optimistic-ui T1).
-- client_id (client-generated uuid) makes POST /tracking/water idempotent.
-- NULL for rows created before this migration or by clients that do not send one.

-- migrate:up
ALTER TABLE public.water_logs ADD COLUMN client_id uuid;
CREATE UNIQUE INDEX uq_water_logs_client_id ON public.water_logs USING btree (client_id) WHERE client_id IS NOT NULL;

-- migrate:down
DROP INDEX public.uq_water_logs_client_id;
ALTER TABLE public.water_logs DROP COLUMN client_id;
