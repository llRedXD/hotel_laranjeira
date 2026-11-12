# Validação do checkpoint 2

Não há nova migração. Reinicie o processo da API para carregar o backend
atualizado e atualize o navegador. Se executa pela CLI Go, encerre o processo
anterior com Ctrl+C e repita o comando de inicialização no mesmo terminal.
O Vite recarrega as mudanças do frontend automaticamente.

## 1. Status na tela

Em `/reservas`, pesquise `Teste CP1`. Os cinco cenários existentes devem mostrar
o status enviado pela API, incluindo **Cancelada** em CP1-05. Os status seguem
a data atual do hotel; reservas podem mudar de situação à medida que passam os dias.

## 2. Filtro por ocupação

Com os dados carregados em 11/11/2026, pesquise `Teste CP1` e filtre estadias
de **11/11/2026 até 11/11/2026**. Apenas CP1-02 deve aparecer: começou em
10/11 e continua ocupando a noite do dia 11. CP1-03 sai no dia 11 e fica fora.

Filtre de **14/11/2026 até 16/11/2026**: CP1-01 e CP1-05 devem aparecer,
incluindo a cancelada pelo período reservado. Teste também somente o início,
somente o fim, intervalo invertido e **Limpar filtros**.

## 3. Histórico por hóspede

Em `/hospedes`, abra **Histórico de reservas** na linha de `Teste CP1 - Cancelada`.
Confirme nome, documento e somente a reserva CP1-05. Aplique e limpe filtros:
o hóspede deve continuar selecionado. Use **Ver todas as reservas** para sair
do histórico. Também é possível abrir o histórico pelo nome em uma reserva.

No cadastro de Gabriel Neira da base local, o histórico deve incluir as duas
reservas existentes, uma antiga e outra em andamento na data da carga.
Hóspedes com o mesmo nome são separados pelo identificador, não pela busca textual.

## 4. Painel e acesso

O painel usa a data do hotel em São Paulo e o resumo padrão do servidor.
A reserva cancelada não deve aparecer entre as próximas estadias nem afetar
ocupação/disponibilidade. Repita a consulta e o histórico com administrador e
recepção; sem login, as páginas internas e a API continuam protegidas.

## Verificações realizadas

- Testes Go e integração PostgreSQL 17, sem SKIP; `go vet ./...`.
- `npm run lint` e `npm run build`.
- API: dois perfis, validação 422 e autenticação 401.
- Navegador: desktop/celular, histórico aberto pelo cadastro, período, limpeza
  de filtros, vazio e erro; painel sem canceladas e com data do hotel.

As ações de cancelar e alterar pertencem aos checkpoints 3 e 4.
