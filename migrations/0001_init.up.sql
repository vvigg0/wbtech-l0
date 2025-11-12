-- 0001_init.up.sql
-- Начальная схема БД

-- ============================
--  Таблица orders (заказы)
-- ============================

CREATE TABLE public.orders (
    order_uid          text PRIMARY KEY,
    track_number       text NOT NULL UNIQUE,
    entry              text NOT NULL,
    locale             text NOT NULL,
    internal_signature text NOT NULL,
    customer_id        text NOT NULL,
    delivery_service   text NOT NULL,
    shardkey           text NOT NULL,
    sm_id              integer NOT NULL,
    date_created       timestamp with time zone NOT NULL,
    oof_shard          text NOT NULL,

    CONSTRAINT orders_order_uid_not_empty
        CHECK (length(btrim(order_uid)) > 0),
    CONSTRAINT orders_customer_id_not_empty
        CHECK (length(btrim(customer_id)) > 0),
    CONSTRAINT orders_track_number_not_empty
        CHECK (length(btrim(track_number)) > 0)
);

CREATE INDEX idx_orders_customer_id ON public.orders (customer_id);


-- ============================
--  Таблица deliveries (доставка)
-- ============================

CREATE TABLE public.deliveries (
    order_uid text PRIMARY KEY,
    name      text NOT NULL,
    phone     text,
    zip       text,
    city      text NOT NULL,
    address   text NOT NULL,
    region    text,
    email     text,

    CONSTRAINT deliveries_email_format
        CHECK (email IS NULL OR position('@' in email) > 1)
);

ALTER TABLE public.deliveries
    ADD CONSTRAINT deliveries_order_fk
        FOREIGN KEY (order_uid)
        REFERENCES public.orders(order_uid)
        ON DELETE CASCADE;


-- ============================
--  Таблица payments (оплата)
-- ============================

CREATE TABLE public.payments (
    transaction   text PRIMARY KEY,
    order_uid     text NOT NULL,
    request_id    text,
    currency      text NOT NULL,
    provider      text NOT NULL,
    amount        integer NOT NULL,
    payment_dt    bigint,
    bank          text,
    delivery_cost integer NOT NULL,
    goods_total   integer NOT NULL,
    custom_fee    integer,

    CONSTRAINT payments_amount_positive
        CHECK (amount > 0),
    CONSTRAINT payments_goods_total_positive
        CHECK (goods_total > 0),
    CONSTRAINT payments_delivery_cost_nonneg
        CHECK (delivery_cost >= 0),
    CONSTRAINT payments_custom_fee_nonneg
        CHECK (custom_fee IS NULL OR custom_fee >= 0)
);

ALTER TABLE public.payments
    ADD CONSTRAINT payments_order_fk
        FOREIGN KEY (order_uid)
        REFERENCES public.orders(order_uid)
        ON DELETE CASCADE;

CREATE INDEX idx_payments_order_uid ON public.payments (order_uid);


-- ============================
--  Таблица items (товары)
-- ============================

CREATE TABLE public.items (
    id           bigserial PRIMARY KEY,
    order_uid    text    NOT NULL,
    chrt_id      integer,
    track_number text    NOT NULL,
    price        integer NOT NULL,
    rid          text,
    name         text    NOT NULL,
    sale         integer,
    size         text,
    total_price  integer NOT NULL,
    nm_id        integer,
    brand        text,
    status       integer,

    CONSTRAINT items_price_positive
        CHECK (price > 0),
    CONSTRAINT items_total_price_positive
        CHECK (total_price > 0),
    CONSTRAINT items_sale_nonneg
        CHECK (sale IS NULL OR sale >= 0),
    CONSTRAINT items_status_nonneg
        CHECK (status IS NULL OR status >= 0)
);

ALTER TABLE public.items
    ADD CONSTRAINT items_order_fk
        FOREIGN KEY (order_uid)
        REFERENCES public.orders(order_uid)
        ON DELETE CASCADE;

CREATE INDEX idx_items_order_uid ON public.items (order_uid);
