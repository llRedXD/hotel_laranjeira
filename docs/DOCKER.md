# Ambiente Docker de desenvolvimento

Requer Docker Engine e Compose com suporte a service_completed_successfully.
Se o plugin não estiver instalado, substitua docker compose por docker-compose.
Na raiz, copie .env.example para .env somente se ainda não houver configuração.

```bash
docker compose up -d --build
docker compose ps -a
docker compose logs --tail=100 backend migrate
```

Acesse http://localhost:5173. O Vite encaminha /api ao backend pela rede interna.
Banco e API também ficam disponíveis em localhost:5432 e localhost:8080.
Todas as portas publicadas são limitadas a 127.0.0.1.

Se as portas padrão estiverem ocupadas, use:

```bash
BACKEND_PORT=8081 FRONTEND_PORT=5174 FRONTEND_ORIGIN=http://localhost:5174 docker compose up -d --build
```

Nesse caso, acesse http://localhost:5174. Para manter a escolha nas próximas
execuções, configure esses três valores no .env.

Ordem: PostgreSQL saudável → migrate concluído → API saudável → frontend.
migrate terminar com código 0 é normal. O serviço aplica scripts em transação,
com lock e histórico em schema_migrations; reiniciar não reaplica versões.
Para adicionar migrações, inclua também um bloco versionado em database/migrate.sql.

## Volume criado pelo ambiente antigo

O antigo entrypoint aplicava SQL sem registrar histórico. Se já existe esse
volume, o primeiro migrate pode falhar com "relation already exists".
Após conferir que as tabelas correspondem às migrações originais:

```bash
docker compose run --rm -e ADOPT_LEGACY_SCHEMA=true migrate
docker compose up -d
```

A adoção registra versões conhecidas por presença de tabelas/coluna e aplica
as restantes; não é comparação completa de schema. Use somente para o esquema
original deste repositório. Não remova o volume para resolver migrações.

## Acessos e configuração

Somente APP_ENV=development ou dev permite bootstrap. Credenciais de exemplo:
admin@hotel.com / admin12345678 e recepcao@hotel.com / recepcao123456.
Contas existentes não são sobrescritas; valores do .env prevalecem.

Consulte também a página de [usuários de teste](USUARIOS_TESTE.md), com os
perfis, as credenciais padrão e as instruções de acesso para avaliação.

FRONTEND_PORT, BACKEND_PORT e POSTGRES_PORT alteram portas publicadas.
Se FRONTEND_ORIGIN estiver no .env, ajuste-o para a porta de FRONTEND_PORT.
Credenciais de um volume PostgreSQL existente não mudam ao editar o .env.
O frontend recebe VITE_API_URL=/api/v1 e usa API_PROXY_TARGET=http://backend:8080.

## Operação

```bash
docker compose logs -f backend frontend
docker compose down
docker compose up -d --build
```

down preserva postgres_data. Não use down -v se deseja preservar dados.
O front-end tem recarga automática via bind mount; mudanças no Go exigem rebuild.
Mudanças em package-lock.json podem exigir renovar o volume anônimo de node_modules:
`docker compose up -d --build --renew-anon-volumes frontend`.

Este ambiente usa Vite e credenciais de desenvolvimento; produção exige
publicação estática, HTTPS e configuração própria.
