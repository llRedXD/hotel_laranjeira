ALTER TABLE reservations
    ADD COLUMN cancelled_at TIMESTAMPTZ,
    ADD COLUMN cancelled_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1,
    ADD CONSTRAINT reservations_version_check CHECK (version > 0),
    ADD CONSTRAINT reservations_cancellation_check CHECK (
        (cancelled_at IS NULL AND cancelled_by IS NULL) OR
        (cancelled_at IS NOT NULL AND cancelled_by IS NOT NULL)
    );
