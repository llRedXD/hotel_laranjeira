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
páginas públicas inicial, de login e de cadastro.

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

## Próximos passos

1. Modelar entidades como hóspedes, quartos, reservas e pagamentos.
2. Criar a API e integrar a autenticação.
3. Criar as telas operacionais de reservas.
