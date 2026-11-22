# Ciclo operacional da reserva — especificação

## Intenção

Permitir que a recepção conclua operacionalmente uma reserva: registrar a chegada, ocupar o quarto, registrar a saída, calcular o valor final e liberar o quarto. Isso estende a gestão existente de reservas, que hoje trabalha com períodos e status previstos.

## Estado atual

- Reservas guardam check-in/check-out previstos, diária e total previstos, além de versão e auditoria.
- O status atual (`confirmed`, `in_progress`, `departure_today`, `ended`, `cancelled`) é derivado das datas previstas; não comprova ações realizadas.
- Quartos têm `is_available`, um controle administrativo de habilitação, sem estado operacional ocupado/vago.
- A API é Go/PostgreSQL e o painel é React/TypeScript. Rotas operacionais da recepção exigem autenticação.
- Auditoria de reservas já registra mudanças transacionalmente, com responsável e estados antes/depois.

## Objetivos

- Registrar check-in realizado em reserva válida e ainda não iniciada.
- Atualizar o estado operacional da reserva para hospedagem e marcar seu quarto como ocupado na mesma transação.
- Registrar check-out realizado, guardar o valor final calculado, finalizar a reserva e liberar o quarto na mesma transação.
- Mostrar na tela as ações disponíveis conforme o estado real da reserva e do quarto.
- Preservar histórico de datas realizadas, valor final, responsável e horário de cada operação.
- Impedir repetições, transições fora de ordem, conflito de ocupação e gravações parciais.

## Regras propostas

- Check-in/check-out realizados ficam separados das datas previstas já existentes.
- As operações são restritas a usuários autenticados da equipe (administrador e recepção), conforme o padrão atual.
- Check-in só pode ocorrer uma vez, em reserva não cancelada e ainda sem check-in, antes do check-out previsto; isso permite chegada antecipada e chegada após o dia previsto enquanto a estadia prevista não terminou. Marca o quarto como ocupado.
- Check-in antes da data prevista é permitido se o quarto estiver livre e não houver outra reserva ativa ocupando a data efetiva de chegada; as noites antecipadas entram no cálculo final.
- Check-out só pode ocorrer após check-in e uma única vez. Calcula valor final pela diária guardada na reserva multiplicada pelo número de noites entre as datas efetivas; não substitui a diária contratada pela diária atual do cadastro do quarto.
- Check-out anterior ao check-in realizado é rejeitado.
- Datas operacionais e número de noites seguem a data civil do hotel (`America/Sao_Paulo`) e a convenção existente de noites: checkout menos check-in. Uma estadia iniciada e encerrada na mesma data resulta em zero noites/valor zero, salvo ajuste aprovado pelo responsável.
- Check-in e check-out, atualização de reserva, estado do quarto e evento de auditoria são atômicos. Falha em qualquer gravação reverte a operação inteira.
- Disponibilidade, criação e indicadores tratam o intervalo efetivo (check-in realizado até checkout realizado ou previsto) como ocupado; reservas futuras não sobrepostas continuam permitidas.
- Quarto ocupado não pode ser marcado como livre por ação administrativa enquanto a reserva estiver em hospedagem. A flag atual `is_available` segue representando habilitação para venda; estado operacional será modelado separadamente.
- Um quarto pode receber reserva futura enquanto está ocupado se o período futuro não se sobrepuser ao período previsto da estadia atual; o check-in futuro valida novamente que o quarto foi liberado.
- Reserva cancelada não pode iniciar hospedagem. Uma reserva já iniciada não pode ser cancelada pelo fluxo de cancelamento existente.
- O estado real da reserva passa a ter precedência na apresentação sobre status previsto por datas; reservas legadas sem operações realizadas continuam com o cálculo atual.

## Fluxo da recepção

1. Na listagem/detalhe da reserva, a equipe confere hóspede, quarto, período previsto e status.
2. Para uma reserva confirmada, escolhe “Realizar check-in”; o servidor valida estado e versão, registra horário/responsável e ocupa o quarto.
3. Durante a hospedagem, a reserva aparece “Em hospedagem” e o quarto “Ocupado”.
4. A equipe escolhe “Realizar check-out”; a interface apresenta datas efetivas, noites e valor final para confirmação.
5. O servidor calcula novamente o valor, registra checkout/responsável/horário, finaliza a reserva e libera o quarto.
6. A reserva concluída mantém os valores previstos e o valor final, e seu histórico exibe os eventos operacionais.

## Fora do escopo

- Pagamentos, descontos, consumo, multas, reembolsos ou divisão de pagamento.
- Alteração de diárias/tarifas durante a hospedagem.
- Check-in/check-out parcial por quarto, reservas com múltiplos quartos ou transferência de quarto durante estadia.
- Integração com fechaduras, sistemas externos ou notificações.
- Cobrança por hora ou regra de tolerância para checkout tardio.

## Persistência e API

- Criar migração aditiva para armazenar timestamps realizados e valor final; adicionar estado operacional do quarto ou mecanismo equivalente que sobreviva a reinício e seja consistente com reservas ativas.
- Criar operações autenticadas e versionadas para check-in e checkout, com autorização alinhada às rotas atuais.
- Expor uma prévia do checkout calculada pelo servidor para a confirmação da recepção; a operação de checkout recalcula o valor dentro da transação e retorna o valor efetivamente gravado.
- Aplicar bloqueio/ordem consistente de linhas para serializar operações concorrentes envolvendo a mesma reserva ou quarto.
- Usar transação PostgreSQL para estado da reserva, quarto e auditoria; nenhuma ação deve produzir estado parcial.
- Retornar conflitos de estado/versão como HTTP 409 e dados inválidos como HTTP 422, seguindo os contratos atuais.
- Expor ao painel os campos realizados, valor final e permissões de transição. Listagens/status previstos devem manter compatibilidade para reservas sem operações realizadas.

## Interface

- Disponibilizar ações de check-in e checkout na área autenticada de reservas.
- Pedir confirmação antes de cada transição e, no checkout, exibir a memória do cálculo final.
- Após sucesso, atualizar a reserva e o estado do quarto; tratar conflito concorrente e permitir recarregar dados.
- Exibir estado ocupado no cadastro/lista de quartos, distinguindo-o de quarto desabilitado para venda.
- Tornar datas, horários, valor calculado e responsável consultáveis no detalhe/histórico.

## Critérios de aceite

- Check-in válido registra data/hora e funcionário, muda a reserva para “Em hospedagem” e o quarto para “Ocupado”.
- Não é possível fazer check-in de reserva cancelada, já iniciada, concluída ou em conflito com quarto ocupado.
- Checkout válido registra data/hora e funcionário, calcula diária contratada × noites efetivas, finaliza a reserva e libera o quarto.
- Repetição de check-in/out, versão obsoleta e checkout sem check-in não alteram dados e retornam conflito compreensível.
- Operações simultâneas não produzem dupla ocupação nem estado parcial; erro de auditoria reverte também quarto e reserva.
- Checkout no mesmo dia segue a regra aprovada de zero noites e valor zero, salvo decisão diferente antes do plano.
- Reservas legadas continuam mostrando status previsto; suas datas e totais originais não são reescritos.
- A equipe consegue realizar e consultar o ciclo pela interface autenticada em desktop e celular.

## Decisões confirmadas

- O escopo inclui check-in, estado “Em hospedagem”, quarto “Ocupado”, checkout, cálculo final, reserva finalizada e quarto liberado.
- O cálculo final usa a diária contratada registrada na reserva multiplicada pelas noites entre as datas efetivas.
- A flag atual `is_available` continua separada do estado operacional de ocupação.
- Check-in antecipado é permitido quando não há conflito de ocupação; a cobrança considera as noites adicionais.
- Checkout anterior à entrada realizada é inválido; reservas futuras não sobrepostas continuam permitidas para o quarto ocupado.

## Decisões de implementação

- Reutilizar versionamento, auditoria e transações existentes, acrescentando ações `checked_in` e `checked_out`.
- Registrar timestamps em instantes UTC e apresentar no fuso `America/Sao_Paulo`; derivar noites das datas civis desse fuso.
- Criar o plano detalhado após validação desta especificação.
