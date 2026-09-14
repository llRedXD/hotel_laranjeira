# Usuários de teste

Estas contas permitem ao professor e à equipe acessar o sistema para avaliação
e testes locais. O Docker Compose utiliza `APP_ENV=development` por padrão.

## Iniciar e acessar

Com Docker e Docker Compose instalados, execute na raiz do projeto:

```bash
docker compose up --build
```

Após os serviços iniciarem, acesse <http://localhost:5173/login> e use uma das
contas abaixo. Se você alterou `FRONTEND_PORT`, ajuste a porta no endereço.

| Perfil | E-mail | Senha |
| --- | --- | --- |
| Administrador | `admin@hotel.com` | `admin12345678` |
| Recepção | `recepcao@hotel.com` | `recepcao123456` |

## Criação das contas e configuração

- As contas são criadas automaticamente na inicialização da API apenas em
  desenvolvimento (`APP_ENV=development` ou `dev`).
- Cada conta só é criada se seu e-mail ainda não existir no banco.
- Os valores acima são os padrões do [Docker Compose](../docker-compose.yml)
  e do [.env.example](../.env.example). Configurações personalizadas podem
  substituir esses valores.
- Para personalizar as contas, copie `.env.example` para `.env` caso ainda não
  tenha esse arquivo e ajuste `INITIAL_ADMIN_NAME`, `INITIAL_ADMIN_EMAIL`,
  `INITIAL_ADMIN_PASSWORD`, `INITIAL_RECEPTION_NAME`, `INITIAL_RECEPTION_EMAIL`
  e `INITIAL_RECEPTION_PASSWORD` antes da primeira inicialização.
- Alterar essas variáveis ou reiniciar os serviços não atualiza a senha de
  uma conta já existente. Se a senha foi alterada, use a senha atual.

Ao executar a API fora do Docker, é necessário exportar as variáveis do `.env`;
a API não carrega esse arquivo automaticamente. Consulte as instruções no
[README principal](../README.md#api-sem-docker).

Estas credenciais são destinadas ao ambiente de desenvolvimento e avaliação.
Para detalhes de inicialização e diagnóstico, consulte o
[guia Docker](DOCKER.md).
