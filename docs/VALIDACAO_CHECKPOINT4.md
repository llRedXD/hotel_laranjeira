# Checkpoint 4 — alteração de reservas

Não há nova migração. O checkpoint utiliza a auditoria da migração 007 e as
regras de status já implantadas. Para carregar a edição em outros ambientes,
recompile/reinicie a API e atualize o frontend.

## Roteiro de aceite

1. Em **Reservas**, abra **Alterar** em uma reserva com entrada futura. Na base
   local, `Teste CP1 - Confirmada` tem entrada em 14/11/2026. Se essa data já
   passou ou a reserva foi cancelada, crie outra reserva futura.
2. Na etapa **Escolher**, selecione **Datas da estadia**, **Quantidade de pessoas**
   ou **Trocar de quarto**. O hóspede responsável permanece o mesmo.
3. Na etapa **Ajustar**, apenas os campos escolhidos devem aparecer. Em datas ou
   pessoas, clique em **Revisar alteração**: o quarto atual será mantido se atender
   à alteração. Se não atender, a tela oferece incluir uma troca de quarto por
   escolha explícita. Em **Trocar de quarto**, as datas e pessoas são mantidas;
   consulte as alternativas e escolha uma acomodação.
4. Na etapa **Revisar**, confira a comparação **Antes / Depois** e o novo total.
   No mesmo quarto, a diária contratada é preservada mesmo
   que o cadastro do quarto tenha outro valor. Para validar isso manualmente,
   altere a diária em **Quartos** antes de consultar a edição.
5. Clique em **Confirmar alteração** e confira **Alteração concluída**. Apenas
   essa confirmação grava os dados. **Fazer outra alteração** reinicia o fluxo
   usando a reserva atualizada. No menu **Auditoria**,
   busque o hóspede e expanda **Ver antes e depois** no evento **Alteração**.
   Confira datas, ocupantes, valores e funcionário.
6. Repita escolhendo outro quarto: a revisão deve usar a diária atual da nova
   acomodação. Após salvar, o quarto anterior fica livre para o período, e o
   destino fica ocupado, respeitando outras reservas que existam.
7. **Voltar ao ajuste** preserva o rascunho e descarta a cotação, exigindo nova
   verificação. **Escolher outra alteração** reinicia os dados ao selecionar outro
   tipo; uma mudança de datas abandonada não deve vazar para a alteração de pessoas.
   Avançar e voltar não podem gerar eventos de auditoria. Sem mudança efetiva,
   **Revisar alteração** fica desabilitado.
8. Para testar um preço desatualizado, consulte uma troca de quarto, altere a
   diária do destino em outra aba e tente salvar. O servidor deve bloquear e a
   tela deve oferecer **Recarregar reserva**, reiniciando a escolha, consulta e revisão.
9. Reservas iniciadas, encerradas ou canceladas não oferecem **Alterar**. Abrir
   diretamente sua URL de edição exibe os dados e o bloqueio, sem formulário.
10. **Voltar às reservas** deve preservar os filtros da listagem de origem.

## Validação técnica realizada

- Testes Go com PostgreSQL 17, sem SKIP, `go vet`, lint e build.
- Diária contratada preservada mesmo com diária atual que excederia o limite
  do total; troca usa a diária atual e mantém cálculo exato em centavos.
- Preservação de hóspede e snapshots, versão e auditoria com autoria;
  requisição sem mudanças não incrementa versão nem gera evento.
- Validação de nova entrada futura, capacidade, quarto habilitado, preço,
  disponibilidade, cancelamento e versão consultada.
- Concorrência: edição contra criação, duas edições da mesma versão, edição
  contra cancelamento e movimentos em sentidos opostos, sem dupla reserva.
- Falha induzida na auditoria desfaz a alteração inteira.
- API: acesso protegido, campo de responsável recusado, disponibilidade pública
  não permite excluir reserva, respostas 400/401/404/409/422.
- Navegador: administrador/recepção, desktop/celular, revisão, preço alterado
  entre consulta e gravação, retorno aos filtros e auditoria do funcionário.

Refinamento por etapas validado no navegador: campos focados por tipo,
preservação/descarte do rascunho ao voltar/trocar de tipo, gravação somente na
confirmação final, seleção explícita de quarto quando a capacidade é insuficiente,
reinício após conflito de preço e revisão móvel sem rolagem horizontal. Os testes
do refinamento usaram banco temporário, preservando os dados locais.

Próximo checkpoint: validação integrada e aceite final das sete funcionalidades.
