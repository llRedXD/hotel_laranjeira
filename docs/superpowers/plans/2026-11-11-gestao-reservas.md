# Gestão de reservas — plano aprovado e checkpoints

## Regras aprovadas

- Equipe autenticada: administrador e recepção.
- Alterar/cancelar somente com entrada posterior à data atual em America/Sao_Paulo.
- Editar quarto, entrada, saída e ocupantes; preservar responsável e snapshots.
- Mesmo quarto mantém diária contratada; novo quarto usa diária atual.
- Status previstos por datas, com prioridade para cancelamento. Sem check-in/checkout realizados, pagamentos, reembolsos ou reativação.
- Filtro por período encontra noites no intervalo inclusivo: entrada <= to e saída > from; aceitar limites isolados e rejeitar intervalo invertido.
- Histórico por guestId inclui reservas antigas e canceladas, sem misturar homônimos.
- Pausar após cada checkpoint para validação do usuário antes de avançar.

## 1. Banco e status

- [x] Migração 006 aditiva: cancelled_at, cancelled_by com referência a users e version positiva, inicialmente 1.
- [x] Exigir data e responsável juntos no cancelamento; preservar dados existentes.
- [x] Calcular status no serviço usando data civil do hotel; disponibilizar status na listagem HTTP.
- [x] Ignorar canceladas na disponibilidade, criação e indicadores; preservar canceladas na listagem.
- [x] Ordenar a listagem usando a mesma data utilizada no cálculo dos status.
- [x] Testar banco novo, migração com reserva existente, status, fuso, liberação/reutilização de quarto e indicadores.
- [x] Aceite do usuário: checkpoint 1 validado na conversa.

Validação realizada: go test ./... -count=1 -v com PostgreSQL 17 isolado, sem testes pulados; go vet ./...; executor migrate.sql em banco novo e reexecução sem reaplicar versões. O banco do projeto não foi migrado nesta validação.

O cancelamento foi exercitado por SQL no teste de integração. A ação HTTP/tela pertence ao checkpoint 3. O frontend consumirá os status do backend no checkpoint 2.

## 2. Consulta, período e histórico

- [x] Adicionar guestId ao GET /api/v1/reservations, preservando busca, paginação e total.
- [x] Retornar status, versão e permissões de alteração/cancelamento; frontend usa status do backend.
- [x] Implementar semântica de ocupação para from/to, inclusive nas canceladas.
- [x] Abrir histórico a partir do cadastro de hóspedes, com identificação do hóspede.
- [x] Usar a data do hotel no painel, inclusive no resumo padrão sem parâmetro date.
- [x] Testar homônimos, limites, filtros combinados, paginação, vazio e autenticação.
- [x] Aceite do usuário: checkpoint 2 validado na conversa.

Validação do checkpoint 2: testes Go com PostgreSQL 17 isolado sem SKIP,
go vet, lint e build; API com administrador/recepção; navegador em desktop e
390px, histórico pelo cadastro, filtros e limpeza preservando hóspede, erro de
intervalo e resultado vazio, painel usando a data do hotel e excluindo canceladas.
Roteiro de aceite: `docs/VALIDACAO_CHECKPOINT2.md`.

## 3. Cancelamento completo

- [x] POST /api/v1/reservations/{id}/cancel autenticado com versão consultada.
- [x] Validar entrada futura, estado e versão em transação; registrar funcionário/horário e incrementar versão.
- [x] Retornar 404 para inexistente e 409 para estado/versão incompatíveis.
- [x] Confirmação na tela com hóspede/quarto/datas e atualização após sucesso.
- [x] Testar concorrência, bloqueio a partir da entrada, histórico e reutilização de quarto.
- [x] Aceite do usuário: checkpoint 3 validado; acesso à auditoria mantido pelo menu, conforme ajuste solicitado.

Extensão solicitada: tela de auditoria global e por reserva, com funcionário, horário e antes/depois. Padrão adotado: ações realizadas; mudanças automáticas pelas datas não geram eventos. Migração 007 cria registros iniciais para dados anteriores sem reconstruir histórico. Criação, cancelamento e futuras alterações são auditados transacionalmente.

Validação: integração PostgreSQL, concorrência, rollback quando auditoria falha, API e navegador com ambos os perfis e desktop/celular. Migração aplicada localmente, preservando nove reservas; API reiniciada com a configuração anterior. Roteiro: `docs/VALIDACAO_CHECKPOINT3.md`.

## 4. Alteração completa

- [x] GET/PATCH /api/v1/reservations/{id} autenticados; PATCH recebe quarto, datas, ocupantes, diária revisada e versão.
- [x] Disponibilidade interna desconsidera a própria reserva, exclusivamente nas rotas autenticadas.
- [x] Registrar a autoria de cada edição no contexto transacional da auditoria.
- [x] Formulário preenchido e revisão de valor; validar nova entrada futura, capacidade, quarto habilitado, preço e versão.
- [x] Recalcular total sem ponto flutuante e aplicar regra de preservação da diária.
- [x] Proteger contra sobreposição e bloquear quartos em ordem consistente ao trocar de quarto.
- [x] Testar alteração concorrente com criação, alteração e cancelamento, além dos erros de validação.
- [x] Aceite do usuário: implementação e refinamento por etapas aceitos ao solicitar o checkpoint 5.


Validação: testes Go com PostgreSQL 17 sem SKIP, go vet, lint/build; concorrência entre criação/edição/cancelamento, movimentos opostos e rollback da auditoria. API e navegador com ambos os perfis, revisão desktop/celular, preço desatualizado, diária preservada e auditoria. Sem nova migração ou alteração dos dados locais de teste; API reiniciada. Roteiro: `docs/VALIDACAO_CHECKPOINT4.md`.

## 5. Validação final e documentação

Refinamento solicitado no checkpoint 4: edição em três etapas (escolher,
ajustar, revisar). Padrão adotado: uma alteração por vez, de datas, pessoas ou
quarto. Caso datas/pessoas exijam outra acomodação, a troca depende de escolha
explícita e aparece na revisão. O fluxo só grava na confirmação final. Validado
no navegador, inclusive voltar, trocar de tipo, conflito de preço e celular;
lint e build aprovados. Auditoria permanece acessível pelo menu.

- [x] Atualizar contratos HTTP e documentação.
- [x] Executar testes Go, go vet, lint e build; integração PostgreSQL não pode ser pulada.
- [x] Validar no navegador: administrador/recepção, desktop/celular, estados de erro e carregamento.
- [x] Demonstrar consulta, filtros, histórico, alteração, cálculo e cancelamento; conferir bloqueios e criação pública/interna.
- [ ] Aceite final das sete funcionalidades.

Validação final aprovada tecnicamente: integração completa pública → gestão → edição → cancelamento → nova reserva interna, seis eventos auditados, migração nova/upgrade/reexecução e testes sem SKIP. Corrigida cotação desatualizada na criação pela recepção e bloqueada a edição de período/pessoas durante consultas. Relatório e roteiro de aceite: `docs/VALIDACAO_FINAL_RESERVAS.md`. Aceite final do responsável permanece pendente.
