# Checkpoint 3 — cancelamento e auditoria

A migração 007 e a API atualizada já foram aplicadas no ambiente local desta
entrega. As nove reservas anteriores foram preservadas; cada uma recebeu um
evento **Registro inicial**, observado no momento da migração.

## Roteiro de aceite

1. Atualize o navegador e abra **Reservas**. Pesquise `CP1-01`, cuja entrada
   está em 14/11/2026. Antes dessa data, a linha deve oferecer **Cancelar**.
2. Abra a confirmação e confira hóspede, quarto e datas. **Voltar** deve fechar
   sem alterar a reserva; confirmar deve mostrar sucesso e o status **Cancelada**.
   O cancelamento é definitivo: não há reativação nesta entrega.
3. Abra **Auditoria** pelo menu e busque o quarto ou hóspede. Confira **Cancelamento**, funcionário,
   horário de São Paulo e **Confirmada → Cancelada**. Expanda **Ver antes e depois**
   e confira a preservação de quarto, datas, ocupantes e valores.
4. Na busca pública, consulte o período da reserva cancelada: o quarto deve
   estar disponível, salvo se existir outra reserva ativa para ele.
5. Em **Hóspedes**, abra o histórico desse hóspede: a reserva continua presente.
6. Abra **Auditoria** pelo menu. Busque nome/quarto/referência e confira os
   registros globais. Reservas antigas têm **Registro inicial**, não uma criação
   ou um cancelamento retroativamente atribuídos a alguém.
7. Reservas iniciadas, encerradas ou já canceladas não oferecem **Cancelar**.
   Um estado desatualizado no diálogo deve bloquear a confirmação e oferecer
   **Atualizar consulta**. Ambos os perfis, administrador e recepção, têm acesso.

As datas acima pertencem à carga de 11/11/2026. Se já passaram, crie uma nova
reserva futura para testar o cancelamento. A API sempre aplica a data do hotel.

## O que a auditoria registra

Registra ações realizadas: criação, cancelamento e alterações de dados. A edição
pela tela será entregue no checkpoint 4. Reservas públicas identificam a origem
**Site público**; ações autenticadas mostram o funcionário, preservado mesmo
se o nome dele mudar depois. Mudanças automáticas pelas datas são calculadas e
não geram eventos. Não existem ações de editar/apagar a auditoria na aplicação.

## Validação técnica realizada

- PostgreSQL 17 isolado, testes Go sem SKIP, go vet, lint e build.
- Concorrência: dois cancelamentos simultâneos resultam em um sucesso e um
  conflito, com apenas um evento de cancelamento.
- Falha induzida ao gravar o evento: reserva permanece ativa e com a mesma versão.
- Versão desatualizada, repetição, entrada hoje, estadia em andamento e encerrada
  são bloqueadas; cancelamento libera o quarto para nova reserva.
- API: funcionário obtido pelo token, rejeição de autor no JSON, acesso protegido,
  respostas 400/401/404/409/422 e consulta de auditoria com ambos os perfis.
- Navegador em desktop/celular: confirmação, desistência, cancelamento, atualização
  da linha, antes/depois, funcionário e recuperação após conflito de versão.

Próximo checkpoint: edição de reservas, com autoria também registrada na auditoria.
