# Banco de dados

Os scripts de criação, evolução e consulta do banco de dados ficam nesta pasta.

- `schema`: estrutura inicial das tabelas e relacionamentos.
- `migrations`: alterações versionadas do banco.
- `seeds`: dados de desenvolvimento e demonstração.

## Executar PostgreSQL

Na raiz do projeto, execute:

```bash
docker-compose up -d database
```

As migrações em `migrations/` são aplicadas automaticamente apenas quando o
volume do banco é criado pela primeira vez.

Se você já tiver um volume criado, aplique a migração `002_create_auth_tokens.sql`
manualmente ou recrie apenas o volume de desenvolvimento antes de usar os
endpoints de autenticação.
