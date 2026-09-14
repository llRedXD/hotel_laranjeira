# Revisão técnica

Revisão consolidada após a separação de serviços e repositórios de hotel e
autenticação. Descreve o código atual; não pressupõe que imagens Docker
construídas anteriormente já contenham as últimas alterações.

## Correções aplicadas

- Configuração validada no início: ambiente, porta, origem e segredo JWT.
- Cookies HttpOnly, rejeição de origem não autorizada e respostas sem cache.
- Limite de tentativas em autenticação e timeouts HTTP/consultas.
- Refresh rotativo transacional; erros de banco não tratados como senha incorreta.
- Usuário ativo e versão de autenticação conferidos nas rotas protegidas.
- Nova migração 004 para invalidar access tokens após troca de senha.
- Validação do cabeçalho JWT, JSON, e-mail e tamanhos dos campos de hóspedes.
- Listagens paginadas e conflitos identificados pelo SQLSTATE PostgreSQL.
- Cliente conserva access token em memória e compartilha refresh concorrente.
- Docker considera go.sum na resolução de dependências.
- Serviços de hotel e autenticação dependem de interfaces de repositório.
- PATCH parcial de hóspedes/categorias preserva colunas omitidas em uma única
  operação SQL, sem carregar e regravar o registro inteiro.
- Bootstrap de usuários de DEV usa o repositório de autenticação.
- Substituição de links de recuperação é transacional e bloqueia o usuário
  durante a operação.

## Organização atual

| Componente | Responsabilidade |
| --- | --- |
| httpapi | Decodificação HTTP, autenticação das rotas e tradução de erros |
| hotel.Service | Validação e normalização de hóspedes/categorias |
| auth.Service | Regras de autenticação, bcrypt, JWT e geração/hash de tokens |
| Repository de cada módulo | Contrato de persistência injetado no serviço |
| PostgresRepository de cada módulo | Consultas SQL e operações transacionais |
| database | Configuração e gerenciamento do pool PostgreSQL |

PATCH preserva campos omitidos, aceita null para limpar opcionais e rejeita
remoção de nome. Strings opcionais vazias também limpam o campo; objetos
vazios recebem 422. Tipos incorretos e campos desconhecidos recebem 400.

Na autenticação, RotateRefresh revoga o refresh anterior e insere o sucessor
atomicamente. ConsumeReset altera o hash da senha, incrementa auth_version,
consome o token de recuperação e revoga refresh tokens na mesma transação.
O repositório recebe hashes, não senhas nem tokens brutos.

## Docker e migrações

- Inicialização ordenada: banco saudável, migrações concluídas, API saudável e
  frontend.
- Migrações registradas em schema_migrations, executadas em transação e
  serializadas por advisory lock.
- Volumes antigos exigem adoção explícita do histórico; a adoção verifica
  presença de objetos conhecidos, não equivalência completa de esquema.
- Frontend usa proxy para a API na rede interna; portas publicadas ficam
  limitadas a 127.0.0.1.
- Novas migrações exigem novo script e registro em database/migrate.sql.

Instruções de operação em [DOCKER.md](DOCKER.md).

## Validação realizada

- Testes Go e go vet passaram após a refatoração de autenticação.
- Integração de hotel com PostgreSQL: preservação de campos omitidos, limpeza
  com null, rejeição de nome nulo/PATCH vazio, inexistência e conflito de categoria.
- Integração de autenticação com PostgreSQL: login, senha incorreta, conta
  ausente, rotação, reutilização, logout, redefinição, invalidação de sessões e
  usuário inativo. Uma falha por hash duplicado confirma rollback da rotação.
- Testes HTTP verificam presença/ausência/null em PATCH, JSON inválido,
  rejeição de origem, paginação e limite de tentativas.
- Testes de integração usam esquemas temporários em transações com rollback;
  sem TEST_DATABASE_URL, são explicitamente pulados.
- Lint/build React e fluxo frontend → API → PostgreSQL via Docker foram
  verificados na preparação do ambiente, antes da última refatoração Go.

Comandos de reprodução em [backend/README.md](../backend/README.md).
Essa cobertura não substitui testes concorrentes com conexões independentes,
testes de carga nem testes ponta a ponta no navegador.

## Desempenho e regras pendentes

Documento é opcional e não tem verificação de CPF/passaporte.
Listas possuem limites; busca ILIKE por substring ainda pode fazer varredura.
Índices especializados devem ser introduzidos com medições representativas.
Não há benchmark que demonstre ganho de latência com estas refatorações.
As rotas protegidas consultam o usuário ativo e a versão de autenticação no
banco; isso permite revogação após troca de senha, com custo de consulta.

## Limitações conhecidas

- Logout não revoga imediatamente access tokens já emitidos.
- Rotação não implementa famílias de tokens/detecção de reutilização.
- Renovações em abas distintas ainda podem disputar o mesmo cookie.
- Limitação de tentativas é local ao processo, não distribuída.
- Recuperação em produção não envia e-mail; integração SMTP está pendente.
- Ainda faltam limpeza agendada de tokens e pipeline CI.
- Migrações não possuem verificação de checksum nem comparação automática
  entre esquema existente e esperado.
- Concorrência entre recuperação, redefinição e refresh ainda precisa de testes
  com conexões independentes, inclusive comportamento em deadlock e retries.
- Front-end Docker usa Vite de desenvolvimento; implantação de produção,
  TLS, backups, observabilidade e gestão de segredos são externos.

Esta revisão não equivale a auditoria de dependências nem certificação de produção.
