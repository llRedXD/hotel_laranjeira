CREATE TABLE reservation_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE RESTRICT,
    action VARCHAR(20) NOT NULL CHECK (action IN ('imported', 'created', 'cancelled', 'updated')),
    source VARCHAR(20) NOT NULL CHECK (source IN ('import', 'public', 'staff', 'database')),
    actor_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    actor_name VARCHAR(120),
    guest_name VARCHAR(120) NOT NULL,
    room_number VARCHAR(20) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    before_state JSONB,
    after_state JSONB NOT NULL
);
CREATE INDEX reservation_events_reservation_time_idx ON reservation_events (reservation_id, occurred_at DESC, id);
CREATE INDEX reservation_events_time_idx ON reservation_events (occurred_at DESC, id);

CREATE FUNCTION reservation_audit_snapshot(r reservations, reference_date DATE)
RETURNS JSONB LANGUAGE SQL STABLE AS $$
    SELECT jsonb_build_object(
        'status', CASE
            WHEN r.cancelled_at IS NOT NULL THEN 'cancelled'
            WHEN r.check_out < reference_date THEN 'ended'
            WHEN r.check_out = reference_date THEN 'departure_today'
            WHEN r.check_in <= reference_date THEN 'in_progress'
            ELSE 'confirmed' END,
        'roomId', r.room_id, 'roomNumber', (SELECT number FROM rooms WHERE id = r.room_id),
        'checkIn', r.check_in, 'checkOut', r.check_out,
        'occupants', r.occupants, 'dailyRate', r.daily_rate::text,
        'totalAmount', r.total_amount::text, 'version', r.version
    );
$$;

-- This is a baseline observed during migration, not a reconstructed history.
INSERT INTO reservation_events (reservation_id, action, source, guest_name, room_number, after_state)
SELECT r.id, 'imported', 'import', r.guest_name_snapshot, rm.number,
       reservation_audit_snapshot(r, (clock_timestamp() AT TIME ZONE 'America/Sao_Paulo')::date)
FROM reservations r JOIN rooms rm ON rm.id = r.room_id;

CREATE FUNCTION record_reservation_event() RETURNS TRIGGER LANGUAGE plpgsql AS $$
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

CREATE TRIGGER reservations_audit AFTER INSERT OR UPDATE ON reservations
FOR EACH ROW EXECUTE FUNCTION record_reservation_event();
