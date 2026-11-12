# Apresentação — gestão de reservas do Hotel Laranjeira

**Duração sugerida:** 8 a 10 minutos.

**História da demonstração:** a recepção precisa localizar a reserva de uma
hóspede, consultar seu histórico, estender a estadia e, depois, registrar um
cancelamento. A equipe acompanha os valores e as ações pela auditoria.

## Preparar antes da apresentação

- Iniciar o sistema e deixar o painel aberto em `http://localhost:5173` com uma
  conta de administrador ou recepção que já funcione no ambiente.
- Usar uma hóspede fictícia: **Ana Apresentação**, tipo de documento **RG**,
  documento **APRES-001**, e-mail **ana.apresentacao@demo.hotel.test**.
- Completar esses dados no cadastro e criar uma reserva para essa hóspede pela
  tela **Nova reserva**. Escolher um quarto disponível para duas pessoas.
- Usar entrada **daqui a 7 dias** e saída **daqui a 9 dias**: duas diárias.
  Anotar as datas concretas, o quarto, a diária e o total exibidos.
- Se quiser um histórico com mais de um registro, preparar outra reserva da
  mesma hóspede para um período diferente, sem conflito.
- Conferir que as reservas usadas para alteração/cancelamento têm entrada
  futura. Reservas iniciadas ou canceladas não permitem essas operações.
- Deixar uma reserva encerrada e uma cancelada localizadas na listagem, se já
  houver exemplos na base, para mostrar a variedade dos status.

Não é necessário recarregar os seeds ou apagar dados para apresentar. O fluxo
abaixo altera e cancela a reserva fictícia preparada; faça uma nova reserva
futura para repetir a apresentação.

## 1. Abertura — 30 segundos

**Mostrar:** painel da equipe.

**Fala sugerida:**

> “Nesta entrega, ampliamos a gestão das reservas existentes. A equipe consegue
> consultar reservas, localizar hóspedes, filtrar estadias, acompanhar status,
> alterar ou cancelar reservas futuras e consultar o histórico. Também incluímos
> uma auditoria para identificar quem realizou cada ação.”

## 2. Consulta e status — 1 minuto

**Fazer:** abrir **Reservas** e mostrar as colunas de hóspede, quarto, datas,
ocupantes, status e total. Mostrar a contagem e a paginação, se houver mais páginas.

**Fala sugerida:**

> “Esta é a visão geral das reservas. A recepção acompanha o período, a acomodação,
> a quantidade de pessoas e o valor previsto sem precisar abrir cada cadastro.”

Apontar os status disponíveis: **Confirmada**, **Estadia prevista em andamento**,
**Saída prevista hoje**, **Encerrada pelo período** e **Cancelada**.

> “Os status de estadia são previstos pelas datas. O registro de chegada e saída
> efetivamente realizadas fica para outra entrega.”

## 3. Filtros por hóspede e período — 1 minuto

**Fazer:** pesquisar `Ana Apresentação`. Mostrar que também é possível buscar
pelo documento. Combinar a busca com um intervalo que comece **daqui a 8 dias**
e termine **daqui a 8 dias**.

**Resultado esperado:** a reserva preparada aparece, apesar de começar um dia
antes do intervalo pesquisado.

**Fala sugerida:**

> “A busca localiza reservas pelo nome ou documento do hóspede. O filtro de
> período considera as noites ocupadas: esta reserva começou antes da data
> pesquisada, mas aparece porque ainda ocupa essa noite.”

Limpar os filtros antes de continuar.

## 4. Histórico do hóspede — 1 minuto

**Fazer:** abrir **Hóspedes**, localizar Ana e clicar em **Histórico de reservas**.
Mostrar a identificação da hóspede e suas reservas.

**Fala sugerida:**

> “Pelo cadastro, acessamos as reservas vinculadas especificamente a esta hóspede.
> O histórico reúne os registros sem misturar pessoas que tenham o mesmo nome.
> Reservas antigas e canceladas também permanecem disponíveis para consulta.”

O nome da hóspede na listagem de reservas também abre esse histórico.

## 5. Alteração por etapas — 2 minutos

**Fazer:** na reserva preparada, clicar em **Alterar**.

**Etapa 1 — Escolher:** mostrar **Datas da estadia**, **Quantidade de pessoas**
e **Trocar de quarto**. Selecionar **Datas da estadia**.

> “A pessoa escolhe o que deseja modificar. Assim, cada etapa mostra apenas os
> campos necessários àquela alteração.”

**Etapa 2 — Ajustar:** manter a entrada e mudar a saída para **daqui a 10 dias**.
Clicar em **Revisar alteração**.

> “Vamos estender a estadia de duas para três diárias. O sistema verifica se o
> quarto continua disponível e calcula o novo total.”

**Etapa 3 — Revisar:** mostrar a comparação **Antes / Depois**, destacando a nova
saída e o valor. Confirmar somente depois de apresentar essa comparação.

> “Como mantivemos o quarto, preservamos a diária contratada. A reserva só é
> gravada quando clicamos em Confirmar alteração.”

**Resultado esperado:** mensagem **Alteração concluída**, com o mesmo hóspede,
mesmo quarto e total de três diárias. Se a diária for R$ 280,00, por exemplo,
o total passa de R$ 560,00 para R$ 840,00; usar os valores reais da tela.

Mencionar as outras opções sem precisar executá-las:

> “Na troca de quarto, usamos a diária atual da nova acomodação. Ao aumentar a
> quantidade de pessoas, verificamos a capacidade; se for necessário outro quarto,
> essa troca precisa ser escolhida explicitamente e aparece na revisão.”

## 6. Auditoria — 1 minuto

**Fazer:** abrir **Auditoria** pelo menu e pesquisar `Ana Apresentação`.
Localizar o evento **Alteração** e expandir **Ver antes e depois**.

**Fala sugerida:**

> “A auditoria mostra a ação, o funcionário responsável e o horário. Aqui vemos
> os dados anteriores e posteriores, incluindo o período e os valores. Isso
> permite acompanhar o que foi feito com a reserva.”

Apontar a criação e a alteração registradas. Eventos anteriores à implantação
da auditoria não são reconstruídos: reservas antigas começam com **Registro inicial**.
Mudanças automáticas de status pela passagem das datas não geram eventos.

## 7. Cancelamento e preservação do histórico — 1 minuto e meio

**Fazer:** voltar a **Reservas**, localizar a mesma reserva e clicar em **Cancelar**.
Mostrar o diálogo com hóspede, quarto e datas antes de confirmar.

**Fala sugerida:**

> “O cancelamento exige confirmação para que a equipe confira a reserva correta.
> Após confirmar, o quarto é liberado para o período, e o registro permanece
> disponível no histórico.”

Confirmar e mostrar **Cancelada** na listagem. Apontar que as ações de alteração
e cancelamento deixam de aparecer nessa reserva.

Abrir o histórico de Ana e mostrar que a reserva continua presente. Pelo menu
**Auditoria**, pesquisar Ana e mostrar **Cancelamento**, o funcionário e a
transição **Confirmada → Cancelada**.

**Se houver tempo:** em **Nova reserva**, consultar o mesmo período e mostrar
o quarto liberado, desde que não exista outra reserva ativa para ele. Não é
necessário confirmar uma nova reserva para demonstrar a disponibilidade.

## 8. Encerramento — 30 segundos

**Fala sugerida:**

> “Demonstramos a consulta, os filtros por hóspede e período, os status, o histórico,
> a alteração por etapas e o cancelamento. A auditoria registra as ações da equipe,
> e as regras de disponibilidade, capacidade e preço são verificadas pelo servidor.”

> “A entrega cobre a administração das reservas. Pagamentos e o registro de
> check-in e checkout realizados ficam fora deste escopo.”

## Perguntas que podem surgir

| Pergunta | Resposta curta |
| --- | --- |
| Quem pode administrar as reservas? | Administrador e recepção autenticados. |
| Pode alterar ou cancelar uma reserva já iniciada? | Não. A entrada deve ser posterior à data atual do hotel, no fuso de São Paulo. |
| Alterar o cadastro do quarto muda reservas antigas? | A reserva guarda a diária contratada. Ao manter o quarto na edição, essa diária é preservada. |
| Como o sistema impede duas reservas no mesmo quarto? | Revalida a disponibilidade na gravação e usa bloqueios transacionais; a concorrência foi testada com PostgreSQL. |
| Cancelar apaga a reserva? | Não. Ela continua no histórico e na auditoria. Não há reativação nesta entrega. |
| O status confirma que o hóspede chegou? | Não. Os status de estadia são previstos pelas datas. |
| O hóspede precisa criar uma conta? | Não para reservar pelo site público; a administração exige acesso da equipe. |

## Versão curta — 5 minutos

Consulta/status e filtros (1 min) → histórico (30 s) → alteração de datas por
etapas (1 min 30 s) → cancelamento (1 min) → auditoria e encerramento (1 min).
