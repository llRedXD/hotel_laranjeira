# Hotel Laranjeira

Sistema de gestão hoteleira desenvolvido para a disciplina de Projeto de Software. O projeto segue uma arquitetura em três camadas: front-end, back-end e banco de dados.

## Estrutura do projeto

```
hotel_laranjeira/
├── frontend/       # Interface da aplicação
├── backend/        # API e regras de negócio
├── database/       # Scripts e estrutura do banco de dados
├── .env.example    # Modelo de variáveis de ambiente
└── README.md
```

## Front-end

O front-end utiliza React, TypeScript, Vite e Tailwind CSS. O hóspede pode
consultar e reservar em `/reservar` sem criar conta; a equipe acessa o painel
autenticado, incluindo reservas da recepção. A rota inicial abre a busca pública
de quartos; o painel e os cadastros continuam protegidos por autenticação.

### Executar localmente

```bash
cd frontend
npm ci
npm run dev
```

Abra `http://localhost:5173` no navegador.

### Executar com Docker

```bash
docker compose up --build frontend
```

O serviço fica disponível em `http://localhost:5173` (ou na porta definida em
`FRONTEND_PORT` no arquivo `.env`).

## Back-end e banco de dados

Guia completo de inicialização, volumes existentes e diagnóstico:
[Ambiente Docker](docs/DOCKER.md).

Copie `.env.example` para `.env` caso queira alterar as portas ou credenciais
de desenvolvimento. Para iniciar a API e o PostgreSQL:

```bash
docker compose up --build
```

- Front-end: `http://localhost:5173`
- API: `http://localhost:8080/api/v1/health`
- PostgreSQL: `localhost:5432`

Os usuários iniciais de administrador e recepção são configurados por
`INITIAL_ADMIN_*` e `INITIAL_RECEPTION_*` no `.env`, mas só são criados
em desenvolvimento. Ao rodar a API fora do Docker, exporte essas variáveis
antes de executar `CGO_ENABLED=0 go run ./cmd/api`.

Para acessar o sistema na avaliação, consulte os
[usuários de teste e instruções de acesso](docs/USUARIOS_TESTE.md).

## Estado atual

- Autenticação e painel React; métricas do painel são demonstrativas.
- API de hóspedes e categorias com cadastro, consulta e edição.
- API e telas operacionais de hóspedes, categorias e quartos.
- Reservas públicas e da recepção, com cálculo de diárias, total previsto e proteção contra conflitos.
- Pagamentos e registro do check-in/check-out realizado continuam fora desta entrega.

## API sem Docker

Requer Go 1.24 ou superior e PostgreSQL com as migrações aplicadas.
A API não carrega arquivos .env automaticamente. Na raiz, em bash/zsh:

```bash
set -a
source .env
set +a
cd backend
CGO_ENABLED=0 go run ./cmd/api
```

Copie .env.example para .env apenas se ainda não tiver um arquivo local.
Veja [migrações](database/README.md), [contratos HTTP](backend/README.md)
e [revisão técnica](docs/REVISAO_TECNICA.md).

## Verificação

```bash
cd backend
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go vet ./...
cd ../frontend
npm ci
npm run lint
npm run build
```

O Compose e o servidor Vite fornecidos são para desenvolvimento.
Produção exige HTTPS, segredo JWT exclusivo, provisionamento administrativo,
serviço de e-mail e publicação dos arquivos compilados do front-end.

## Próximos passos

1. Ampliar a gestão de reservas com cancelamento e histórico, conforme decisões futuras.
2. Avaliar pagamentos e registro do check-in/check-out realizado em entregas separadas.
