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
em `/reservas/nova`, dentro do painel autenticado. A consulta das reservas fica em
`/reservas`, também no painel. Ambos os formulários consultam a API para obter
diárias, noites e total previsto; o servidor é a fonte do preço e revalida
conflitos ao confirmar.

A gestão interna permite busca textual, filtro por noites ocupadas e histórico
exato por hóspede, acessível pelo cadastro em `/hospedes`. A listagem usa status
e permissões do servidor. Cancelamento exige confirmação e mantém o histórico.
A edição em `/reservas/:id/editar` tem três etapas: escolher datas/pessoas/quarto,
ajustar apenas os campos correspondentes e revisar o antes/depois. A gravação
acontece somente na confirmação final. A auditoria fica no menu `/auditoria`,
com responsável, horário e estados registrados.

Alterar datas ou ocupantes invalida quartos e valores consultados nos canais
público e interno. Esses campos ficam bloqueados durante a consulta/confirmação
para impedir que uma resposta de disponibilidade seja usada com outro período.
O mesmo quarto preserva a diária contratada na edição; trocar de quarto usa a
diária atual da nova acomodação. O backend repete a validação ao gravar.

O relatório de [validação final de reservas](../docs/VALIDACAO_FINAL_RESERVAS.md)
reúne os cenários conferidos e o roteiro de aceite das sete funcionalidades.
