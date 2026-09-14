-- Invalidates previously issued access tokens after password reset.
ALTER TABLE users ADD COLUMN auth_version BIGINT NOT NULL DEFAULT 0;
