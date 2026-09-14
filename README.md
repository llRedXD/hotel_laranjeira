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

O front-end utiliza React, TypeScript, Vite e Tailwind CSS. Ele contém as
páginas públicas inicial, de login e de recuperação de senha. O acesso à área
interna é destinado à equipe autorizada; contas não são criadas publicamente.

### Executar localmente

```bash
cd frontend
npm install
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

Copie `.env.example` para `.env` caso queira alterar as portas ou credenciais
de desenvolvimento. Para iniciar a API e o PostgreSQL:

```bash
docker-compose up --build
```

- Front-end: `http://localhost:5173`
- API: `http://localhost:8080/api/v1/health`
- PostgreSQL: `localhost:5432`

Os usuários iniciais de administrador e recepção são configurados por
`INITIAL_ADMIN_*` e `INITIAL_RECEPTION_*` no `.env`, mas só são criados
em desenvolvimento. Ao rodar a API fora do Docker, exporte essas variáveis
antes de executar `CGO_ENABLED=0 go run ./cmd/api`.

## Próximos passos

1. Modelar entidades como hóspedes, quartos, reservas e pagamentos.
2. Criar as telas operacionais de reservas.
