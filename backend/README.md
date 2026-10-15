# Back-end

API em Go para as regras de negócio e acesso ao PostgreSQL.

## Executar com Docker

Na raiz do repositório:

```bash
docker compose up --build backend
```

O endpoint de saúde estará em `http://localhost:8080/api/v1/health`.

## Organização

- `cmd/api`: ponto de entrada da aplicação.
- `internal/config`: leitura de configurações.
- `internal/database`: conexão com PostgreSQL.
- `internal/httpapi`: handlers HTTP.

## Autenticação

Configure `JWT_SECRET` e, para o primeiro acesso, `INITIAL_ADMIN_*` e
`INITIAL_RECEPTION_*`. As contas são criadas somente em ambiente de
desenvolvimento (`APP_ENV=development` ou `dev`) e quando seus e-mails ainda
não existem. Em produção, o provisionamento de usuários deve ocorrer por um
processo administrativo seguro. O front-end usa o access token retornado pelo
login e um refresh token em cookie HttpOnly.

Endpoints disponíveis:

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/password-recovery`
- `POST /api/v1/auth/password-reset`

Em desenvolvimento, o link de recuperação é escrito no log da API. Em
produção, configure um provedor de e-mail antes de expor esse fluxo.

## Contratos HTTP

Envie JSON com Content-Type: application/json. Limite de corpo: 64 KiB.
Erros usam `{"error":{"code":"...","message":"...","fields":{}}}`;
fields aparece quando há erros por campo.

| Método e caminho | Entrada | Saída |
| --- | --- | --- |
| POST /api/v1/auth/login | email, password | user, accessToken e cookie |
| POST /api/v1/auth/refresh | cookie | user, accessToken e novo cookie |
| POST /api/v1/auth/logout | cookie | 204 |
| GET /api/v1/auth/me | Bearer | user |
| POST /api/v1/auth/password-recovery | email | 202, mensagem genérica |
| POST /api/v1/auth/password-reset | token, password | 204 |
| POST /api/v1/guests | name; documentType, document, email, phone, address opcionais | 201, guest |
| GET /api/v1/guests | search, limit, offset | guests |
| GET /api/v1/guests/{id} | UUID | guest |
| PATCH /api/v1/guests/{id} | campos a alterar | guest |
| POST /api/v1/room-categories | name, description opcional | 201, category |
| GET /api/v1/room-categories | limit, offset | categories |
| PATCH /api/v1/room-categories/{id} | campos a alterar | category |

Hóspedes e categorias exigem Bearer JWT. Ambos os papéis podem acessá-los.
Listas usam limit=50 por padrão (1–100), offset=0 (0–100000), ordenação por
nome e id. Continue com offset maior até receber uma página menor que limit.
PATCH preserva campos omitidos. null ou string vazia limpa campos opcionais.
name não aceita null, vazio ou espaços. Um objeto vazio recebe 422; campos
desconhecidos e tipos diferentes de string/null recebem 400. As atualizações
ocorrem em uma operação SQL, preservando as colunas não enviadas.

Exemplo: `{"phone":null}` limpa somente o telefone de um hóspede;
`{"name":"Suíte"}` renomeia a categoria preservando sua descrição.
Strings são aparadas; e-mail é normalizado para minúsculas. O tipo do documento
aceita RG ou CPF e pode permanecer nulo em cadastros antigos até classificação
explícita. Limites contam caracteres Unicode: nome de hóspede 120, categoria 80,
documento/telefone 30, e-mail 255 e endereço/descrição 2000. Documento continua
opcional nos cadastros internos; a reserva pública exige tipo e número, sem
validação formal de RG/CPF.

### Reservas

| Método e caminho | Acesso | Entrada | Saída |
| --- | --- | --- | --- |
| GET `/api/v1/public/availability` | público | `checkIn`, `checkOut` (`YYYY-MM-DD`), `occupants` | `rooms` com quarto, categoria, capacidade, diária, noites e total estimado |
| POST `/api/v1/public/reservations` | público | `roomId`, datas, `occupants`, `expectedDailyRate`, `guest` (`name`, `documentType`, `document`, `email`) | 201, confirmação sem documento nem e-mail |
| GET `/api/v1/reservations/availability` | Bearer JWT | mesmos filtros públicos | `rooms` com disponibilidade e preço estimado |
| POST `/api/v1/reservations` | Bearer JWT | `roomId`, datas, `occupants`, `expectedDailyRate`, `guestId` | 201, reserva operacional |

O intervalo é `[checkIn, checkOut)`: uma nova reserva pode começar no dia do
checkout anterior. A API calcula o total em centavos inteiros, verifica
capacidade, disponibilidade e diária novamente durante a confirmação; se a
diária mudou desde a consulta, é preciso consultar novamente. A reserva persiste
a diária/total aplicados. Conflito de quarto, preço ou duplicidade de hóspede retorna
409; entrada inválida retorna 422. As rotas públicas não dão acesso às rotas
administrativas de hóspedes/quartos e a resposta pública não revela o cadastro
associado ao documento informado.

## Serviços e repositórios de hotel

Handlers decodificam HTTP e traduzem erros; hotel.Service valida/normaliza
entradas e depende da interface Repository. PostgresRepository concentra SQL e
tradução de conflitos.

## Serviços e repositórios de autenticação

auth.Service cuida de bcrypt, JWT, geração/hash de tokens, prazos e regras de
acesso. auth.Repository define as operações sem depender de pgx.
auth.PostgresRepository concentra SQL, tradução de ausência de usuário e
transações. O bootstrap de DEV também usa esse contrato.

RotateRefresh revoga o token anterior e insere o sucessor na mesma transação.
ConsumeReset troca a senha, incrementa auth_version, consome o token e revoga
refresh tokens atomicamente. ReplaceReset substitui links anteriores em uma
transação com bloqueio por usuário. O repositório recebe hashes, nunca tokens
brutos ou senhas em texto puro. Os endpoints e respostas HTTP foram preservados.

## Testes de integração

Com PostgreSQL local ativo e credenciais de desenvolvimento:

```bash
TEST_DATABASE_URL='postgres://hotel_laranjeira:hotel_laranjeira_dev@localhost:5432/hotel_laranjeira?sslmode=disable' CGO_ENABLED=0 go test ./internal/hotel -v -count=1
```

Os testes de hotel usam esquema exclusivo, aplicam as migrações 003 e 005 e
fazem rollback/limpeza ao terminar. O usuário precisa de permissão CREATE SCHEMA.
Eles cobrem preservação de documentos legados, cálculo, conflito de período,
checkout sucessivo e duas confirmações simultâneas. A migração versionada
completa inclui também a versão 004 de autenticação.
Sem TEST_DATABASE_URL, os testes PostgreSQL são explicitamente pulados. Testes
de regras, contratos HTTP, autenticação e resposta pública executam com
`go test ./...`.

Para incluir autenticação, use a mesma variável e execute:
`CGO_ENABLED=0 go test ./internal/auth ./internal/hotel -v -count=1`.
O teste de auth aplica 001, 002 e 004 em esquema temporário, com rollback ao
terminar. Cobre login, rotação, rollback em conflito de hash, reutilização,
logout, redefinição de senha, revogação e usuário inativo.

Respostas: 400 JSON inválido, 401 autenticação, 403 origem recusada,
404 inexistente, 409 duplicidade, 415 Content-Type, 422 validação,
429 limite de tentativas e 500 erro interno.

Access token dura 15 minutos. Refresh dura 30 dias e é rotacionado em transação.
Troca de senha invalida access tokens por versão no banco e revoga refresh tokens.
Logout revoga o refresh atual; um access token já emitido pode durar até expirar.
As rotas protegidas consultam o usuário ativo no banco.

Limite de autenticação: 30 POSTs/minuto por IP do socket, por processo.
Não há confiança automática em X-Forwarded-For; atrás de proxy, configure
limitação adicional no proxy. Cookies exigem implantação same-site; fora de DEV,
origem HTTPS e segredo de pelo menos 32 bytes são obrigatórios.
