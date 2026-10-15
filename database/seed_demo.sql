-- Reseta somente os dados operacionais do hotel e carrega uma base de demonstração.
-- Mantém users, refresh_tokens, password_reset_tokens e schema_migrations intactos.
BEGIN;

DELETE FROM reservations;
DELETE FROM guests;
DELETE FROM rooms;
DELETE FROM room_categories;

INSERT INTO room_categories (id, name, description) VALUES
    ('10000000-0000-4000-8000-000000000001', 'Standard', 'Acomodação confortável para até duas pessoas.'),
    ('10000000-0000-4000-8000-000000000002', 'Suíte', 'Acomodação ampla com espaço de estar.'),
    ('10000000-0000-4000-8000-000000000003', 'Família', 'Acomodação para viagens em família.');

INSERT INTO rooms (id, number, category_id, capacity, daily_rate, is_available) VALUES
    ('20000000-0000-4000-8000-000000000001', '101', '10000000-0000-4000-8000-000000000001', 2, 280.00, TRUE),
    ('20000000-0000-4000-8000-000000000002', '102', '10000000-0000-4000-8000-000000000001', 2, 280.00, TRUE),
    ('20000000-0000-4000-8000-000000000003', '201', '10000000-0000-4000-8000-000000000002', 2, 420.00, TRUE),
    ('20000000-0000-4000-8000-000000000004', '202', '10000000-0000-4000-8000-000000000002', 3, 460.00, TRUE),
    ('20000000-0000-4000-8000-000000000005', '301', '10000000-0000-4000-8000-000000000003', 4, 560.00, TRUE),
    ('20000000-0000-4000-8000-000000000006', '302', '10000000-0000-4000-8000-000000000003', 5, 620.00, TRUE);

INSERT INTO guests (id, name, document_type, document, document_normalized, email, phone) VALUES
    ('30000000-0000-4000-8000-000000000001', 'Ana Souza', 'CPF', '529.982.247-25', '52998224725', 'ana.souza@demo.hotel.test', '11987654321'),
    ('30000000-0000-4000-8000-000000000002', 'Bruno Martins', 'RG', 'MG-12.345.678', 'MG12345678', 'bruno.martins@demo.hotel.test', '11976543210'),
    ('30000000-0000-4000-8000-000000000003', 'Carla Oliveira', 'CPF', '111.444.777-35', '11144477735', 'carla.oliveira@demo.hotel.test', '11965432109');

-- Reserva futura para demonstrar a exclusão de um quarto em período conflitante.
INSERT INTO reservations (
    id, guest_id, room_id, check_in, check_out, occupants, daily_rate, total_amount,
    guest_name_snapshot, document_type_snapshot, document_snapshot, contact_email_snapshot
) VALUES (
    '40000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000001',
    '20000000-0000-4000-8000-000000000001',
    CURRENT_DATE + 14,
    CURRENT_DATE + 17,
    2,
    280.00,
    840.00,
    'Ana Souza',
    'CPF',
    '529.982.247-25',
    'ana.souza@demo.hotel.test'
);

COMMIT;
