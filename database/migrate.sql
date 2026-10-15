-- Run with psql -X -v ON_ERROR_STOP=1. One transaction and advisory lock
-- serialize migrations; errors roll back the complete batch.
BEGIN;
SELECT pg_advisory_xact_lock(72401923);
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Legacy databases were initialized by docker-entrypoint-initdb.d.
-- Explicitly opt in to adopting their existing schema; never erase volumes.
\getenv adopt_legacy ADOPT_LEGACY_SCHEMA
\if :{?adopt_legacy}
\else
\set adopt_legacy false
\endif
\if :adopt_legacy
INSERT INTO schema_migrations(version)
SELECT 1 WHERE to_regclass('public.users') IS NOT NULL
ON CONFLICT DO NOTHING;
INSERT INTO schema_migrations(version)
SELECT 2 WHERE to_regclass('public.refresh_tokens') IS NOT NULL
AND to_regclass('public.password_reset_tokens') IS NOT NULL
ON CONFLICT DO NOTHING;
INSERT INTO schema_migrations(version)
SELECT 3 WHERE to_regclass('public.guests') IS NOT NULL
AND to_regclass('public.room_categories') IS NOT NULL
AND to_regclass('public.rooms') IS NOT NULL
ON CONFLICT DO NOTHING;
INSERT INTO schema_migrations(version)
SELECT 4 WHERE EXISTS (SELECT 1 FROM information_schema.columns
WHERE table_schema='public' AND table_name='users' AND column_name='auth_version')
ON CONFLICT DO NOTHING;
\endif

SELECT NOT EXISTS(SELECT 1 FROM schema_migrations WHERE version=1) AS pending \gset
\if :pending
\ir migrations/001_create_users.sql
INSERT INTO schema_migrations(version) VALUES (1);
\endif
SELECT NOT EXISTS(SELECT 1 FROM schema_migrations WHERE version=2) AS pending \gset
\if :pending
\ir migrations/002_create_auth_tokens.sql
INSERT INTO schema_migrations(version) VALUES (2);
\endif
SELECT NOT EXISTS(SELECT 1 FROM schema_migrations WHERE version=3) AS pending \gset
\if :pending
\ir migrations/003_create_guests_and_rooms.sql
INSERT INTO schema_migrations(version) VALUES (3);
\endif
SELECT NOT EXISTS(SELECT 1 FROM schema_migrations WHERE version=4) AS pending \gset
\if :pending
\ir migrations/004_auth_version.sql
INSERT INTO schema_migrations(version) VALUES (4);
\endif
SELECT NOT EXISTS(SELECT 1 FROM schema_migrations WHERE version=5) AS pending \gset
\if :pending
\ir migrations/005_reservations.sql
INSERT INTO schema_migrations(version) VALUES (5);
\endif
COMMIT;
