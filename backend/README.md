# Back-end

API em Go para as regras de negócio e acesso ao PostgreSQL.

## Executar com Docker

Na raiz do repositório:

```bash
docker-compose up --build backend
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
