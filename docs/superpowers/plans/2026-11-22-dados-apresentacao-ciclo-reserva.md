# Dados de Apresentação do Ciclo de Reserva — Plano de Implementação

> **For agentic workers:** Implementar inline, seguindo os passos abaixo. Cada tarefa deve ser conferida antes de avançar.

**Goal:** Preparar dados locais repetíveis para demonstrar check-in e checkout no sistema sem apagar dados operacionais que não pertencem à demonstração.

**Architecture:** Um script SQL separado cria dois quartos e hóspedes identificados como demonstração, uma reserva pronta para check-in e uma estadia iniciada há duas noites. O script remove e recria apenas os registros com IDs dedicados à demonstração; a documentação explica como carregar e apresentar o cenário.

**Tech Stack:** PostgreSQL, SQL de seed, documentação Markdown.

**Spec:** `docs/superpowers/specs/2026-11-22-checkin-checkout-design.md` e as regras do ciclo operacional já implementadas.

## Restrições

- Não executar o seed destrutivo `database/seed_demo.sql` como parte da preparação deste cenário.
- Não apagar hóspedes, reservas, eventos ou quartos fora dos IDs exclusivos `DEMO`.
- A reserva pronta para check-in deve permitir a ação na data de execução.
- A reserva em hospedagem deve permitir checkout com duas noites reais e usar a diária contratada.

---

### Tarefa 1: Criar o seed isolado e repetível

**Arquivos:** criar `database/seed_presentation.sql`.

- [x] Inserir categoria, quartos e hóspedes com IDs exclusivos de apresentação.
- [x] Exigir usuário administrador ou recepção existente para auditar a estadia iniciada.
- [x] Limpar apenas eventos, ocupação e reservas de IDs exclusivos, em uma transação.
- [x] Criar reserva pronta para check-in e reserva já em hospedagem com duas noites efetivas.

### Tarefa 2: Documentar a carga e o roteiro

**Arquivos:** modificar `database/README.md`; criar `docs/ROTEIRO_APRESENTACAO_CICLO_RESERVA.md`.

- [x] Explicar que o script não limpa a base e depende das migrações 001–008 e usuário de demonstração.
- [x] Documentar login de desenvolvimento, cenários, valor esperado e sequência de telas.

### Tarefa 3: Verificar o artefato

- [x] Revisar o SQL contra as constraints do schema e conferir o diff sem erros de whitespace.
- [ ] Executar o script duas vezes em uma base PostgreSQL descartável; os serviços locais estavam parados nesta sessão.
- [ ] Conferir em execução que a primeira reserva fica confirmada e a segunda em hospedagem, com quarto ocupado e total de duas diárias.
