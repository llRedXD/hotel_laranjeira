# Roteiro de apresentação — ciclo da reserva

## Preparação

1. Inicie o ambiente local com `docker compose up --build`.
2. Carregue os dados preparados, na raiz do projeto:

   ```bash
   docker compose exec -T database psql -U hotel_laranjeira -d hotel_laranjeira -v ON_ERROR_STOP=1 < database/seed_presentation.sql
   ```

3. Acesse <http://localhost:5173/login> e entre como recepção:
   - E-mail: `recepcao@hotel.com`
   - Senha: `recepcao123456`

O script usa uma transação e pode ser executado novamente para restaurar o
cenário. Ele não apaga reservas, quartos ou hóspedes fora dos registros
exclusivos da apresentação. Use apenas em ambiente local de desenvolvimento.

## Cenários preparados

| Quarto | Hóspede | Situação inicial | O que demonstrar |
| --- | --- | --- | --- |
| DEMO-901 | Marina Costa (Apresentação) | Confirmada, período de amanhã até quatro dias a partir de hoje, diária R$ 350 | Check-in antecipado, reserva em hospedagem e quarto ocupado |
| DEMO-902 | Rafael Lima (Apresentação) | Em hospedagem desde dois dias atrás, diária contratada R$ 350 | Cotação e checkout imediato, reserva finalizada e quarto liberado |

O checkout da reserva DEMO-902 calcula duas noites × R$ 350, totalizando
**R$ 700,00**. A cotação usa o relógio do servidor; carregue o script novamente
se a data ou o estado da apresentação tiver mudado.

## Apresentação em sequência

### 1. Check-in

1. Abra **Reservas** e localize Marina Costa ou o quarto DEMO-901.
2. Abra os detalhes e mostre a reserva confirmada, a diária de R$ 350 e o
   período planejado.
3. Selecione **Check-in** e confirme.
4. Mostre o status **Em hospedagem**, o horário e funcionário registrados e o
   evento no histórico.
5. Abra **Quartos** e mostre o DEMO-901 como **Ocupado**.

Explique que o check-in pode ser feito antes da data prevista quando o quarto
não está ocupado por outra hospedagem.

### 2. Checkout e fechamento

1. Em **Reservas**, abra Rafael Lima no quarto DEMO-902.
2. Mostre que o cenário já está em hospedagem e selecione **Check-out**.
3. Apresente a cotação do servidor: duas noites, diária contratada de R$ 350 e
   total final de **R$ 700,00**.
4. Confirme o checkout.
5. Mostre o status **Finalizada**, valor total e horário/responsável pelo
   checkout no detalhe e no histórico.
6. Abra **Quartos** e confirme que DEMO-902 não está mais ocupado.

### 3. Repetir a demonstração

Depois de percorrer os dois cenários, execute novamente
`database/seed_presentation.sql`. Isso restaura a reserva DEMO-901 para check-in
e a DEMO-902 para checkout, apagando somente o histórico associado a essas duas
reservas de apresentação.
