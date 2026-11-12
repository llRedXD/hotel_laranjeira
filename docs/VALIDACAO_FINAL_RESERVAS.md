# Checkpoint 5 — validação final da gestão de reservas

Verificação realizada em 11/11/2026, no fuso de São Paulo. **Validação técnica
aprovada; aceite final do responsável pendente.** Os testes usaram PostgreSQL 17,
API e frontend temporários; as reservas do ambiente local foram preservadas.

## Resultado por funcionalidade

| Funcionalidade | Resultado conferido |
| --- | --- |
| Consultar reservas | Listagem autenticada, contagem, paginação, erro, carregamento e recuperação |
| Filtrar por hóspede | Busca por nome/documento e histórico exato por UUID, com homônimos separados nos testes PostgreSQL |
| Filtrar por período | Estadias iniciadas antes do intervalo aparecem; noite do checkout fica fora; canceladas mantêm o período reservado |
| Cancelar | Confirmação, versão incrementada, bloqueio de novas ações, quarto liberado e reserva preservada |
| Alterar | Três etapas por tipo, revisão antes/depois, diária contratada ou nova diária conforme quarto, autoria na auditoria |
| Visualizar status | Cinco status conferidos na API e tela: confirmada, em andamento, saída hoje, encerrada e cancelada |
| Histórico do hóspede | Acesso pelo cadastro e nome na reserva; ativa e cancelada permanecem no mesmo histórico |

Auditoria adicional conferida pelo menu: origem pública sem autor funcionário,
alterações/cancelamento com funcionário e horário, estados anteriores/posteriores
e registros iniciais de reservas legadas. Datas e status são previstos; check-in
e checkout realizados, pagamentos, reembolsos e reativação continuam fora da entrega.

## Fluxo completo executado no navegador

1. Reserva pública sem login para `Teste Integrado CP5`: quarto 101, 01–03/12/2026,
   duas pessoas, diária de R$ 280,00 e total de **R$ 560,00**.
2. Administrador localizou a reserva pelo hóspede e pela noite de 02/12, abriu
   o histórico e alterou a saída para 04/12: **R$ 840,00**, mantendo a diária.
3. Alterou pessoas de duas para três; a capacidade exigiu uma escolha explícita
   de outro quarto. Quarto 301, diária de R$ 560,00: **R$ 1.680,00**.
4. Recepção trocou para o quarto 302, diária de R$ 620,00: **R$ 1.860,00**.
5. Recepção cancelou. A reserva permaneceu no histórico, com ações bloqueadas;
   a disponibilidade pública mostrou o quarto liberado.
6. Recepção criou outra reserva para o mesmo hóspede, quarto e período liberado:
   **R$ 1.860,00**. O histórico mostrou uma cancelada e outra confirmada.
7. Auditoria e banco confirmaram **seis eventos**: criação pública, três edições,
   cancelamento e criação interna, com origens/autores corretos. As duas reservas
   mantiveram o mesmo identificador de hóspede. Nenhuma sobreposição ativa foi encontrada.

Também foram conferidos: ausência de acesso à auditoria sem login, comparação
antes/depois, filtros preservados, intervalos inválidos, resultado vazio,
carregamento e recuperação de uma indisponibilidade 503 simulada no navegador.
As funcionalidades foram exercitadas em desktop e celular com os dois perfis.

Evidências locais:

- [Confirmação pública](evidencias/reservas/confirmacao-publica.png)
- [Auditoria no celular](evidencias/reservas/auditoria-mobile.png)
- [Histórico no celular](evidencias/reservas/historico-mobile.png)

## Correção encontrada na validação

Na criação pela recepção, trocar datas depois de selecionar o quarto mantinha a
cotação anterior visível e confirmável. Corrigido: mudar período ou ocupantes
limpa os quartos, a seleção e o total; uma nova consulta passa a ser obrigatória.
Os campos de período/pessoas ficam bloqueados durante consulta/confirmação nos
dois canais para impedir respostas de disponibilidade associadas a dados antigos.

A reprodução foi feita sem confirmar a cotação incorreta. Após a correção,
o navegador confirmou a invalidação, o bloqueio durante resposta atrasada e o
recálculo correto para duas, três e quatro diárias.

## Verificações automatizadas e migrações

- `CGO_ENABLED=0 go test ./... -count=1 -v`, com `TEST_DATABASE_URL` definido:
  todos aprovados, incluindo integração, sem SKIP.
- `CGO_ENABLED=0 go vet ./...`: aprovado.
- `npm run lint` e `npm run build`: aprovados após a correção da recepção.
- `git diff --check`: aprovado.
- Executor de migrações 001–007 em banco novo e reexecução: aprovado.
- Upgrade 005 → 007 com reserva existente: snapshot, diária, total e reserva
  preservados; versão inicial 1 e um evento de importação, sem duplicação ao repetir.
- Concorrência: duas criações, duas edições, edição/criação, edição/cancelamento,
  cancelamentos simultâneos e movimentos opostos entre quartos.
- Auditoria: falhas induzidas desfazem cancelamento/edição; autoria preservada
  mesmo após renomeação do funcionário; nenhuma gravação ao apenas navegar no fluxo.

Para reproduzir os testes Go, configure `TEST_DATABASE_URL` e execute em `backend`.
Eles criam esquemas exclusivos e os removem ao terminar. Os testes de navegador
foram executados pela CLI Playwright; as evidências não equivalem a uma suíte de
navegador adicionada ao CI.

## Aceite final

No ambiente local, conferir e registrar o aceite:

- [ ] Consultar e combinar filtros de hóspede/período.
- [ ] Abrir o histórico pelo cadastro do hóspede.
- [ ] Alterar uma reserva futura pelo fluxo de etapas e conferir valores/auditoria.
- [ ] Cancelar uma reserva futura e conferir a liberação do quarto e o histórico.
- [ ] Conferir os status e o bloqueio de alterações em reservas iniciadas/canceladas.
- [ ] Confirmar que a auditoria permanece acessível pelo menu.

Roteiros específicos: [checkpoint 2](VALIDACAO_CHECKPOINT2.md),
[checkpoint 3](VALIDACAO_CHECKPOINT3.md) e [checkpoint 4](VALIDACAO_CHECKPOINT4.md).
