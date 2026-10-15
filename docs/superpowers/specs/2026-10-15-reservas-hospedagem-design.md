# Reservas de hospedagem — especificação de projeto

## Intenção

Permitir que um hóspede reserve um quarto pelo site público, sem criar uma conta, e que a recepção registre reservas pelo painel interno. Os dois canais devem consultar a mesma disponibilidade e usar as mesmas regras de preço e conflito. O hóspede informa nome, tipo e número de documento (RG ou CPF) e e-mail antes da confirmação. Na chegada, a recepção poderá usar esses dados para identificar o hóspede; registrar o check-in e o check-out realizados não faz parte desta entrega.

## Contexto atual

O projeto usa React, TypeScript, Go e PostgreSQL. A aplicação atual é voltada à equipe e protege as páginas internas por autenticação. Já existem cadastros de hóspedes, categorias e quartos. O hóspede tem nome, documento, e-mail, telefone e endereço; somente o nome é obrigatório hoje. O quarto possui categoria, capacidade, diária (`daily_rate`) e disponibilidade manual (`is_available`).

Ainda não existem reservas, calendário de ocupação, consulta de disponibilidade por período, cálculo de diárias, tela de reserva ou ações de check-in/check-out. As rotas atuais de hóspedes e quartos exigem autenticação, e `/cadastro` redireciona para o login. A coluna atual de documento não identifica se o valor é RG ou CPF.

## Objetivos

- Permitir consulta pública de quartos disponíveis para um período e quantidade de pessoas.
- Exibir diária, quantidade de diárias e valor previsto antes de confirmar.
- Permitir reserva pública sem conta, coletando nome, tipo/número de documento e e-mail.
- Permitir que a recepção selecione um hóspede cadastrado e crie a mesma reserva pelo painel autenticado.
- Vincular cada reserva a um hóspede responsável, a um quarto, ao período previsto e à quantidade total de ocupantes.
- Impedir reservas sobrepostas, inclusive quando duas confirmações ocorrerem ao mesmo tempo.
- Reutilizar os mesmos cálculos e regras de disponibilidade nos dois canais.

## Fora do escopo

- Criar contas ou autenticar hóspedes.
- Registrar o check-in ou check-out realmente realizado.
- Pagamentos, cobrança de sinal, cancelamentos, reembolsos ou tarifas sazonais.
- Notificações por e-mail ou integração com sistemas externos.
- Reservar vários quartos em uma única operação.

## Fluxos

### Hóspede pelo site público

1. Informa check-in e check-out previstos e a quantidade total de pessoas.
2. Consulta os quartos livres que comportam essa quantidade.
3. Vê a diária, o número de noites e o valor previsto de cada opção.
4. Escolhe um quarto e informa nome, tipo de documento (RG ou CPF), número do documento e e-mail.
5. Revisa os dados e confirma. O servidor verifica a disponibilidade novamente antes de gravar a reserva.
6. Recebe uma confirmação da reserva na tela.

### Recepção pelo painel

1. Seleciona um hóspede cadastrado. Se faltarem os dados exigidos para uma reserva, completa o cadastro antes de continuar.
2. Informa o período previsto e a quantidade total de ocupantes.
3. Consulta quartos disponíveis com capacidade suficiente, diária, noites e total previsto.
4. Confirma a reserva. O servidor aplica a mesma verificação de conflito usada pelo site.

## Regras de negócio

- Check-in e check-out são datas previstas, sem horário. Check-out deve ser posterior ao check-in; cada noite é uma diária.
- A quantidade total de ocupantes inclui o hóspede responsável e não pode ultrapassar a capacidade do quarto.
- A diária atual do quarto é a base do cálculo. O servidor calcula o total usando decimal/moeda, nunca ponto flutuante binário.
- Ao confirmar, a reserva guarda um retrato da diária e do valor total para preservar o preço apresentado, mesmo que a diária do quarto seja alterada depois.
- O intervalo de ocupação é `[check-in, check-out)`: o quarto fica ocupado na noite anterior ao check-out e pode receber outra chegada na própria data do check-out.
- Um quarto só pode ser reservado se estiver habilitado (`is_available = true`) e não tiver outra reserva ocupando qualquer noite do período solicitado.
- O servidor repete a consulta de conflito dentro da operação de gravação, com proteção transacional no PostgreSQL, para impedir dupla reserva sob concorrência.
- Documento deve registrar explicitamente o tipo RG ou CPF. Números devem ser normalizados para comparação e unicidade; o projeto ainda não valida formato de RG/CPF.
- A confirmação da reserva é imediata; não há etapa de pagamento nem aprovação manual nesta entrega.

## Dados e API

Criar uma nova migração PostgreSQL para reservas e para identificar o tipo de documento do hóspede sem apagar dados existentes. Hóspedes novos devem ter tipo do documento; cadastros existentes recebem tipo desconhecido (`NULL`) e só são classificados pela recepção. A reserva precisa armazenar, no mínimo: hóspede responsável, quarto, check-in previsto, check-out previsto, quantidade de ocupantes, diária aplicada, total previsto, nome/documento/e-mail informados para aquela reserva e data de criação. Os dados de contato da reserva preservam o que foi informado na confirmação, sem substituir automaticamente os dados principais do hóspede. Definir índices e restrições para consultas por período e prevenção de sobreposição.

Adicionar operações para consultar disponibilidade por período e ocupação, criar reserva pelo canal público e criar reserva pela recepção autenticada. A operação pública deve aceitar somente os dados necessários para buscar quartos e confirmar uma reserva; ela não deve liberar os endpoints administrativos existentes de hóspedes e quartos. Manter a arquitetura atual: handlers HTTP, regras no serviço `hotel` e persistência em `PostgresRepository`.

O fluxo público deve apresentar apenas informações necessárias para escolher e reservar (categoria, capacidade e preço), sem expor dados cadastrais de hóspedes. A criação pública deve localizar ou criar o cadastro do hóspede e associá-lo à reserva, respeitando as restrições de unicidade existentes. Quando o documento informado corresponder a um cadastro legado cujo tipo é desconhecido, a reserva pode referenciar esse hóspede sem preencher ou alterar o tipo legado; o tipo e número informados ficam preservados na reserva. Se o documento corresponder a um hóspede tipado e o e-mail informado for diferente, a reserva guarda o e-mail informado como contato daquela estadia e não altera o e-mail principal do hóspede.

## Interface

- Criar páginas públicas de busca, resultados, identificação, revisão e confirmação.
- Criar uma área autenticada de reservas para a recepção, incluindo seleção/localização do hóspede.
- Exibir estados de carregamento, nenhum quarto disponível, dados inválidos e conflito detectado ao confirmar.
- Atualizar a navegação do painel para habilitar Reservas.
- Não incluir controles de check-in ou check-out realizado.

## Critérios de aceite

- O hóspede consulta quartos sem conta e só precisa informar seus dados para concluir a reserva.
- A recepção consegue criar reserva para um hóspede existente, completando dados obrigatórios se necessário.
- O sistema rejeita datas inválidas e quantidade de pessoas acima da capacidade do quarto.
- A quantidade de diárias e o total exibidos correspondem às datas e à diária do quarto.
- Quartos indisponíveis manualmente ou já ocupados em qualquer noite não aparecem como disponíveis.
- Uma segunda reserva conflitante é recusada mesmo quando a disponibilidade mudou após a consulta inicial.
- Duas confirmações simultâneas não conseguem reservar o mesmo quarto para noites sobrepostas.
- Reservas que se sucedem na data de check-out anterior são permitidas.
- A reserva mantém o valor confirmado se a diária do cadastro do quarto mudar depois.
- Não existe fluxo para conta do hóspede, pagamento ou check-in/check-out realizado.

## Decisões do responsável pelo produto

- Os dois canais fazem parte da entrega: hóspede no site e recepção no painel.
- O hóspede não terá conta nesta entrega.
- Nome, tipo/número de documento (RG ou CPF) e e-mail são obrigatórios para reservar.
- A busca considera o total de ocupantes para respeitar a capacidade do quarto.
- Check-in e check-out realizados ficam fora do escopo; somente as datas previstas são registradas.

## Decisões de implementação propostas

- Uma reserva corresponde a um quarto; pedidos com vários quartos são registrados como reservas separadas.
- O responsável pela reserva conta no total de ocupantes.
- O período segue a convenção `[check-in, check-out)`, liberando o quarto no dia do check-out.
- O preço é fixo por quarto conforme o `daily_rate` atual, sem sazonalidade; diária e total são guardados na reserva.
- Cadastros existentes mantêm tipo de documento desconhecido até a recepção classificá-los; não inferir RG/CPF pelos números já armazenados.
- A reserva pública reutiliza um cadastro quando encontra o mesmo número de documento normalizado. Para cadastros legados sem tipo, preserva o tipo desconhecido no registro principal e guarda o tipo informado na reserva.
- Se o e-mail informado na reserva divergir do e-mail principal do hóspede, salvar o informado como contato da reserva sem alterar o registro principal.
- RG e CPF serão armazenados com seu tipo. Validação formal do número (incluindo dígitos verificadores) fica para uma decisão posterior, pois RG tem formatos variáveis e o escopo não definiu validação documental.

## Pontos para o plano de implementação

- Verificar se a migração consegue evoluir a coluna de documento e o índice únicos existentes sem ambiguidade para cadastros atuais, nos quais o tipo ainda não está registrado.
- Escolher a proteção PostgreSQL contra sobreposição e demonstrar o comportamento sob confirmações concorrentes.
- Evitar divulgar dados cadastrais no retorno público ao reutilizar hóspede por documento.
- Atualizar `README.md`, `backend/README.md` e `frontend/README.md`, que hoje descrevem partes já implementadas como pendentes e não documentam reservas.
