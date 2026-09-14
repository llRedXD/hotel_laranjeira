# Banco de dados

Os scripts de criação, evolução e consulta do banco de dados ficam nesta pasta.

- `migrations`: alterações versionadas do banco.

## Executar PostgreSQL

Na raiz do projeto, execute:

```bash
docker-compose up -d database
```

O serviço Compose migrate aplica as migrações antes da API, inclusive em
reinicializações, registrando versões em schema_migrations.

Em volumes existentes, aplique apenas os arquivos ainda não executados, em ordem.
No Docker, prefira o executor database/migrate.sql; para volumes antigos,
consulte a adoção de histórico em [DOCKER.md](../docs/DOCKER.md).

| Migração | Conteúdo |
| --- | --- |
| 001 | usuários |
| 002 | refresh tokens e recuperação |
| 003 | hóspedes, categorias e quartos |
| 004 | versão de autenticação para revogação após troca de senha |

Exemplo para banco que já tem 001–003 (na raiz; credenciais padrão de DEV):

```bash
docker compose exec -T database psql -U hotel_laranjeira -d hotel_laranjeira -v ON_ERROR_STOP=1 --single-transaction < database/migrations/004_auth_version.sql
```

Execute 003 antes de 004 se ainda não estiver aplicada, trocando o caminho.
Não reaplique arquivos já executados. Faça backup antes de evoluir dados importantes.
As migrações não removem dados. Não é necessário excluir volumes.
Em banco novo, o serviço migrate aplica todos os arquivos automaticamente.
