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
| 005 | tipo e normalização de documento; reservas, snapshots e índices por quarto/período |

Para banco existente, aplique as versões pendentes pelo executor versionado (na raiz; credenciais padrão de DEV):

```bash
docker compose exec -T database psql -U hotel_laranjeira -d hotel_laranjeira -v ON_ERROR_STOP=1 < database/migrate.sql
```

O executor registra versões e aplica apenas as que faltam. A migração 005
preserva o texto de documentos existentes e interrompe com erro se encontrar
colisões após a normalização; resolva os registros duplicados antes de tentar
novamente. Faça backup antes de evoluir dados importantes. As migrações não
removem dados. Não é necessário excluir volumes.
Em banco novo, o serviço migrate aplica todos os arquivos automaticamente.
