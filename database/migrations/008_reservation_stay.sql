ALTER TABLE reservations
    ADD COLUMN checked_in_at TIMESTAMPTZ,
    ADD COLUMN checked_in_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN checked_out_at TIMESTAMPTZ,
    ADD COLUMN checked_out_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN actual_nights INTEGER,
    ADD COLUMN final_amount NUMERIC(12, 2),
    ADD CONSTRAINT reservations_checkin_pair CHECK (
        (checked_in_at IS NULL AND checked_in_by IS NULL) OR
        (checked_in_at IS NOT NULL AND checked_in_by IS NOT NULL)
    ),
    ADD CONSTRAINT reservations_checkout_pair CHECK (
        (checked_out_at IS NULL AND checked_out_by IS NULL AND actual_nights IS NULL AND final_amount IS NULL) OR
        (checked_out_at IS NOT NULL AND checked_out_by IS NOT NULL AND actual_nights IS NOT NULL AND final_amount IS NOT NULL)
    ),
    ADD CONSTRAINT reservations_checkout_requires_checkin CHECK (
        checked_out_at IS NULL OR (checked_in_at IS NOT NULL AND checked_out_at >= checked_in_at)
    ),
    ADD CONSTRAINT reservations_actual_nights_check CHECK (actual_nights IS NULL OR actual_nights >= 0),
    ADD CONSTRAINT reservations_final_amount_check CHECK (final_amount IS NULL OR final_amount >= 0);

ALTER TABLE rooms
    ADD COLUMN occupied_by_reservation_id UUID UNIQUE REFERENCES reservations(id) ON DELETE RESTRICT;

ALTER TABLE reservation_events DROP CONSTRAINT reservation_events_action_check;
ALTER TABLE reservation_events ADD CONSTRAINT reservation_events_action_check
    CHECK (action IN ('imported', 'created', 'cancelled', 'updated', 'checked_in', 'checked_out'));

CREATE OR REPLACE FUNCTION reservation_audit_snapshot(r reservations, reference_date DATE)
RETURNS JSONB LANGUAGE SQL STABLE AS $$
    SELECT jsonb_build_object(
        'status', CASE
            WHEN r.cancelled_at IS NOT NULL THEN 'cancelled'
            WHEN r.checked_out_at IS NOT NULL THEN 'completed'
            WHEN r.checked_in_at IS NOT NULL THEN 'in_house'
            WHEN r.check_out < reference_date THEN 'ended'
            WHEN r.check_out = reference_date THEN 'departure_today'
            WHEN r.check_in <= reference_date THEN 'in_progress'
            ELSE 'confirmed' END,
        'roomId', r.room_id, 'roomNumber', (SELECT number FROM rooms WHERE id = r.room_id),
        'checkIn', r.check_in, 'checkOut', r.check_out,
        'checkedInAt', r.checked_in_at, 'checkedOutAt', r.checked_out_at,
        'actualNights', r.actual_nights, 'finalAmount', r.final_amount::text,
        'occupants', r.occupants, 'dailyRate', r.daily_rate::text,
        'totalAmount', r.total_amount::text, 'version', r.version
    );
$$;

CREATE OR REPLACE FUNCTION record_reservation_event() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    event_time TIMESTAMPTZ := clock_timestamp();
    event_date DATE := (event_time AT TIME ZONE 'America/Sao_Paulo')::date;
    previous_state JSONB;
    next_state JSONB := reservation_audit_snapshot(NEW, event_date);
    event_action TEXT := 'created';
    employee UUID := nullif(current_setting('hotel.actor_id', true), '')::uuid;
    event_source TEXT := coalesce(nullif(current_setting('hotel.event_source', true), ''), 'database');
BEGIN
    IF TG_OP = 'UPDATE' THEN
        previous_state := reservation_audit_snapshot(OLD, event_date);
        IF previous_state = next_state THEN RETURN NEW; END IF;
        event_action := 'updated';
        IF OLD.cancelled_at IS NULL AND NEW.cancelled_at IS NOT NULL THEN
            event_action := 'cancelled';
            employee := NEW.cancelled_by;
            event_source := 'staff';
        ELSIF OLD.checked_in_at IS NULL AND NEW.checked_in_at IS NOT NULL THEN
            event_action := 'checked_in';
            employee := NEW.checked_in_by;
            event_source := 'staff';
        ELSIF OLD.checked_out_at IS NULL AND NEW.checked_out_at IS NOT NULL THEN
            event_action := 'checked_out';
            employee := NEW.checked_out_by;
            event_source := 'staff';
        END IF;
    END IF;
    INSERT INTO reservation_events (
        reservation_id, action, source, actor_id, actor_name, guest_name, room_number,
        occurred_at, before_state, after_state
    ) VALUES (
        NEW.id, event_action, event_source, employee,
        (SELECT name FROM users WHERE id = employee), NEW.guest_name_snapshot,
        (SELECT number FROM rooms WHERE id = NEW.room_id),
        event_time, previous_state, next_state
    );
    RETURN NEW;
END;
$$;
