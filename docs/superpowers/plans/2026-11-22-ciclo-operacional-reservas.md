# Ciclo operacional da reserva — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Concluir uma reserva com check-in realizado, quarto ocupado, checkout, cálculo do valor final e quarto liberado.

**Architecture:** Reutilizar o serviço `hotel`, o repositório PostgreSQL e a auditoria existentes. Persistir os fatos operacionais e uma referência única entre quarto ocupado e reserva; executar cada transição em transação com bloqueio de quarto, versão e auditoria. Expor ações autenticadas e atualizar a área de reservas e o cadastro de quartos.

**Tech Stack:** Go 1.24+, PostgreSQL, React, TypeScript, Vite e Tailwind CSS.

**Spec:** `docs/superpowers/specs/2026-11-22-checkin-checkout-design.md`

## Global Constraints

- Datas e horários realizados são armazenados como instantes UTC e exibidos em `America/Sao_Paulo`.
- O cálculo final é a diária guardada na reserva multiplicada pelas noites entre as datas civis realizadas; check-in antecipado permitido aumenta as noites cobradas.
- Checkout anterior ao check-in realizado é rejeitado; checkout no mesmo dia resulta em zero noites e valor zero.
- `rooms.is_available` continua indicando habilitação para venda; ocupação é um estado separado.
- Todas as operações são autenticadas para administrador e recepção e usam transação para manter reserva, quarto e auditoria consistentes.
- Reservas legadas preservam datas e valores previstos e continuam usando status derivado enquanto não houver operação realizada.
- Não incluir pagamentos, descontos, consumo, reembolsos, multas, cobrança por hora ou troca de quarto durante a estadia.

---

## Mapa de arquivos

**Banco e domínio**

- Criar `database/migrations/008_reservation_stay.sql` para timestamps reais, autoria, cálculo final, referência de quarto ocupado e extensão da auditoria.
- Modificar `database/migrate.sql` e `database/README.md` para registrar/aplicar a versão 008.
- Modificar `database/seed_demo.sql` para limpar a referência de ocupação antes de apagar reservas de demonstração.
- Modificar `backend/internal/hotel/rooms.go` e `rooms_postgres.go` para expor ocupação atual separada da habilitação.
- Modificar `backend/internal/hotel/reservations.go`, `reservations_postgres.go`, `reservation_editing_postgres.go`, `reservation_cancellation.go` e `reservation_status.go` para transportar fatos realizados, usar ocupação efetiva nos conflitos/indicadores e impedir cancelamento após check-in.
- Criar `backend/internal/hotel/reservation_stay.go` e `reservation_stay_postgres.go` para contratos e transações de check-in/checkout.
- Criar testes unitários e de integração em `reservation_stay_test.go` e `reservation_stay_integration_test.go`; estender testes de status, quartos e cancelamento onde necessário.

**HTTP e cliente**

- Criar `backend/internal/httpapi/reservation_stay.go`; modificar `router.go`, `reservations.go`, `reservation_cancellation.go` e os testes HTTP pertinentes.
- Modificar `frontend/src/services/hotel.ts` para campos/status, dados de quarto e funções de ação.
- Modificar `frontend/src/components/ReservationStatusBadge.tsx`, `frontend/src/pages/ReservationsListPage.tsx`, `frontend/src/pages/ReservationEditPage.tsx`, `frontend/src/pages/ReservationAuditPage.tsx` e `frontend/src/pages/DashboardPage.tsx` para status e ações operacionais.
- Modificar `frontend/src/pages/RecordsPage.tsx` para distinguir quarto ocupado de quarto habilitado/desabilitado.
- Criar `frontend/src/components/ReservationStayActions.tsx` para confirmações e memória de cálculo, usando o detalhe da reserva.
- Atualizar `README.md`, `backend/README.md`, `frontend/README.md` e criar roteiro `docs/VALIDACAO_CICLO_RESERVA.md`.

---

### Task 1: Persistir estado operacional e dados realizados

**Files:**
- Create: `database/migrations/008_reservation_stay.sql`
- Modify: `backend/internal/hotel/reservations.go`
- Modify: `backend/internal/hotel/reservation_cancellation.go`
- Modify: `backend/internal/hotel/reservation_status.go`
- Modify: `backend/internal/hotel/rooms.go`
- Modify: `backend/internal/hotel/rooms_postgres.go`
- Test: `backend/internal/hotel/reservation_status_test.go`
- Test: `backend/internal/hotel/reservation_stay_integration_test.go`
- Test: `backend/internal/hotel/integration_test.go`
- Modify: `backend/internal/hotel/reservation_integration_test.go` e `reservation_management_integration_test.go` para aplicar a migração 008 nos bancos isolados.

**Interfaces:**
- Adicionar à reserva `CheckedInAt`, `CheckedInBy`, `CheckedOutAt`, `CheckedOutBy`, `ActualNights` e `FinalAmount` (JSON `checkedInAt`, `checkedInBy`, `checkedOutAt`, `checkedOutBy`, `actualNights`, `finalAmount`).
- Adicionar à sala `IsOccupied bool` e `OccupiedByReservationID *string` (JSON `isOccupied`, `occupiedByReservationId`).
- Adicionar status `in_house` e `completed`; status real tem precedência sobre os status previstos e cancelamento continua prioritário.
- Migração adiciona `rooms.occupied_by_reservation_id UUID UNIQUE REFERENCES reservations(id) ON DELETE RESTRICT`; reserva adiciona colunas de check-in/out e valor final com constraints de consistência.

- [ ] **Step 1: Escrever testes de status e de leitura dos campos de ocupação.** Cobrir reserva legada, check-in real, checkout real, cancelamento e quarto ocupado/vago.
- [ ] **Step 2: Rodar os testes focados para confirmar a falha inicial.**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/hotel -run 'TestReservationStatus|TestRoomOccupancy' -count=1`

Expected: falhas de compilação ou de comportamento por falta dos novos campos/status.

- [ ] **Step 3: Criar a migração aditiva.** Adicionar `checked_in_at`, `checked_in_by`, `checked_out_at`, `checked_out_by`, `actual_nights`, `final_amount` e `occupied_by_reservation_id`; preservar registros antigos como não operados. Adicionar constraints para autoria/data em pares, checkout exigir check-in, noites e valor não negativos e campos de checkout consistentes. Atualizar o check de `reservation_events.action` e a função/snapshot de auditoria para registrar `checked_in` e `checked_out`, ator, horário e estado operacional.
- [ ] **Step 4: Atualizar os modelos e consultas de leitura.** Expandir os `SELECT`/`Scan` centrais para ler colunas novas; incluir `isOccupied` na consulta de quartos sem alterar `is_available`.
- [ ] **Step 5: Atualizar o status calculado.** Retornar `cancelled` para canceladas, `completed` para checkout realizado, `in_house` para check-in realizado e, nos demais casos, manter a regra existente por data.
- [ ] **Step 6: Rodar os testes focados e integração.**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/hotel -run 'TestReservationStatus|TestRoomOccupancy|TestIntegration' -count=1`

Expected: PASS; a migração mantém reservas antigas sem datas realizadas nem valor final.

---

### Task 2: Implementar transições transacionais no domínio

**Files:**
- Create: `backend/internal/hotel/reservation_stay.go`
- Create: `backend/internal/hotel/reservation_stay_postgres.go`
- Modify: `backend/internal/hotel/reservation_cancellation.go`
- Modify: `backend/internal/hotel/reservations.go`
- Test: `backend/internal/hotel/reservation_stay_test.go`
- Test: `backend/internal/hotel/reservation_stay_integration_test.go`
- Test: `backend/internal/hotel/reservation_cancellation_integration_test.go`

**Interfaces:**
- `type ReservationTransitionInput struct { ReservationID string; ActorID string; Version int }`
- `func (s *Service) CheckInReservation(ctx context.Context, input ReservationTransitionInput) (Reservation, error)`
- `func (s *Service) CheckOutReservation(ctx context.Context, input ReservationTransitionInput) (Reservation, error)`
- `func (s *Service) PreviewCheckOutReservation(ctx context.Context, input ReservationTransitionInput) (ReservationCheckoutQuote, error)` retorna `CheckoutAt`, `ActualNights`, `DailyRate` e `FinalAmount` usando relógio e fuso do servidor.
- Repositório privado implementa os mesmos métodos; erros existentes `ErrNotFound` e `ErrReservationState` representam inexistência e transição/versão incompatível.

- [ ] **Step 1: Escrever testes unitários de validação.** UUID/ator inválidos e versão menor que 1 devem produzir `ValidationError`; teste de cálculo para noites negativas, zero noites, noites normais e decimal em centavos.
- [ ] **Step 2: Rodar os testes para verificar a falha inicial.**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/hotel -run 'TestCheckInValidation|TestCheckOutValidation|TestActualStayTotal' -count=1`

Expected: FAIL até existirem os contratos e o cálculo.

- [ ] **Step 3: Implementar cálculo de noites civis e dinheiro decimal.** Converter timestamps para `America/Sao_Paulo`, calcular diferença de datas UTC civis; rejeitar checkout anterior à entrada; permitir mesmo dia com zero noites; multiplicar centavos inteiros por noites e serializar com duas casas.
- [ ] **Step 4: Testar check-in em integração PostgreSQL.** Deve bloquear a sala e reserva, validar status/versão/cancelamento, permitir antecipação sem conflito de outra reserva ativa, gravar ator/horário e incrementar versão.
- [ ] **Step 5: Implementar check-in transacional.** Bloquear quarto antes da reserva conforme ordem já usada em criação/cancelamento; rejeitar `occupied_by_reservation_id` preenchido; exigir que a data efetiva seja anterior ao checkout previsto; rejeitar outra reserva não cancelada ocupando a data real do check-in (considerar saída real quando existir); definir ator/source da auditoria; gravar timestamp, referência do quarto e versão numa transação.
- [ ] **Step 5a: Aplicar o intervalo realizado em disponibilidade e indicadores.** Criar/editar reservas e resumo usam check-in realizado como início e check-out realizado como fim quando disponível; enquanto hospedada, checkout previsto é o fim de referência. Testar tentativa sobreposta com chegada antecipada e criação para período futuro posterior ao checkout previsto.
- [ ] **Step 6: Testar checkout em integração PostgreSQL.** Cobrir checkout sem entrada, checkout anterior, mesma data, múltiplas noites, diária do quarto alterada após reserva, repetição e concorrência.
- [ ] **Step 7: Implementar checkout transacional.** Bloquear quarto e reserva na mesma ordem; calcular noites efetivas e valor pela `daily_rate` congelada; gravar horário, noites, valor final e versão; remover `occupied_by_reservation_id` somente se apontar para a reserva em checkout; produzir o evento auditado na mesma transação.
- [ ] **Step 8: Impedir cancelamento de reserva iniciada.** Atualizar `CanCancel`/validação do fluxo atual para rejeitar reserva com `checked_in_at` preenchido; manter cancelamento de reserva futura sem check-in.
- [ ] **Step 9: Verificar atomicidade e concorrência.** Em integração, forçar falha de auditoria e garantir rollback conjunto; disparar duas entradas simultâneas e duas saídas simultâneas, confirmando uma única transição bem-sucedida e nenhuma sala presa/liberada indevidamente.
- [ ] **Step 10: Rodar integração do domínio.**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/hotel -run 'Test(CheckIn|CheckOut|ActualStay|Cancel)' -count=1`

Expected: PASS com `TEST_DATABASE_URL` configurada; nenhum teste de PostgreSQL pode ser pulado.

---

### Task 3: Expor endpoints autenticados e estado consistente

**Files:**
- Create: `backend/internal/httpapi/reservation_stay.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/httpapi/reservations.go`
- Modify: `backend/internal/httpapi/reservation_cancellation.go`
- Modify: `backend/internal/httpapi/reservations_test.go`
- Test: `backend/internal/httpapi/reservation_stay_test.go`

**Interfaces:**
- `POST /api/v1/reservations/{id}/check-in` com corpo `{"version": n}`.
- `POST /api/v1/reservations/{id}/check-out` com corpo `{"version": n}`.
- `GET /api/v1/reservations/{id}/check-out-quote?version=n` calcula a prévia no servidor sem gravar.
- Os dois POSTs retornam `{"reservation": ...}` com status, datas realizadas, noites, valor final e versão atual.
- A prévia retorna `{"quote":{"checkoutAt":...,"actualNights":n,"dailyRate":"...","finalAmount":"...","version":n}}`; checkout sempre recalcula em transação e sua resposta é o valor definitivo.
- Endpoints exigem `requireUser`; erros usam `writeReservationError`, com 404 inexistente, 409 transição/versão e 422 validação.

- [ ] **Step 1: Escrever testes HTTP para autenticação, parsing e transições.** Incluir sem token (401), body malformado (422), versão stale (409), reserva ausente (404) e payload de sucesso.
- [ ] **Step 2: Rodar os testes HTTP focados e verificar falha inicial.**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/httpapi -run 'TestReservation(CheckIn|CheckOut)' -count=1`

Expected: FAIL até as rotas e handlers serem registrados.

- [ ] **Step 3: Implementar handlers e registrar rotas.** Extrair user autenticado, validar body/ID, chamar os métodos de serviço e serializar com o response mapper usado nas reservas atuais.
- [ ] **Step 4: Implementar prévia do checkout.** Validar versão e reserva em hospedagem; calcular `checkoutAt`, noites e valor com o relógio do servidor, sem mutação; documentar que o checkout recalcula e retorna o valor autoritativo.
- [ ] **Step 5: Incluir novos campos no retorno das reservas e quartos.** Atualizar `reservationDateResponse` e `staffReservationResponse`; `GET /rooms` e `GET /rooms/{id}` expõem ocupação junto da flag de habilitação.
- [ ] **Step 6: Rodar os testes HTTP e os testes completos do pacote.**

Run: `cd backend && CGO_ENABLED=0 go test ./internal/httpapi ./internal/hotel -count=1`

Expected: PASS; endpoints operacionais não ficam disponíveis sem autenticação.

---

### Task 4: Adicionar ações à recepção

**Files:**
- Create: `frontend/src/components/ReservationStayActions.tsx`
- Modify: `frontend/src/services/hotel.ts`
- Modify: `frontend/src/components/ReservationStatusBadge.tsx`
- Modify: `frontend/src/pages/ReservationsListPage.tsx`
- Modify: `frontend/src/pages/ReservationEditPage.tsx`
- Modify: `frontend/src/pages/RecordsPage.tsx`

**Interfaces:**
- Tipos incluem `in_house` e `completed`; reserva inclui timestamps reais, noites, `finalAmount` e ações permitidas pelo estado.
- Cliente exporta `checkInReservation(id, version)` e `checkOutReservation(id, version)` para os novos endpoints.
- `ReservationStayActions` recebe a reserva e callbacks de atualização; checkout mostra datas/horários, noites efetivas, diária contratada e total final retornado pelo servidor.

- [ ] **Step 1: Implementar tipos e chamadas autenticadas no serviço frontend.** Reutilizar `authenticatedRequest` e manter `version` no payload.
- [ ] **Step 2: Adicionar rótulos distintos aos status.** Usar “Em hospedagem” para `in_house` e “Finalizada” para `completed`, preservando labels previstos e cancelada.
- [ ] **Step 3: Construir componente de ações.** Mostrar check-in apenas em reserva elegível; checkout apenas em hospedagem; buscar e exibir a prévia do servidor antes de confirmar; estados busy/erro; em 409 oferecer recarga da reserva; mostrar recibo com o valor definitivo após sucesso.
- [ ] **Step 4: Integrar ações às superfícies atuais de reservas.** Incluir acesso no detalhe e ação rápida na listagem sem remover edição/cancelamento existentes; atualizar conteúdo após sucesso.
- [ ] **Step 5: Exibir ocupação na tela de quartos.** Mostrar “Ocupado” separado de “Disponível para venda”/“Indisponível”; impedir toggle administrativo para livre enquanto ocupado.
- [ ] **Step 6: Executar lint e build do frontend.**

Run: `cd frontend && npm run lint && npm run build`

Expected: PASS sem erros TypeScript ou lint.

---

### Task 5: Atualizar contrato, roteiro e validação integrada

**Files:**
- Modify: `README.md`
- Modify: `backend/README.md`
- Modify: `frontend/README.md`
- Create: `docs/VALIDACAO_CICLO_RESERVA.md`

- [ ] **Step 1: Documentar migração, estados, endpoints e regra de preço.** Explicar a diferença entre datas previstas e realizadas, o fuso do hotel, o campo de habilitação de quarto e o estado ocupado.
- [ ] **Step 2: Escrever roteiro manual de aceitação.** Incluir administrador/recepção, chegada antecipada com conflito e sem conflito, check-in normal, tentativa repetida, prévia do checkout, checkout no mesmo dia, checkout com noites, diária alterada, quarto ocupado, quarto liberado, reserva futura e concorrência com versão obsoleta.
- [ ] **Step 3: Rodar todos os testes backend e validações estáticas.**

Run: `cd backend && CGO_ENABLED=0 go test ./... -count=1 && CGO_ENABLED=0 go vet ./...`

Expected: PASS com banco de integração configurado; nenhuma integração PostgreSQL pulada.

- [ ] **Step 4: Rodar lint/build e conferir diff.**

Run: `cd frontend && npm run lint && npm run build`

Run: `git diff --check`

Expected: todos os comandos PASS.

- [ ] **Step 5: Validar migração em banco novo e upgrade existente.** Executar as migrações 001–008 em banco vazio; repetir executor; atualizar banco com reservas antes da 008 e confirmar reservas preservadas sem check-in, valor final nulo e quartos desocupados.
- [ ] **Step 6: Validar fluxo autenticado no navegador.** Seguir roteiro em desktop e celular, conferir ambos os perfis, eventos de auditoria, memória de cálculo e estado de quarto após cada transição.

## Cobertura da especificação

- Check-in, check-out, cálculo, estados, ocupação e transações: Tasks 1–2.
- Endpoints e autenticação: Task 3.
- Ações e estado visível no painel: Task 4.
- Documentação e aceite fim a fim: Task 5.
- Pagamentos e regras fora do escopo permanecem excluídos.
