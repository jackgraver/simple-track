-- Tables shaped like today's single-user schema, with some rows, for TestRLSPatterns to convert
-- with rls_patterns.sql. Runs as the schema owner in a migrated database that has an admin user.

CREATE TABLE public.example_lists (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    client_id uuid
);
CREATE UNIQUE INDEX uq_example_lists_name ON public.example_lists USING btree (name);
CREATE UNIQUE INDEX uq_example_lists_client_id ON public.example_lists USING btree (client_id) WHERE client_id IS NOT NULL;

CREATE TABLE public.example_list_items (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    list_id bigint NOT NULL,
    text text NOT NULL,
    CONSTRAINT fk_example_lists_items FOREIGN KEY (list_id) REFERENCES public.example_lists (id) ON DELETE CASCADE
);

CREATE TABLE public.example_foods (
    id bigserial PRIMARY KEY,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL
);

INSERT INTO public.example_lists (name) VALUES ('groceries'), ('packing');
INSERT INTO public.example_list_items (list_id, text) VALUES (1, 'milk'), (1, 'eggs'), (2, 'socks');
INSERT INTO public.example_foods (name) VALUES ('apple');
