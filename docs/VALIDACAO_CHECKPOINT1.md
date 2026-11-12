# Validação do checkpoint 1

Carregar dados adicionais na raiz do projeto:

```bash
docker-compose exec -T database psql -U hotel_laranjeira -d hotel_laranjeira -v ON_ERROR_STOP=1 < database/seed_checkpoint1.sql
```

O script não remove dados existentes. Cria cinco quartos, cinco hóspedes e cinco
reservas com nomes `Teste CP1`. As datas são relativas à data de São Paulo na
primeira execução. Reexecutar não altera registros já criados.

## 1. Status e preservação

Abra `/reservas`, pesquise `Teste CP1` e confira os cinco registros. Na aba
Rede/Network do navegador, abra a resposta de `GET /api/v1/reservations`:

| Quarto | Entrada / saída em relação à carga | status esperado na data da carga |
| --- | --- | --- |
| CP1-01 | +3 / +5 dias | confirmed |
| CP1-02 | -1 / +2 dias | in_progress |
| CP1-03 | -2 / hoje | departure_today |
| CP1-04 | -5 / -3 dias | ended |
| CP1-05 | +3 / +5 dias, cancelada | cancelled |

Neste checkpoint, a tela ainda calcula seus próprios rótulos pelas datas.
O campo `status` da resposta da API é a referência para validar o backend;
a exibição da reserva cancelada será corrigida no checkpoint 2.
A ação de cancelar pela tela/API será entregue no checkpoint 3; aqui o script
insere uma reserva já cancelada com funcionário e horário registrados.

## 2. Disponibilidade

Na busca pública `/reservar`, informe duas pessoas e as datas de CP1-01 (+3 a
+5 dias). CP1-01 deve estar ausente, enquanto CP1-05 deve estar disponível,
apesar da reserva cancelada no mesmo período.

Para hoje até amanhã, CP1-02 deve estar ausente e CP1-03 disponível: a noite
do checkout pode ser usada por uma nova reserva. Não confirme novas reservas
nesses quartos até concluir essa conferência, para preservar os resultados.

## 3. Indicadores

No resumo de hoje, considerando somente os cinco quartos CP1:

- 2 hóspedes previstos em estadia (CP1-02).
- 0 entradas previstas hoje.
- 4 quartos disponíveis.

O painel inclui também os dados anteriores do projeto. Para comparar o resumo
com a API, consulte os totais completos no terminal:

```sql
WITH reference AS (
    SELECT (now() AT TIME ZONE 'America/Sao_Paulo')::date AS today
)
SELECT
    (SELECT coalesce(sum(occupants), 0) FROM reservations, reference
     WHERE cancelled_at IS NULL AND check_in <= today AND check_out > today) AS guests_in_house,
    (SELECT count(*) FROM reservations, reference
     WHERE cancelled_at IS NULL AND check_in = today) AS check_ins_today,
    (SELECT count(*) FROM rooms rm, reference
     WHERE rm.is_available AND NOT EXISTS (
         SELECT 1 FROM reservations r WHERE r.room_id = rm.id
         AND r.cancelled_at IS NULL AND r.check_in <= today AND r.check_out > today
     )) AS available_rooms;
```

## 4. Migração e fuso

As validações de preservação de reservas anteriores, restrições do banco,
virada de data de São Paulo e concorrência são automatizadas. Para repeti-las,
configure `TEST_DATABASE_URL` e execute em `backend`:

```bash
CGO_ENABLED=0 go test ./internal/hotel -v -count=1
```

Os testes PostgreSQL precisam executar sem `SKIP`; usam esquemas temporários.
