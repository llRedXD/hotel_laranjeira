# Reservas de hospedagem — plano de implementação

> **Para agentes de implementação:** use `superpowers:subagent-driven-development` (recomendado) ou `superpowers:executing-plans` para executar este plano tarefa por tarefa. Marque cada etapa com checkbox.

**Goal:** Permitir que hóspedes reservem um quarto pelo site e que a recepção faça reservas pelo painel, com preço previsto e proteção contra conflitos.

**Architecture:** A API Go mantém as regras de disponibilidade, capacidade, diárias e criação transacional no domínio `hotel`. O PostgreSQL guarda reservas e snapshots dos dados informados. A interface React oferece dois pontos de entrada — público e painel autenticado — que consomem os mesmos contratos de disponibilidade e criação.

**Tech Stack:** Go, `net/http`, pgx v5, PostgreSQL, React, TypeScript, React Router e Vite, seguindo a estrutura existente no repositório.

**Spec:** `docs/superpowers/specs/2026-10-15-reservas-hospedagem-design.md`

## Restrições globais

- O hóspede não cria conta nem autentica nesta entrega.
- Check-in e check-out são datas previstas (`DATE`); check-out deve ser posterior ao check-in.
- O intervalo ocupado é `[check-in, check-out)`, e o hóspede responsável conta no total de ocupantes.
- Só quartos com `is_available = true`, capacidade suficiente e sem conflito podem ser oferecidos.
- A API calcula o número de diárias e o total usando moeda decimal; nunca usar `float64` para valores monetários.
- A confirmação persiste a diária aplicada e o total, sem pagamento ou aprovação manual.
- A reserva pública exige nome, tipo de documento (`RG` ou `CPF`), número e e-mail; sua interface mascara e valida os dígitos do CPF antes da revisão. O RG permanece sem máscara fixa e a API conserva validação de campos obrigatórios e tamanho.
- Registros legados mantêm tipo de documento `NULL` até classificação explícita pela recepção.
- O contato informado na reserva é salvo em snapshot e não substitui os dados principais do hóspede.
- Rotas atuais de gestão de hóspedes e quartos permanecem autenticadas; a API pública não retorna dados cadastrais.
- Não incluir contas de hóspedes, pagamento, cancelamento, notificações, tarifa sazonal, vários quartos por operação ou check-in/check-out realizado.

## Mapa de arquivos

| Arquivo | Responsabilidade nesta entrega |
| --- | --- |
| `database/migrations/005_reservations.sql` | Adicionar tipo/normalização de documento e tabelas, restrições e índices de reservas. |
| `database/migrate.sql` | Aplicar a versão 5 na sequência versionada existente. |
| `backend/internal/hotel/guests.go` | Expor `documentType` para os contratos internos de hóspede. |
| `backend/internal/hotel/service.go` | Validar e delegar tipo de documento e operações de reserva ao repositório. |
| `backend/internal/hotel/reservations.go` | Tipos, erros e regras de reserva e disponibilidade. |
| `backend/internal/hotel/reservations_postgres.go` | Consultas e transação PostgreSQL, incluindo associação/criação segura do hóspede. |
| `backend/internal/hotel/service_test.go` | Atualizar os dublês de repositório e preservar regressões do serviço de hóspedes. |
| `backend/internal/hotel/reservations_test.go` | Testes unitários de validação, cálculo e regras de datas/capacidade. |
| `backend/internal/hotel/integration_test.go` | Cobertura PostgreSQL de migração, disponibilidade, criação, snapshots e concorrência. |
| `backend/internal/httpapi/hotel.go` | Aceitar e serializar tipo de documento nas rotas internas existentes. |
| `backend/internal/httpapi/reservations.go` | Handlers públicos e autenticados de disponibilidade e criação. |
| `backend/internal/httpapi/router.go` | Registrar rotas públicas e rotas que exigem Bearer JWT. |
| `backend/internal/httpapi/reservations_test.go` | Contratos HTTP, autenticação e ausência de dados privados no público. |
| `frontend/src/services/reservations.ts` | Tipos e chamadas públicas de disponibilidade/criação. |
| `frontend/src/services/api.ts` | Expor uma chamada HTTP sem autenticação, compartilhando base URL, cookies e tratamento de erros. |
| `frontend/src/services/hotel.ts` | Campo `documentType` e chamadas autenticadas da recepção. |
| `frontend/src/pages/ReservationPage.tsx` | Busca pública, escolha de quarto, dados do hóspede, revisão e confirmação. |
| `frontend/src/pages/ReceptionReservationsPage.tsx` | Fluxo de reserva para hóspede existente no painel. |
| `frontend/src/components/ReservationSearchForm.tsx` | Formulário compartilhado de período e ocupantes. |
| `frontend/src/components/ReservationRoomOptions.tsx` | Opções disponíveis com diária, noites e total estimado. |
| `frontend/src/App.tsx` | Rota pública e rota interna protegida. |
| `frontend/src/components/DashboardLayout.tsx` | Habilitar navegação “Reservas”. |
| `frontend/src/components/EntityDialog.tsx` | Classificar documento legado como RG ou CPF no cadastro interno. |
| `frontend/src/pages/RecordsPage.tsx` | Exibir tipo de documento no cadastro/listagem de hóspedes. |
| `README.md`, `backend/README.md`, `frontend/README.md`, `database/README.md` | Atualizar estado da entrega, contratos e instruções locais. |

## Contratos compartilhados

O domínio usa estes tipos em `backend/internal/hotel/reservations.go`:

```go
type AvailabilityQuery struct {
	CheckIn   time.Time
	CheckOut  time.Time
	Occupants int
}

type AvailabilityOption struct {
	RoomID       string `json:"roomId"`
	Number       string `json:"number"`
	CategoryName string `json:"categoryName"`
	Capacity     int    `json:"capacity"`
	DailyRate    string `json:"dailyRate"`
	Nights       int    `json:"nights"`
	EstimatedTotal string `json:"estimatedTotal"`
}

type ReservationGuestInput struct {
	Name         string `json:"name"`
	DocumentType string `json:"documentType"`
	Document     string `json:"document"`
	Email        string `json:"email"`
}

type ReservationInput struct {
	RoomID    string                 `json:"roomId"`
	GuestID   string                 `json:"guestId,omitempty"`
	Guest     *ReservationGuestInput `json:"guest,omitempty"`
	CheckIn   time.Time
	CheckOut  time.Time
	Occupants int                    `json:"occupants"`
	ExpectedDailyRate string
}

type Reservation struct {
	ID             string    `json:"id"`
	GuestID        string    `json:"guestId"`
	RoomID         string    `json:"roomId"`
	CheckIn        time.Time `json:"checkIn"`
	CheckOut       time.Time `json:"checkOut"`
	Occupants      int       `json:"occupants"`
	DailyRate      string    `json:"dailyRate"`
	TotalAmount    string    `json:"totalAmount"`
	GuestName      string    `json:"guestName"`
	DocumentType   string    `json:"documentType"`
	Document       string    `json:"document"`
	ContactEmail   string    `json:"contactEmail"`
	CreatedAt      time.Time `json:"createdAt"`
}
```

Os handlers decodificam datas HTTP `YYYY-MM-DD` em campos string de DTOs próprios e só então constroem `AvailabilityQuery`/`ReservationInput` com `time.Time`; não decodificar `YYYY-MM-DD` diretamente em `time.Time`, que espera timestamp RFC3339 no JSON padrão.

O serviço expõe `ListAvailability(context.Context, AvailabilityQuery) ([]AvailabilityOption, error)` e `CreateReservation(context.Context, ReservationInput) (Reservation, error)`. O repositório acrescenta `ListAvailability(context.Context, AvailabilityQuery) ([]AvailabilityOption, error)` e `CreateReservation(context.Context, ReservationInput) (Reservation, error)` à interface `Repository`. `ReservationInput` aceita exatamente um dos campos `GuestID` ou `Guest`: o primeiro é exclusivo do painel autenticado; o segundo é exclusivo do site. O mesmo método de domínio valida ambos.

Contratos HTTP que o front-end consome:

- `GET /api/v1/public/availability?checkIn=YYYY-MM-DD&checkOut=YYYY-MM-DD&occupants=N` retorna `{"rooms":[AvailabilityOption...]}` sem informações de hóspedes.
- `POST /api/v1/public/reservations` recebe `roomId`, `checkIn`, `checkOut`, `occupants`, `expectedDailyRate` e `guest` e retorna HTTP 201 com confirmação sem dados documentais ou e-mail: `{"reservation":{"id":"...","roomId":"...","checkIn":"YYYY-MM-DD","checkOut":"YYYY-MM-DD","occupants":2,"dailyRate":"125.50","totalAmount":"251.00"}}`.
- `GET /api/v1/reservations/availability?...` exige Bearer JWT e retorna o mesmo formato de disponibilidade.
- `POST /api/v1/reservations` exige Bearer JWT e recebe `roomId`, `checkIn`, `checkOut`, `occupants`, `expectedDailyRate` e `guestId`; retorna HTTP 201 com os dados operacionais da reserva, sem telefone/endereço.
- Erros de validação retornam HTTP 422; conflito de disponibilidade, documento ou e-mail retorna HTTP 409; IDs inexistentes retornam HTTP 404. Respostas públicas não confirmam dados do cadastro eventualmente associado.

## Tarefas

### Tarefa 1: Evoluir documento de hóspede e esquema de reservas

**Arquivos:** criar `database/migrations/005_reservations.sql`; modificar `database/migrate.sql`, `backend/internal/hotel/guests.go` e `backend/internal/httpapi/hotel.go`; criar/estender testes de integração em `backend/internal/hotel/integration_test.go`.

**Interfaces:** `Guest` recebe `DocumentType *string` com `json:"documentType"`; `GuestInput` e `GuestPatch` carregam o tipo. A coluna aceita `RG`, `CPF` ou `NULL` para legados.

- [ ] **Passo 1: Escrever o teste da migração preservando hóspedes legados.** No teste PostgreSQL, criar a estrutura 003, inserir hóspede com documento antes de aplicar a migração 005 e verificar que documento permanece igual e `document_type IS NULL`.

```go
if _, err := tx.Exec(ctx, `INSERT INTO guests (name, document) VALUES ('Legado', '12.345-X')`); err != nil {
	t.Fatal(err)
}
// Aplicar a migração 005 e consultar document/document_type.
```

- [ ] **Passo 2: Rodar o teste para confirmar a falha inicial.** Execute `cd backend && TEST_DATABASE_URL="$TEST_DATABASE_URL" go test ./internal/hotel -run TestReservationMigrationPreservesLegacyGuest -v`. Resultado esperado antes da migração: falha porque `document_type` e a migração 005 ainda não existem.
- [ ] **Passo 3: Criar a migração e registrar a versão 5.** Adicionar colunas `document_type` e `document_normalized`; normalizar documentos existentes sem alterar `document`; abortar a transação com erro legível se houver colisão de normalização; trocar o índice antigo por índice único parcial em `document_normalized`; manter unicidade do e-mail. Criar `reservations` com `id UUID`, `guest_id UUID`, `room_id UUID`, `check_in DATE`, `check_out DATE`, `occupants INTEGER`, `daily_rate NUMERIC(12,2)`, `total_amount NUMERIC(12,2)`, `guest_name_snapshot VARCHAR(120)`, `document_type_snapshot VARCHAR(3)`, `document_snapshot VARCHAR(30)`, `contact_email_snapshot VARCHAR(255)` e `created_at TIMESTAMPTZ`; FKs com `ON DELETE RESTRICT`; checks de checkout posterior, ocupantes positivos, RG/CPF e preços não negativos; índices em `(room_id, check_in, check_out)` e `(guest_id, created_at)`. Incluir no `database/migrate.sql` o bloco condicional da versão 5 após a versão 4.
- [ ] **Passo 4: Completar o teste para verificar constraints e índices.** Testar tipo legado nulo, tipo inválido rejeitado, `check_out <= check_in` rejeitado, ocupantes não positivos rejeitados, valores com duas casas e documento normalizado único.
- [ ] **Passo 5: Executar o ciclo da migração.** Execute `psql -X -v ON_ERROR_STOP=1 --single-transaction "$TEST_DATABASE_URL" -f database/migrate.sql` em banco descartável, depois `cd backend && TEST_DATABASE_URL="$TEST_DATABASE_URL" go test ./internal/hotel -run TestReservationMigration -v`. Resultado esperado: migração aplicada uma vez, segunda aplicação sem alterações e testes aprovados.
- [ ] **Passo 6: Registrar o incremento.** `git add database/migrations/005_reservations.sql database/migrate.sql backend/internal/hotel/integration_test.go && git commit -m "feat: adiciona estrutura de reservas"`.

### Tarefa 2: Definir regras de domínio e incluir tipo de documento no cadastro

**Arquivos:** criar `backend/internal/hotel/reservations.go` e `backend/internal/hotel/reservations_test.go`; modificar `backend/internal/hotel/service.go`, `backend/internal/hotel/guests.go`, `backend/internal/hotel/service_test.go` e `backend/internal/httpapi/hotel.go`.

**Interfaces:** usar literalmente os tipos e métodos da seção “Contratos compartilhados”. Erros de domínio são `ErrReservationConflict`, `ErrInvalidReservation` e os existentes `ErrNotFound`/`ErrConflict`.

- [ ] **Passo 1: Escrever testes unitários para intervalo, noites e total.** Cobrir intervalo válido de duas noites, checkout no mesmo dia, checkout anterior e total de `125.50 × 2 = 251.00`, sem conversão para `float64`.

```go
func TestReservationNightsAndTotal(t *testing.T) {
	checkIn := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	nights, total, err := calculateReservationTotal(checkIn, checkOut, "125.50")
	if err != nil || nights != 2 || total != "251.00" {
		t.Fatalf("got nights=%d total=%q err=%v", nights, total, err)
	}
}
```

- [ ] **Passo 2: Executar o teste unitário antes da implementação.** `cd backend && go test ./internal/hotel -run 'TestReservationNightsAndTotal|TestValidateReservationDates' -v`. Resultado esperado: falha por funções/tipos ausentes.
- [ ] **Passo 3: Implementar tipos, validação e cálculo decimal.** Validar datas civis no formato ISO, checkout posterior, ocupantes positivos, exatamente um identificador de hóspede e tipo RG/CPF para entrada pública. Calcular noites em dias UTC; converter diária com até duas casas para centavos inteiros, multiplicar por noites e formatar o resultado com exatamente duas casas.
- [ ] **Passo 4: Adicionar testes de capacidade e de entradas documentais.** Cobrir documento obrigatório, tipo diferente de RG/CPF, nome/e-mail vazios, ocupante acima da capacidade e total que termina em `.00`/`.50`.
- [ ] **Passo 5: Persistir `documentType` no CRUD autenticado.** Atualizar campos de JSON, scan/insert/update e queries de listagem. Tipo pode ser apagado (`NULL`) para corrigir cadastro, mas reserva pública sempre o exige. Não inferir tipo em cadastro antigo.
- [ ] **Passo 6: Rodar suíte e formatar.** `cd backend && gofmt -w internal/hotel/reservations.go internal/hotel/reservations_test.go internal/hotel/service.go internal/hotel/guests.go internal/hotel/service_test.go internal/httpapi/hotel.go && go test ./internal/hotel ./internal/httpapi`. Resultado esperado: testes aprovados e sem regressão de hóspedes.
- [ ] **Passo 7: Registrar o incremento.** `git add backend/internal/hotel backend/internal/httpapi/hotel.go && git commit -m "feat: define regras de reserva e documento"`.

### Tarefa 3: Implementar disponibilidade e confirmação transacional

**Arquivos:** criar `backend/internal/hotel/reservations_postgres.go`; modificar `backend/internal/hotel/service.go` e `backend/internal/hotel/integration_test.go`.

**Interfaces:** `ListAvailability` e `CreateReservation` implementam os contratos compartilhados. A consulta usa as regras `is_available`, capacidade e sobreposição estrita `existing.check_in < requested.check_out AND existing.check_out > requested.check_in`.

- [ ] **Passo 1: Escrever integração de disponibilidade.** Com banco isolado, criar dois quartos e reservas de referência; verificar quarto bloqueado manualmente, capacidade insuficiente, conflito parcial e intervalo sucessivo no dia do checkout.
- [ ] **Passo 2: Rodar integração para confirmar falha.** `cd backend && TEST_DATABASE_URL="$TEST_DATABASE_URL" go test ./internal/hotel -run TestListAvailability -v`. Resultado esperado: falha até os métodos de disponibilidade existirem.
- [ ] **Passo 3: Implementar consulta de disponibilidade.** Retornar apenas os campos de `AvailabilityOption`, diária atual em string decimal, noites e total previsto calculados pela regra de domínio; ordenar resultados por número do quarto.
- [ ] **Passo 4: Escrever integração de criação e snapshots.** Cobrir hóspede existente no painel, hóspede novo público, documento legado com tipo nulo, e-mail divergente e rejeição de conflito.
- [ ] **Passo 5: Implementar a transação de criação.** Abrir transação pgx; bloquear o quarto com `SELECT ... FOR UPDATE`; carregar capacidade, disponibilidade e diária; rejeitar com conflito se `expectedDailyRate` divergir do preço consultado; revalidar intervalo e conflito depois do lock; associar hóspede por `document_normalized` ou criar um novo; inserir snapshot de nome/tipo/documento/e-mail e preço; confirmar transação. Quando documento existente tiver tipo nulo, preservá-lo. Quando o e-mail divergir, preservar o e-mail principal e usar o informado no snapshot.
- [ ] **Passo 6: Testar duas confirmações concorrentes.** Disparar duas goroutines com quarto e período iguais; exigir exatamente um sucesso e um `ErrReservationConflict`. Repetir com períodos que apenas encostam no checkout e exigir dois sucessos.
- [ ] **Passo 7: Rodar verificações do pacote.** `cd backend && gofmt -w internal/hotel && TEST_DATABASE_URL="$TEST_DATABASE_URL" go test ./internal/hotel -v`. Resultado esperado: testes unitários e PostgreSQL aprovados (testes de integração informam skip apenas quando a variável não foi fornecida).
- [ ] **Passo 8: Registrar o incremento.** `git add backend/internal/hotel && git commit -m "feat: reserva quartos sem conflito"`.

### Tarefa 4: Expor disponibilidade e reserva pela API

**Arquivos:** criar `backend/internal/httpapi/reservations.go` e `backend/internal/httpapi/reservations_test.go`; modificar `backend/internal/httpapi/router.go` e `backend/README.md`.

**Interfaces:** implementar os quatro endpoints da seção “Contratos compartilhados”. Rotas `/public/` não usam `requireAuth`; rotas internas chamam `requireAuth` antes de ler ou processar o corpo.

- [ ] **Passo 1: Escrever testes HTTP de contrato.** Cobrir resposta pública de disponibilidade sem chaves de hóspede, criação válida com HTTP 201, validação com 422, conflito com 409 e endpoint interno sem token com 401.
- [ ] **Passo 2: Rodar os testes para confirmar falha.** `cd backend && go test ./internal/httpapi -run TestReservation -v`. Resultado esperado: rota/teste ainda não implementados.
- [ ] **Passo 3: Implementar decodificação e handlers.** Ler datas `YYYY-MM-DD`, validar query, limitar corpo pelo padrão do pacote, encaminhar ao serviço e serializar os envelopes acordados. Traduzir `ErrReservationConflict` e `ErrConflict` para mensagens seguras; nunca incluir `Guest` na resposta pública.
- [ ] **Passo 4: Registrar as rotas.** Adicionar os dois padrões públicos e os dois padrões internos ao `NewRouter`; manter rotas já existentes protegidas como estão.
- [ ] **Passo 5: Atualizar a documentação da API.** Documentar parâmetros, formatos, status, autenticação e exemplos JSON em `backend/README.md`.
- [ ] **Passo 6: Rodar testes e análise estática.** `cd backend && gofmt -w internal/httpapi && go test ./internal/httpapi ./internal/hotel && go vet ./...`. Resultado esperado: todos os testes aprovados e `go vet` sem diagnósticos.
- [ ] **Passo 7: Registrar o incremento.** `git add backend/internal/httpapi backend/README.md && git commit -m "feat: expõe api de reservas"`.

### Tarefa 5: Implementar reserva pública no site

**Arquivos:** criar `frontend/src/services/reservations.ts`, `frontend/src/pages/ReservationPage.tsx`, `frontend/src/components/ReservationSearchForm.tsx` e `frontend/src/components/ReservationRoomOptions.tsx`; modificar `frontend/src/App.tsx`.

**Interfaces:** os tipos TypeScript correspondem a `AvailabilityOption`, `ReservationGuestInput` e ao payload público. O serviço chama `GET /public/availability` e `POST /public/reservations` por uma nova função `publicRequest` em `api.ts`, sem `authenticatedRequest`.

- [ ] **Passo 1: Definir tipos e serviço público.** Declarar `AvailabilityOption`, `ReservationGuestInput`, `PublicReservationInput` (datas string `YYYY-MM-DD`) e `ReservationConfirmation` sem dados de hóspede; exportar `publicRequest<T>` de `api.ts`, reutilizando base URL, `credentials: 'include'`, JSON e `APIError`; implementar `searchAvailability` e `createPublicReservation` com esse helper.
- [ ] **Passo 2: Montar o formulário de busca.** Reutilizar o componente para check-in, check-out e ocupantes; exigir ocupantes inteiros positivos e check-out posterior.
- [ ] **Passo 3: Montar opções de quarto.** Mostrar categoria, capacidade, diária, diárias e total vindo da API. Não calcular preço oficial no navegador.
- [ ] **Passo 4: Implementar dados, revisão e confirmação.** Após a escolha, solicitar nome, seletor RG/CPF, número e e-mail; exibir resumo com período/quarto/preço antes de enviar; em 409 manter formulário e busca preenchidos e permitir consultar de novo.
- [ ] **Passo 5: Registrar rota pública.** Adicionar `/reservar` sob `Layout`, fora de `ProtectedRoute` e `PublicOnlyRoute`; apontar a rota inicial e a marca do cabeçalho para `/reservar` para que o visitante encontre a busca de quartos.
- [ ] **Passo 6: Validar build e lint.** `cd frontend && npm run build && npm run lint`. Resultado esperado: build e lint sem erros.
- [ ] **Passo 7: Registrar o incremento.** `git add frontend/src && git commit -m "feat: adiciona reserva publica"`.

### Tarefa 6: Implementar reserva pela recepção

**Arquivos:** criar `frontend/src/pages/ReceptionReservationsPage.tsx`; modificar `frontend/src/services/hotel.ts`, `frontend/src/App.tsx`, `frontend/src/components/DashboardLayout.tsx`, `frontend/src/components/EntityDialog.tsx` e `frontend/src/pages/RecordsPage.tsx`.

**Interfaces:** usar os endpoints autenticados `/reservations/availability` e `/reservations`, e a busca interna de hóspedes existente. A criação envia `guestId`, nunca dados públicos do hóspede.

- [ ] **Passo 1: Atualizar tipos e edição de hóspede.** Adicionar `documentType?: 'RG' | 'CPF' | null` ao tipo `Entity`; permitir selecionar tipo ao criar/editar; mostrar tipo e valor na listagem. Cadastros legados permanecem sem seleção até intervenção explícita.
- [ ] **Passo 2: Implementar chamadas autenticadas.** Adicionar `searchReservationAvailability` e `createStaffReservation` em `hotel.ts`, usando `authenticatedRequest` e os mesmos tipos de retorno do serviço público.
- [ ] **Passo 3: Construir seleção de hóspede e busca.** Selecionar hóspede via serviço existente; se faltar nome, tipo, documento ou e-mail requerido, completar o cadastro antes de continuar; consultar disponibilidade e exibir o mesmo resumo de preço.
- [ ] **Passo 4: Construir confirmação operacional.** Enviar `guestId` e os dados de estadia; exibir sucesso, validação e conflito; em conflito preservar hóspede e período para nova consulta.
- [ ] **Passo 5: Conectar navegação protegida.** Adicionar rota `/reservas` dentro de `ProtectedRoute` e habilitar o item “Reservas” em `DashboardLayout` apontando para ela.
- [ ] **Passo 6: Validar build e lint.** `cd frontend && npm run build && npm run lint`. Resultado esperado: build e lint sem erros.
- [ ] **Passo 7: Registrar o incremento.** `git add frontend/src && git commit -m "feat: adiciona reservas da recepcao"`.

### Tarefa 7: Documentar e revisar critérios de aceite

**Arquivos:** modificar `README.md`, `backend/README.md`, `frontend/README.md` e `database/README.md`.

**Interfaces:** documentação descreve os quatro endpoints, os dois fluxos, aplicação da migração 005 e a configuração de integração com `TEST_DATABASE_URL`.

- [ ] **Passo 1: Atualizar README raiz.** Descrever que reserva pública e da recepção estão disponíveis e indicar como iniciar os serviços.
- [ ] **Passo 2: Atualizar README do banco.** Incluir a versão 005 na sequência de migrações, sem instrução de apagar volumes ou dados.
- [ ] **Passo 3: Atualizar READMEs do backend e frontend.** Registrar contratos, autenticação por canal, tratamento de conflito e execução de build/testes.
- [ ] **Passo 4: Rodar verificação final.** `cd backend && go test ./... && go vet ./...`; depois `cd ../frontend && npm run build && npm run lint`; aplicar `database/migrate.sql` num banco descartável com `ON_ERROR_STOP=1`. Resultado esperado: verificações aprovadas.
- [ ] **Passo 5: Conferir cobertura dos critérios de aceite.** Confirmar consulta sem conta, capacidade, preço/noites, snapshot de preço/contato, conflitos parciais e concorrentes, checkout sucessivo, autenticação interna e ausência de pagamento/cancelamento/check-in realizado.
- [ ] **Passo 6: Registrar documentação final.** `git add README.md backend/README.md frontend/README.md database/README.md && git commit -m "docs: documenta fluxo de reservas"`.

## Revisão de cobertura da especificação

| Requisito | Tarefa |
| --- | --- |
| Migração preserva hóspedes e identifica RG/CPF explicitamente | 1 e 2 |
| Consulta pública e interna de quartos por período e ocupação | 3, 4, 5 e 6 |
| Cálculo de noites, diária e preço previsto com snapshot | 2 e 3 |
| Reserva pública sem conta com nome, documento e e-mail | 3, 4 e 5 |
| Reserva da recepção para hóspede existente | 3, 4 e 6 |
| Hóspede legado e e-mail de contato sem alteração indevida do cadastro | 1 e 3 |
| Capacidade, bloqueio manual e intervalo `[check-in, check-out)` | 2 e 3 |
| Proteção transacional contra reservas concorrentes | 3 |
| Rotas administrativas protegidas e resposta pública sem dados privados | 4, 5 e 6 |
| Estados de carregamento, ausência de opções, validação, conflito e sucesso | 5 e 6 |
| Documentação, build, lint e testes | 1 a 7 |

## Checkpoints desta execução

- [x] Migração 005 validada em schemas PostgreSQL temporários, preservando dados legados e rejeitando colisões de normalização.
- [x] Regras de domínio, preço em centavos, capacidade, intervalo de datas e snapshots implementados.
- [x] Confirmação transacional testada com conflito simultâneo e alteração da diária depois da consulta.
- [x] Endpoints públicos e autenticados implementados; resposta pública sem documento/e-mail e rota interna protegida.
- [x] Fluxos público e da recepção implementados; build e lint do front-end aprovados.
- [x] READMEs atualizados; `go test ./... -count=1`, com PostgreSQL, e `go vet ./...` aprovados com `CGO_ENABLED=0`.

O teste de integração aplica as migrações 003 e 005 em schemas temporários. O executor `database/migrate.sql` não foi aplicado ao schema persistente do ambiente.
