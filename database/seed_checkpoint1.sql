-- Dados adicionais para validar o checkpoint 1. Não apaga dados existentes.
-- Reexecutar não redefine os cenários existentes nem desfaz testes manuais.
-- Requer migração 006 e ao menos um funcionário ativo.
BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE is_active AND role IN ('admin', 'reception')) THEN
        RAISE EXCEPTION 'Crie um funcionário ativo antes de carregar os cenários CP1.';
    END IF;
END;
$$;

CREATE TEMP TABLE cp1_cases ON COMMIT DROP AS
SELECT n, label, entry_offset, exit_offset, cancelled,
       ('a2000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid AS room_id,
       ('a3000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid AS guest_id,
       ('a4000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid AS reservation_id,
       (now() AT TIME ZONE 'America/Sao_Paulo')::date AS today
FROM (VALUES
    (1, 'Confirmada', 3, 5, false),
    (2, 'Em andamento', -1, 2, false),
    (3, 'Saída hoje', -2, 0, false),
    (4, 'Encerrada', -5, -3, false),
    (5, 'Cancelada', 3, 5, true)
) AS cases(n, label, entry_offset, exit_offset, cancelled);

INSERT INTO room_categories (id, name, description)
VALUES ('a1000000-0000-4000-8000-000000000001', 'Teste CP1', 'Cenários de validação de status, disponibilidade e indicadores.')
ON CONFLICT (id) DO NOTHING;

INSERT INTO rooms (id, number, category_id, capacity, daily_rate, is_available)
SELECT room_id, 'CP1-0' || n, 'a1000000-0000-4000-8000-000000000001'::uuid, 2, 100.00, true
FROM cp1_cases
ON CONFLICT (id) DO NOTHING;

INSERT INTO guests (id, name, document_type, document, document_normalized, email)
SELECT guest_id, 'Teste CP1 - ' || label, 'RG', 'CP1-0' || n, 'CP10' || n,
       'cp1-' || n || '@demo.hotel.test'
FROM cp1_cases
ON CONFLICT (id) DO NOTHING;

INSERT INTO reservations (
    id, guest_id, room_id, check_in, check_out, occupants, daily_rate, total_amount,
    guest_name_snapshot, document_type_snapshot, document_snapshot, contact_email_snapshot,
    cancelled_at, cancelled_by, version
)
SELECT reservation_id, guest_id, room_id, today + entry_offset, today + exit_offset,
       2, 100.00, (exit_offset - entry_offset) * 100.00,
       'Teste CP1 - ' || label, 'RG', 'CP1-0' || n, 'cp1-' || n || '@demo.hotel.test',
       CASE WHEN cancelled THEN now() END,
       CASE WHEN cancelled THEN (
           SELECT id FROM users WHERE is_active AND role IN ('admin', 'reception') ORDER BY id LIMIT 1
       ) END,
       CASE WHEN cancelled THEN 2 ELSE 1 END
FROM cp1_cases
ON CONFLICT (id) DO NOTHING;

SELECT rm.number AS quarto, r.guest_name_snapshot AS hospede,
       r.check_in AS entrada, r.check_out AS saida,
       r.cancelled_at, r.version
FROM reservations r JOIN rooms rm ON rm.id = r.room_id
WHERE r.id IN (SELECT reservation_id FROM cp1_cases)
ORDER BY rm.number;

COMMIT;
