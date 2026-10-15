# Front-end

Interface pública de reserva de quartos e painel autenticado para a recepção,
além de login, recuperação de acesso e gestão de cadastros.

## Tecnologias

- React + TypeScript
- Vite
- Tailwind CSS
- React Router

## Comandos

```bash
npm ci
npm run dev
npm run build
npm run lint
```

## Organização

- `src/pages`: páginas e fluxos principais.
- `src/components`: componentes visuais reutilizáveis.
- `src/services`: comunicação com a API.
- `src/styles`: estilos globais e temas.

VITE_API_URL configura a base da API (padrão http://localhost:8080/api/v1).
Para desenvolvimento local, defina em frontend/.env.local. Variáveis VITE são
públicas e incorporadas na build; nunca coloque segredos nelas.
FRONTEND_ORIGIN no back-end deve corresponder à origem do navegador.
Use authenticatedRequest em serviços protegidos: mantém access token em memória
e tenta uma renovação ao receber 401. Renovações concorrentes na mesma aba
compartilham uma chamada. Cookie refresh é HttpOnly.

O fluxo público está em `/reservar` e não exige conta. A recepção cria reservas
em `/reservas`, dentro do painel autenticado. Ambos consultam a API para obter
diárias, noites e total previsto; o servidor é a fonte do preço e revalida
conflitos ao confirmar.
