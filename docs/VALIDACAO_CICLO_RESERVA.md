# Roteiro de validação — ciclo operacional da reserva

Use os usuários de [teste](USUARIOS_TESTE.md) e uma reserva de desenvolvimento.
O roteiro cobre ações autenticadas da recepção e do administrador; não usa
reservas de produção.

## Fluxo principal

- [ ] Abra uma reserva futura e confira período previsto, diária, total e status.
- [ ] Realize check-in antes da data prevista, sem outra estadia no quarto.
- [ ] Confira status “Em hospedagem”, horário/responsável no detalhe e no histórico,
      além do quarto “Ocupado”.
- [ ] Confirme que nova reserva com noites sobrepostas à estadia efetiva não é
      oferecida na disponibilidade e que a API recusa a tentativa concorrente.
- [ ] Confirme que uma reserva com período posterior ao checkout previsto ainda
      pode ser criada para o quarto ocupado.
- [ ] Faça checkout e confira a cotação do servidor: noites, diária contratada,
      timestamp e valor previsto.
- [ ] Confirme checkout; verifique status “Finalizada”, valor final persistido,
      horário/responsável no detalhe/histórico e quarto liberado.

## Regras e conflitos

- [ ] Tente check-in em reserva cancelada, já iniciada ou após o checkout previsto;
      a operação deve ser recusada sem alterar reserva/quarto.
- [ ] Tente check-in com o quarto ocupado por outra reserva; receba conflito e
      confira que a segunda reserva continua sem check-in.
- [ ] Tente check-in ou checkout com versão obsoleta e repita cada operação já
      concluída; receba conflito sem evento duplicado.
- [ ] Faça checkout no mesmo dia do check-in; o valor deve ser R$ 0,00.
- [ ] Faça checkout em data posterior; o total deve ser diária salva na reserva
      multiplicada pelas noites civis efetivas em `America/Sao_Paulo`.
- [ ] Altere a diária atual do cadastro do quarto entre check-in e checkout;
      a reserva deve continuar usando a diária contratada.
- [ ] Tente cancelar uma reserva em hospedagem; o cancelamento deve ser recusado.
- [ ] Confira reservas legadas: status previsto e total previsto permanecem
      disponíveis, com dados realizados e valor final vazios.

## Migração e concorrência

- [ ] Em banco descartável, aplique 001–008 em sequência e reexecute o executor;
      a segunda execução não deve reaplicar migrações.
- [ ] Em banco descartável com reserva anterior à 008, aplique a migração e
      confirme preservação da reserva e do quarto desocupado.
- [ ] Force falha na gravação do evento de check-in e checkout; reserva e quarto
      devem permanecer no estado anterior.
- [ ] Dispare duas ações simultâneas para a mesma reserva/quarto; somente uma
      transição deve ser efetivada.

## Interface

- [ ] Faça o fluxo em desktop e largura móvel com os perfis administrador e
      recepção.
- [ ] Confira erros, carregamento da cotação, atualização após sucesso e recarga
      após conflito/versionamento obsoleto.
- [ ] Confirme que “Ocupado” aparece separado de “Disponível para venda” e que
      a tela de quartos não oferece uma ação manual para liberar ocupação real.
