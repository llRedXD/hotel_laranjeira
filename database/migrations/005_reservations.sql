ALTER TABLE guests
    ADD COLUMN document_type VARCHAR(3),
    ADD COLUMN document_normalized VARCHAR(30);

ALTER TABLE guests
    ADD CONSTRAINT guests_document_type_check
    CHECK (document_type IS NULL OR document_type IN ('RG', 'CPF'));

UPDATE guests
SET document_normalized = CASE
    WHEN document IS NULL THEN NULL
    ELSE regexp_replace(upper(trim(document)), '[^[:alnum:]]', '', 'g')
END;

DO $$
BEGIN
    IF EXISTS (
        SELECT document_normalized
        FROM guests
        WHERE document_normalized IS NOT NULL
        GROUP BY document_normalized
        HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION 'Migração 005 abortada: existem documentos duplicados após normalização.';
    END IF;
END;
$$;

DROP INDEX guests_document_unique_idx;
CREATE UNIQUE INDEX guests_document_normalized_unique_idx
    ON guests (document_normalized)
    WHERE document_normalized IS NOT NULL;

CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guest_id UUID NOT NULL REFERENCES guests(id) ON DELETE RESTRICT,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE RESTRICT,
    check_in DATE NOT NULL,
    check_out DATE NOT NULL,
    occupants INTEGER NOT NULL,
    daily_rate NUMERIC(12, 2) NOT NULL,
    total_amount NUMERIC(12, 2) NOT NULL,
    guest_name_snapshot VARCHAR(120) NOT NULL,
    document_type_snapshot VARCHAR(3) NOT NULL,
    document_snapshot VARCHAR(30) NOT NULL,
    contact_email_snapshot VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT reservations_dates_check CHECK (check_out > check_in),
    CONSTRAINT reservations_occupants_check CHECK (occupants > 0),
    CONSTRAINT reservations_document_type_check CHECK (document_type_snapshot IN ('RG', 'CPF')),
    CONSTRAINT reservations_daily_rate_check CHECK (daily_rate >= 0),
    CONSTRAINT reservations_total_amount_check CHECK (total_amount >= 0)
);

CREATE INDEX reservations_room_period_idx ON reservations (room_id, check_in, check_out);
CREATE INDEX reservations_guest_created_idx ON reservations (guest_id, created_at DESC);
