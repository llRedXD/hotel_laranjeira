-- Prepara somente os dados exclusivos da apresentação do ciclo de reserva.
-- Repetível: recria estas duas reservas sem apagar dados operacionais alheios.
-- Requer migrations 001–008 e pelo menos um usuário admin/reception.
BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE role IN ('admin', 'reception')) THEN
        RAISE EXCEPTION 'Crie a conta de desenvolvimento (admin ou recepção) antes de carregar os dados de apresentação.';
    END IF;
END;
$$;

INSERT INTO room_categories (id, name, description) VALUES
    ('91000000-0000-4000-8000-000000000001', 'Demonstração do ciclo', 'Categoria reservada aos dados da apresentação.')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO rooms (id, number, category_id, capacity, daily_rate, is_available) VALUES
    ('92000000-0000-4000-8000-000000000001', 'DEMO-901', '91000000-0000-4000-8000-000000000001', 2, 350.00, TRUE),
    ('92000000-0000-4000-8000-000000000002', 'DEMO-902', '91000000-0000-4000-8000-000000000001', 2, 350.00, TRUE)
ON CONFLICT (id) DO UPDATE SET
    number = EXCLUDED.number,
    category_id = EXCLUDED.category_id,
    capacity = EXCLUDED.capacity,
    daily_rate = EXCLUDED.daily_rate,
    is_available = EXCLUDED.is_available,
    updated_at = NOW();

INSERT INTO guests (id, name, document_type, document, document_normalized, email, phone) VALUES
    ('93000000-0000-4000-8000-000000000001', 'Marina Costa (Apresentação)', 'RG', 'DEMO-RG-901', 'DEMORG901', 'marina.demo901@hotel.test', '1199000901'),
    ('93000000-0000-4000-8000-000000000002', 'Rafael Lima (Apresentação)', 'RG', 'DEMO-RG-902', 'DEMORG902', 'rafael.demo902@hotel.test', '1199000902')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    document_type = EXCLUDED.document_type,
    document = EXCLUDED.document,
    document_normalized = EXCLUDED.document_normalized,
    email = EXCLUDED.email,
    phone = EXCLUDED.phone,
    updated_at = NOW();

-- Apaga apenas eventos e reservas dos IDs fixos deste roteiro, liberando antes
-- qualquer quarto que esteja ocupado por uma execução anterior do próprio seed.
DELETE FROM reservation_events
WHERE reservation_id IN (
    '94000000-0000-4000-8000-000000000001',
    '94000000-0000-4000-8000-000000000002'
);

UPDATE rooms
SET occupied_by_reservation_id = NULL, updated_at = NOW()
WHERE occupied_by_reservation_id IN (
    '94000000-0000-4000-8000-000000000001',
    '94000000-0000-4000-8000-000000000002'
);

DELETE FROM reservations
WHERE id IN (
    '94000000-0000-4000-8000-000000000001',
    '94000000-0000-4000-8000-000000000002'
);

INSERT INTO reservations (
    id, guest_id, room_id, check_in, check_out, occupants, daily_rate, total_amount,
    guest_name_snapshot, document_type_snapshot, document_snapshot, contact_email_snapshot
) VALUES
    (
        '94000000-0000-4000-8000-000000000001',
        '93000000-0000-4000-8000-000000000001',
        '92000000-0000-4000-8000-000000000001',
        CURRENT_DATE + 1, CURRENT_DATE + 4, 2, 350.00, 1050.00,
        'Marina Costa (Apresentação)', 'RG', 'DEMO-RG-901', 'marina.demo901@hotel.test'
    ),
    (
        '94000000-0000-4000-8000-000000000002',
        '93000000-0000-4000-8000-000000000002',
        '92000000-0000-4000-8000-000000000002',
        CURRENT_DATE - 2, CURRENT_DATE + 1, 2, 350.00, 1050.00,
        'Rafael Lima (Apresentação)', 'RG', 'DEMO-RG-902', 'rafael.demo902@hotel.test'
    );

-- A segunda reserva já está em hospedagem há duas noites para demonstrar
-- checkout imediato com valor final de R$ 700,00.
SELECT set_config(
    'hotel.actor_id',
    (SELECT id::text FROM users WHERE role IN ('admin', 'reception')
     ORDER BY CASE role WHEN 'admin' THEN 0 ELSE 1 END, created_at, id LIMIT 1),
    TRUE
);
SELECT set_config('hotel.event_source', 'staff', TRUE);

UPDATE reservations
SET checked_in_at = (CURRENT_DATE - 2 + TIME '15:00') AT TIME ZONE 'America/Sao_Paulo',
    checked_in_by = (SELECT id FROM users WHERE role IN ('admin', 'reception')
                     ORDER BY CASE role WHEN 'admin' THEN 0 ELSE 1 END, created_at, id LIMIT 1),
    version = version + 1
WHERE id = '94000000-0000-4000-8000-000000000002';

UPDATE rooms
SET occupied_by_reservation_id = '94000000-0000-4000-8000-000000000002', updated_at = NOW()
WHERE id = '92000000-0000-4000-8000-000000000002';

COMMIT;
