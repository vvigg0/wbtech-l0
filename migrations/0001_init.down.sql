-- 0001_init.down.sql
-- Откат начальной схемы (удаление таблиц)

DROP TABLE IF EXISTS public.items      CASCADE;
DROP TABLE IF EXISTS public.payments   CASCADE;
DROP TABLE IF EXISTS public.deliveries CASCADE;
DROP TABLE IF EXISTS public.orders     CASCADE;
