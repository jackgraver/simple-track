-- Baseline: the schema as GORM AutoMigrate left it (2026-11-16), generated with
-- pg_dump --schema-only from a fresh database. Databases that already had this
-- schema adopt it with `migrate baseline 1` instead of running it.

-- migrate:up
CREATE TABLE public.account_balances (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    date timestamp with time zone NOT NULL,
    balance numeric(14,2) NOT NULL
);

CREATE SEQUENCE public.account_balances_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.account_balances_id_seq OWNED BY public.account_balances.id;

CREATE TABLE public.body_weight_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    date timestamp with time zone NOT NULL,
    weight_lbs numeric NOT NULL
);

CREATE SEQUENCE public.body_weight_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.body_weight_logs_id_seq OWNED BY public.body_weight_logs.id;

CREATE TABLE public.cardios (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    workout_log_id bigint NOT NULL,
    minutes bigint,
    type text,
    notes text
);

CREATE SEQUENCE public.cardios_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.cardios_id_seq OWNED BY public.cardios.id;

CREATE TABLE public.composite_food_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    composite_food_id bigint NOT NULL,
    food_id bigint NOT NULL,
    amount numeric
);

CREATE SEQUENCE public.composite_food_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.composite_food_items_id_seq OWNED BY public.composite_food_items.id;

CREATE TABLE public.composite_foods (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL
);

CREATE SEQUENCE public.composite_foods_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.composite_foods_id_seq OWNED BY public.composite_foods.id;

CREATE TABLE public.day_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    day_id bigint NOT NULL,
    meal_id bigint NOT NULL
);

CREATE SEQUENCE public.day_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.day_logs_id_seq OWNED BY public.day_logs.id;

CREATE TABLE public.days (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    date timestamp with time zone NOT NULL,
    plan_id bigint
);

CREATE SEQUENCE public.days_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.days_id_seq OWNED BY public.days.id;

CREATE TABLE public.drink_size_presets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    amount_oz numeric NOT NULL
);

CREATE SEQUENCE public.drink_size_presets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.drink_size_presets_id_seq OWNED BY public.drink_size_presets.id;

CREATE TABLE public.exercises (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    rep_rollover bigint,
    cues text,
    load_type text DEFAULT 'plate_loaded_with_bar'::text NOT NULL
);

CREATE SEQUENCE public.exercises_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.exercises_id_seq OWNED BY public.exercises.id;

CREATE TABLE public.foods (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    serving_type text NOT NULL,
    serving_amount numeric NOT NULL,
    calories numeric NOT NULL,
    protein numeric,
    fiber numeric,
    carbs numeric,
    fat numeric,
    variant_group_id bigint,
    quick_entry boolean DEFAULT false
);

CREATE SEQUENCE public.foods_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.foods_id_seq OWNED BY public.foods.id;

CREATE TABLE public.grocery_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    completed_at timestamp with time zone
);

CREATE SEQUENCE public.grocery_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.grocery_items_id_seq OWNED BY public.grocery_items.id;

CREATE TABLE public.investment_account_types (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    contribution_start_year bigint,
    eligibility_age bigint,
    carry_forward_cap numeric(14,2),
    lifetime_limit numeric(14,2),
    withdrawals_restore_room boolean DEFAULT false NOT NULL,
    opening_room_year bigint,
    opening_room numeric(14,2)
);

CREATE SEQUENCE public.investment_account_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.investment_account_types_id_seq OWNED BY public.investment_account_types.id;

CREATE TABLE public.investment_accounts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    kind character varying(32) DEFAULT 'investment'::character varying NOT NULL,
    institution text DEFAULT ''::text NOT NULL,
    notes text DEFAULT ''::text NOT NULL,
    investment_account_type_id bigint,
    current_balance numeric(14,2) DEFAULT 0 NOT NULL,
    archived_at timestamp with time zone
);

CREATE SEQUENCE public.investment_accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.investment_accounts_id_seq OWNED BY public.investment_accounts.id;

CREATE TABLE public.investment_contribution_rules (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    investment_account_type_id bigint NOT NULL,
    year bigint NOT NULL,
    annual_limit numeric(14,2) NOT NULL
);

CREATE SEQUENCE public.investment_contribution_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.investment_contribution_rules_id_seq OWNED BY public.investment_contribution_rules.id;

CREATE TABLE public.investment_deposits (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    kind character varying(16) DEFAULT 'deposit'::character varying NOT NULL,
    amount numeric(14,2) NOT NULL,
    date timestamp with time zone NOT NULL,
    note text DEFAULT ''::text NOT NULL
);

CREATE SEQUENCE public.investment_deposits_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.investment_deposits_id_seq OWNED BY public.investment_deposits.id;

CREATE TABLE public.logged_exercises (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    workout_log_id bigint,
    exercise_id bigint,
    notes text
);

CREATE SEQUENCE public.logged_exercises_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.logged_exercises_id_seq OWNED BY public.logged_exercises.id;

CREATE TABLE public.logged_sets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    logged_exercise_id bigint,
    reps bigint,
    weight numeric,
    weight_setup text
);

CREATE SEQUENCE public.logged_sets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.logged_sets_id_seq OWNED BY public.logged_sets.id;

CREATE TABLE public.meal_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    meal_id bigint NOT NULL,
    food_id bigint NOT NULL,
    amount numeric,
    group_id text,
    group_label text,
    composite_food_id bigint
);

CREATE SEQUENCE public.meal_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.meal_items_id_seq OWNED BY public.meal_items.id;

CREATE TABLE public.meals (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL
);

CREATE SEQUENCE public.meals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.meals_id_seq OWNED BY public.meals.id;

CREATE TABLE public.media_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    type character varying(16) NOT NULL,
    title text NOT NULL,
    creator text,
    release_year bigint,
    genres text,
    cover_url text,
    synopsis text,
    status character varying(16) NOT NULL,
    rating numeric(2,1),
    review text,
    favorite boolean DEFAULT false NOT NULL,
    progress bigint,
    length bigint,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    times_consumed bigint DEFAULT 0 NOT NULL,
    source character varying(32),
    external_id character varying(64)
);

CREATE SEQUENCE public.media_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.media_items_id_seq OWNED BY public.media_items.id;

CREATE TABLE public.planned_meals (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    day_id bigint NOT NULL,
    meal_id bigint NOT NULL,
    saved_meal_id bigint,
    logged boolean,
    display_order bigint DEFAULT 0 NOT NULL
);

CREATE SEQUENCE public.planned_meals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.planned_meals_id_seq OWNED BY public.planned_meals.id;

CREATE TABLE public.plans (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    calories numeric,
    protein numeric,
    fiber numeric,
    carbs numeric,
    fat numeric,
    effective_from timestamp with time zone
);

CREATE SEQUENCE public.plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.plans_id_seq OWNED BY public.plans.id;

CREATE TABLE public.saved_meal_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    saved_meal_id bigint NOT NULL,
    food_id bigint NOT NULL,
    amount numeric,
    group_id text,
    group_label text,
    composite_food_id bigint
);

CREATE SEQUENCE public.saved_meal_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.saved_meal_items_id_seq OWNED BY public.saved_meal_items.id;

CREATE TABLE public.saved_meals (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL
);

CREATE SEQUENCE public.saved_meals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.saved_meals_id_seq OWNED BY public.saved_meals.id;

CREATE TABLE public.step_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    date timestamp with time zone NOT NULL,
    steps bigint NOT NULL
);

CREATE SEQUENCE public.step_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.step_logs_id_seq OWNED BY public.step_logs.id;

CREATE TABLE public.user_profiles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    height_in numeric NOT NULL,
    age bigint NOT NULL,
    sex text NOT NULL,
    activity_level text NOT NULL
);

CREATE SEQUENCE public.user_profiles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.user_profiles_id_seq OWNED BY public.user_profiles.id;

CREATE TABLE public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    username text NOT NULL,
    password text NOT NULL,
    email text,
    birth_year bigint
);

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

CREATE TABLE public.water_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    date timestamp with time zone NOT NULL,
    amount_oz numeric NOT NULL,
    preset_id bigint
);

CREATE SEQUENCE public.water_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.water_logs_id_seq OWNED BY public.water_logs.id;

CREATE TABLE public.workout_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    date timestamp with time zone,
    workout_plan_id bigint,
    pre_mobility_checked jsonb,
    post_mobility_checked jsonb
);

CREATE SEQUENCE public.workout_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.workout_logs_id_seq OWNED BY public.workout_logs.id;

CREATE TABLE public.workout_plan_days (
    workout_plan_id bigint NOT NULL,
    day_of_week bigint NOT NULL,
    created_at timestamp with time zone
);

CREATE TABLE public.workout_plan_exercises (
    workout_plan_id bigint NOT NULL,
    exercise_id bigint NOT NULL,
    display_order bigint DEFAULT 0 NOT NULL
);

CREATE TABLE public.workout_plans (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    workout_program_id bigint,
    day_of_week bigint,
    planned_cardio_type text,
    planned_cardio_minutes bigint,
    pre_mobility_items jsonb,
    post_mobility_items jsonb
);

CREATE SEQUENCE public.workout_plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.workout_plans_id_seq OWNED BY public.workout_plans.id;

CREATE TABLE public.workout_programs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    is_active boolean
);

CREATE SEQUENCE public.workout_programs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.workout_programs_id_seq OWNED BY public.workout_programs.id;

ALTER TABLE ONLY public.account_balances ALTER COLUMN id SET DEFAULT nextval('public.account_balances_id_seq'::regclass);

ALTER TABLE ONLY public.body_weight_logs ALTER COLUMN id SET DEFAULT nextval('public.body_weight_logs_id_seq'::regclass);

ALTER TABLE ONLY public.cardios ALTER COLUMN id SET DEFAULT nextval('public.cardios_id_seq'::regclass);

ALTER TABLE ONLY public.composite_food_items ALTER COLUMN id SET DEFAULT nextval('public.composite_food_items_id_seq'::regclass);

ALTER TABLE ONLY public.composite_foods ALTER COLUMN id SET DEFAULT nextval('public.composite_foods_id_seq'::regclass);

ALTER TABLE ONLY public.day_logs ALTER COLUMN id SET DEFAULT nextval('public.day_logs_id_seq'::regclass);

ALTER TABLE ONLY public.days ALTER COLUMN id SET DEFAULT nextval('public.days_id_seq'::regclass);

ALTER TABLE ONLY public.drink_size_presets ALTER COLUMN id SET DEFAULT nextval('public.drink_size_presets_id_seq'::regclass);

ALTER TABLE ONLY public.exercises ALTER COLUMN id SET DEFAULT nextval('public.exercises_id_seq'::regclass);

ALTER TABLE ONLY public.foods ALTER COLUMN id SET DEFAULT nextval('public.foods_id_seq'::regclass);

ALTER TABLE ONLY public.grocery_items ALTER COLUMN id SET DEFAULT nextval('public.grocery_items_id_seq'::regclass);

ALTER TABLE ONLY public.investment_account_types ALTER COLUMN id SET DEFAULT nextval('public.investment_account_types_id_seq'::regclass);

ALTER TABLE ONLY public.investment_accounts ALTER COLUMN id SET DEFAULT nextval('public.investment_accounts_id_seq'::regclass);

ALTER TABLE ONLY public.investment_contribution_rules ALTER COLUMN id SET DEFAULT nextval('public.investment_contribution_rules_id_seq'::regclass);

ALTER TABLE ONLY public.investment_deposits ALTER COLUMN id SET DEFAULT nextval('public.investment_deposits_id_seq'::regclass);

ALTER TABLE ONLY public.logged_exercises ALTER COLUMN id SET DEFAULT nextval('public.logged_exercises_id_seq'::regclass);

ALTER TABLE ONLY public.logged_sets ALTER COLUMN id SET DEFAULT nextval('public.logged_sets_id_seq'::regclass);

ALTER TABLE ONLY public.meal_items ALTER COLUMN id SET DEFAULT nextval('public.meal_items_id_seq'::regclass);

ALTER TABLE ONLY public.meals ALTER COLUMN id SET DEFAULT nextval('public.meals_id_seq'::regclass);

ALTER TABLE ONLY public.media_items ALTER COLUMN id SET DEFAULT nextval('public.media_items_id_seq'::regclass);

ALTER TABLE ONLY public.planned_meals ALTER COLUMN id SET DEFAULT nextval('public.planned_meals_id_seq'::regclass);

ALTER TABLE ONLY public.plans ALTER COLUMN id SET DEFAULT nextval('public.plans_id_seq'::regclass);

ALTER TABLE ONLY public.saved_meal_items ALTER COLUMN id SET DEFAULT nextval('public.saved_meal_items_id_seq'::regclass);

ALTER TABLE ONLY public.saved_meals ALTER COLUMN id SET DEFAULT nextval('public.saved_meals_id_seq'::regclass);

ALTER TABLE ONLY public.step_logs ALTER COLUMN id SET DEFAULT nextval('public.step_logs_id_seq'::regclass);

ALTER TABLE ONLY public.user_profiles ALTER COLUMN id SET DEFAULT nextval('public.user_profiles_id_seq'::regclass);

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

ALTER TABLE ONLY public.water_logs ALTER COLUMN id SET DEFAULT nextval('public.water_logs_id_seq'::regclass);

ALTER TABLE ONLY public.workout_logs ALTER COLUMN id SET DEFAULT nextval('public.workout_logs_id_seq'::regclass);

ALTER TABLE ONLY public.workout_plans ALTER COLUMN id SET DEFAULT nextval('public.workout_plans_id_seq'::regclass);

ALTER TABLE ONLY public.workout_programs ALTER COLUMN id SET DEFAULT nextval('public.workout_programs_id_seq'::regclass);

ALTER TABLE ONLY public.account_balances
    ADD CONSTRAINT account_balances_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.body_weight_logs
    ADD CONSTRAINT body_weight_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.cardios
    ADD CONSTRAINT cardios_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.composite_food_items
    ADD CONSTRAINT composite_food_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.composite_foods
    ADD CONSTRAINT composite_foods_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.day_logs
    ADD CONSTRAINT day_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.days
    ADD CONSTRAINT days_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.drink_size_presets
    ADD CONSTRAINT drink_size_presets_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.exercises
    ADD CONSTRAINT exercises_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.foods
    ADD CONSTRAINT foods_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.grocery_items
    ADD CONSTRAINT grocery_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.investment_account_types
    ADD CONSTRAINT investment_account_types_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.investment_accounts
    ADD CONSTRAINT investment_accounts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.investment_contribution_rules
    ADD CONSTRAINT investment_contribution_rules_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.investment_deposits
    ADD CONSTRAINT investment_deposits_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.logged_exercises
    ADD CONSTRAINT logged_exercises_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.logged_sets
    ADD CONSTRAINT logged_sets_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.meal_items
    ADD CONSTRAINT meal_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.meals
    ADD CONSTRAINT meals_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.media_items
    ADD CONSTRAINT media_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.planned_meals
    ADD CONSTRAINT planned_meals_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.plans
    ADD CONSTRAINT plans_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.saved_meal_items
    ADD CONSTRAINT saved_meal_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.saved_meals
    ADD CONSTRAINT saved_meals_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.step_logs
    ADD CONSTRAINT step_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.water_logs
    ADD CONSTRAINT water_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workout_logs
    ADD CONSTRAINT workout_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workout_plan_days
    ADD CONSTRAINT workout_plan_days_pkey PRIMARY KEY (workout_plan_id, day_of_week);

ALTER TABLE ONLY public.workout_plan_exercises
    ADD CONSTRAINT workout_plan_exercises_pkey PRIMARY KEY (workout_plan_id, exercise_id);

ALTER TABLE ONLY public.workout_plans
    ADD CONSTRAINT workout_plans_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.workout_programs
    ADD CONSTRAINT workout_programs_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX idx_account_balance_day ON public.account_balances USING btree (account_id, date);

CREATE INDEX idx_account_balances_deleted_at ON public.account_balances USING btree (deleted_at);

CREATE UNIQUE INDEX idx_body_weight_logs_date ON public.body_weight_logs USING btree (date);

CREATE INDEX idx_body_weight_logs_deleted_at ON public.body_weight_logs USING btree (deleted_at);

CREATE INDEX idx_cardios_deleted_at ON public.cardios USING btree (deleted_at);

CREATE UNIQUE INDEX idx_cardios_workout_log_id ON public.cardios USING btree (workout_log_id);

CREATE INDEX idx_composite_food_items_composite_food_id ON public.composite_food_items USING btree (composite_food_id);

CREATE INDEX idx_composite_food_items_deleted_at ON public.composite_food_items USING btree (deleted_at);

CREATE INDEX idx_composite_food_items_food_id ON public.composite_food_items USING btree (food_id);

CREATE INDEX idx_composite_foods_deleted_at ON public.composite_foods USING btree (deleted_at);

CREATE UNIQUE INDEX idx_composite_foods_name ON public.composite_foods USING btree (name);

CREATE INDEX idx_day_logs_deleted_at ON public.day_logs USING btree (deleted_at);

CREATE UNIQUE INDEX idx_days_date ON public.days USING btree (date);

CREATE INDEX idx_days_deleted_at ON public.days USING btree (deleted_at);

CREATE INDEX idx_drink_size_presets_deleted_at ON public.drink_size_presets USING btree (deleted_at);

CREATE INDEX idx_exercises_deleted_at ON public.exercises USING btree (deleted_at);

CREATE UNIQUE INDEX idx_exercises_name ON public.exercises USING btree (name);

CREATE INDEX idx_foods_deleted_at ON public.foods USING btree (deleted_at);

CREATE UNIQUE INDEX idx_foods_name ON public.foods USING btree (name);

CREATE INDEX idx_foods_quick_entry ON public.foods USING btree (quick_entry);

CREATE INDEX idx_foods_variant_group_id ON public.foods USING btree (variant_group_id);

CREATE INDEX idx_grocery_items_deleted_at ON public.grocery_items USING btree (deleted_at);

CREATE INDEX idx_grocery_items_name ON public.grocery_items USING btree (name);

CREATE UNIQUE INDEX idx_investment_account_type_contribution_rule ON public.investment_contribution_rules USING btree (investment_account_type_id, year);

CREATE INDEX idx_investment_account_types_deleted_at ON public.investment_account_types USING btree (deleted_at);

CREATE UNIQUE INDEX idx_investment_account_types_name ON public.investment_account_types USING btree (name);

CREATE INDEX idx_investment_accounts_deleted_at ON public.investment_accounts USING btree (deleted_at);

CREATE INDEX idx_investment_contribution_rules_deleted_at ON public.investment_contribution_rules USING btree (deleted_at);

CREATE INDEX idx_investment_deposits_account_id ON public.investment_deposits USING btree (account_id);

CREATE INDEX idx_investment_deposits_date ON public.investment_deposits USING btree (date);

CREATE INDEX idx_investment_deposits_deleted_at ON public.investment_deposits USING btree (deleted_at);

CREATE INDEX idx_logged_exercises_deleted_at ON public.logged_exercises USING btree (deleted_at);

CREATE INDEX idx_logged_sets_deleted_at ON public.logged_sets USING btree (deleted_at);

CREATE INDEX idx_meal_items_deleted_at ON public.meal_items USING btree (deleted_at);

CREATE INDEX idx_meal_items_food_id ON public.meal_items USING btree (food_id);

CREATE INDEX idx_meal_items_group_id ON public.meal_items USING btree (group_id);

CREATE INDEX idx_meal_items_meal_id ON public.meal_items USING btree (meal_id);

CREATE INDEX idx_meals_deleted_at ON public.meals USING btree (deleted_at);

CREATE INDEX idx_media_items_deleted_at ON public.media_items USING btree (deleted_at);

CREATE INDEX idx_media_items_status ON public.media_items USING btree (status);

CREATE INDEX idx_media_items_title ON public.media_items USING btree (title);

CREATE INDEX idx_media_items_type ON public.media_items USING btree (type);

CREATE INDEX idx_planned_meals_deleted_at ON public.planned_meals USING btree (deleted_at);

CREATE INDEX idx_planned_meals_saved_meal_id ON public.planned_meals USING btree (saved_meal_id);

CREATE INDEX idx_plans_deleted_at ON public.plans USING btree (deleted_at);

CREATE INDEX idx_plans_effective_from ON public.plans USING btree (effective_from);

CREATE INDEX idx_saved_meal_items_deleted_at ON public.saved_meal_items USING btree (deleted_at);

CREATE INDEX idx_saved_meal_items_food_id ON public.saved_meal_items USING btree (food_id);

CREATE INDEX idx_saved_meal_items_group_id ON public.saved_meal_items USING btree (group_id);

CREATE INDEX idx_saved_meal_items_saved_meal_id ON public.saved_meal_items USING btree (saved_meal_id);

CREATE INDEX idx_saved_meals_deleted_at ON public.saved_meals USING btree (deleted_at);

CREATE UNIQUE INDEX idx_step_logs_date ON public.step_logs USING btree (date);

CREATE INDEX idx_step_logs_deleted_at ON public.step_logs USING btree (deleted_at);

CREATE INDEX idx_user_profiles_deleted_at ON public.user_profiles USING btree (deleted_at);

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);

CREATE UNIQUE INDEX idx_users_username ON public.users USING btree (username);

CREATE INDEX idx_water_logs_date ON public.water_logs USING btree (date);

CREATE INDEX idx_water_logs_deleted_at ON public.water_logs USING btree (deleted_at);

CREATE INDEX idx_workout_logs_deleted_at ON public.workout_logs USING btree (deleted_at);

CREATE INDEX idx_workout_plans_deleted_at ON public.workout_plans USING btree (deleted_at);

CREATE INDEX idx_workout_programs_deleted_at ON public.workout_programs USING btree (deleted_at);

ALTER TABLE ONLY public.composite_food_items
    ADD CONSTRAINT fk_composite_food_items_food FOREIGN KEY (food_id) REFERENCES public.foods(id);

ALTER TABLE ONLY public.composite_food_items
    ADD CONSTRAINT fk_composite_foods_items FOREIGN KEY (composite_food_id) REFERENCES public.composite_foods(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.day_logs
    ADD CONSTRAINT fk_day_logs_meal FOREIGN KEY (meal_id) REFERENCES public.meals(id);

ALTER TABLE ONLY public.day_logs
    ADD CONSTRAINT fk_days_logs FOREIGN KEY (day_id) REFERENCES public.days(id);

ALTER TABLE ONLY public.days
    ADD CONSTRAINT fk_days_plan FOREIGN KEY (plan_id) REFERENCES public.plans(id);

ALTER TABLE ONLY public.planned_meals
    ADD CONSTRAINT fk_days_planned_meals FOREIGN KEY (day_id) REFERENCES public.days(id);

ALTER TABLE ONLY public.investment_contribution_rules
    ADD CONSTRAINT fk_investment_account_types_limits FOREIGN KEY (investment_account_type_id) REFERENCES public.investment_account_types(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.account_balances
    ADD CONSTRAINT fk_investment_accounts_balances FOREIGN KEY (account_id) REFERENCES public.investment_accounts(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.investment_accounts
    ADD CONSTRAINT fk_investment_accounts_contribution_type FOREIGN KEY (investment_account_type_id) REFERENCES public.investment_account_types(id);

ALTER TABLE ONLY public.investment_deposits
    ADD CONSTRAINT fk_investment_accounts_transactions FOREIGN KEY (account_id) REFERENCES public.investment_accounts(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.logged_exercises
    ADD CONSTRAINT fk_logged_exercises_exercise FOREIGN KEY (exercise_id) REFERENCES public.exercises(id);

ALTER TABLE ONLY public.logged_sets
    ADD CONSTRAINT fk_logged_exercises_sets FOREIGN KEY (logged_exercise_id) REFERENCES public.logged_exercises(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.meal_items
    ADD CONSTRAINT fk_meal_items_food FOREIGN KEY (food_id) REFERENCES public.foods(id);

ALTER TABLE ONLY public.meal_items
    ADD CONSTRAINT fk_meals_items FOREIGN KEY (meal_id) REFERENCES public.meals(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.planned_meals
    ADD CONSTRAINT fk_planned_meals_meal FOREIGN KEY (meal_id) REFERENCES public.meals(id);

ALTER TABLE ONLY public.saved_meal_items
    ADD CONSTRAINT fk_saved_meal_items_food FOREIGN KEY (food_id) REFERENCES public.foods(id);

ALTER TABLE ONLY public.saved_meal_items
    ADD CONSTRAINT fk_saved_meals_items FOREIGN KEY (saved_meal_id) REFERENCES public.saved_meals(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.water_logs
    ADD CONSTRAINT fk_water_logs_preset FOREIGN KEY (preset_id) REFERENCES public.drink_size_presets(id);

ALTER TABLE ONLY public.cardios
    ADD CONSTRAINT fk_workout_logs_cardio FOREIGN KEY (workout_log_id) REFERENCES public.workout_logs(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.logged_exercises
    ADD CONSTRAINT fk_workout_logs_exercises FOREIGN KEY (workout_log_id) REFERENCES public.workout_logs(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.workout_logs
    ADD CONSTRAINT fk_workout_logs_workout_plan FOREIGN KEY (workout_plan_id) REFERENCES public.workout_plans(id);

ALTER TABLE ONLY public.workout_plans
    ADD CONSTRAINT fk_workout_programs_plans FOREIGN KEY (workout_program_id) REFERENCES public.workout_programs(id);

-- The app expects exactly one active workout program.
INSERT INTO public.workout_programs (name, is_active, created_at, updated_at)
VALUES ('Default', true, now(), now());

-- migrate:down irreversible
-- Rolling back the baseline would drop every table. Recreate the database instead.
