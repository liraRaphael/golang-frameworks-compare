# ADR-103: Cenário complexo — Ingresso Esgotado

**Status**: Proposto — cenário não implementado em `api`

**Data**: 2026-10-25

**Tags**: [benchmark, concorrencia, ingressos, postgres, transacoes, migration, wiremock, pix]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001a: Detalhamento da camada Core](001a-camada-core.md)
- [ADR-001b: Detalhamento da camada App](001b-camada-adapter.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)
- [ADR-002: Cliente](002-cliente.md)
- [ADR-003: Handler](003-handler.md)
- [ADR-004: Roteamento](004-roteamento.md)
- [ADR-005: Observability](005-observability.md)
- [ADR-006: Estratégia de benchmarking e testes](006-benchmarking-e-testes.md)
- [ADR-009: Padrões de Request, Response e HttpParam no Core](009-request-response-core.md)
- [ADR-010: Injeção de Dependência (DI)](010-injecao-dependencia.md)
- [ADR-011: Gestão de Configuração](011-gestao-configuracao.md)
- [ADR-012: Mecanismo de Validação de Dados](012-validacao-campo.md)
- [ADR-013: Implementação dos frameworks web comparados](013-implementacao-frameworks-web.md)

## Contexto
Este ADR especifica o cenário de benchmark Ingresso Esgotado listado no [README](../../readme.md); não descreve funcionalidade presente. A API registra atualmente somente `/health` e `/hello`, e não possui banco de dados nem este fluxo implementado. `db/init.sql` está vazio e `mocks` não contém mappings WireMock.

O cenário mede validação de uma compra, dependência externa simulada, contenção por estoque limitado, transações SQL e comportamento sob concorrência. A persistência PostgreSQL é a autoridade para o estoque e para a unicidade da venda. WireMock fornece uma resposta previsível com latência controlada para uma checagem simulada de elegibilidade do comprador; não realiza pagamento real.

Os resultados devem ser interpretados de acordo com o que realmente é medido: quando a chamada ao WireMock estiver dentro da requisição, a latência inclui aplicação, rede e mock, e não representa isoladamente o custo do framework. O mesmo mock, delay, banco, seed e perfil de carga devem ser usados para todos os adapters.

## Decisão

### Responsabilidades por camada
- `api/app/controller`: validar o request, chamar o serviço de compra e traduzir o resultado para a response HTTP; não executar SQL nem chamar WireMock diretamente.
- `api/app/service`: orquestrar a elegibilidade simulada e a compra; depender de contratos, não de PostgreSQL ou framework concretos.
- `api/core/ports`: declarar os contratos necessários para a compra/persistência e para o serviço consumido pelo controller. Não há hoje uma porta de repositório implementada.
- `api/infra`: implementar a persistência PostgreSQL e o client HTTP injetados manualmente por construtor em `api/main.go`, conforme ADR-010.
- O contrato comum do adapter não garante equivalência completa de binding, path params ou escrita de respostas; validar o mesmo comportamento HTTP nos adapters conforme ADR-013.

### Migration inicial e schema
O schema e o seed determinístico serão criados em `db/init.sql`, tratado como a migration inicial do cenário. A migration cria as tabelas e restrições, insere o evento/estoque inicial e registra as 100 chaves PIX sintéticas conhecidas para os testes.

O arquivo é uma migration de bootstrap de banco novo; não deve ser executado novamente sobre um banco já inicializado. Para evolução posterior, criar migrations versionadas e ordenadas em vez de editar silenciosamente uma migration já aplicada. A migration deve ser aplicada antes de iniciar a API ou o benchmark; falha de schema/seed deve interromper a preparação explicitamente.

**Integração com Docker Compose:** a configuração atual monta `./db` em `/var/lib/postgresql/data`, que é o diretório de dados do PostgreSQL e não o diretório de scripts de inicialização. Para executar `db/init.sql` em uma instância nova, separar o volume persistente do banco e montar o arquivo somente leitura em `/docker-entrypoint-initdb.d/001-init.sql`. Scripts dessa pasta são executados apenas durante a inicialização de um diretório de dados vazio; mudanças subsequentes exigem execução explícita da migration ou ferramenta de migração.

Schema proposto:

```sql
CREATE TABLE events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE pix_test_keys (
    pix_key TEXT PRIMARY KEY,
    fixture_number SMALLINT NOT NULL UNIQUE
        CHECK (fixture_number BETWEEN 1 AND 100),
    key_type TEXT NOT NULL DEFAULT 'evp'
        CHECK (key_type = 'evp'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tickets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id),
    ticket_code TEXT NOT NULL UNIQUE,
    price NUMERIC(12, 2) NOT NULL CHECK (price >= 0),
    status TEXT NOT NULL DEFAULT 'available'
        CHECK (status IN ('available', 'sold')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sold_at TIMESTAMPTZ,
    CHECK ((status = 'available' AND sold_at IS NULL)
        OR (status = 'sold' AND sold_at IS NOT NULL))
);

CREATE INDEX tickets_available_by_event
    ON tickets (event_id, id) WHERE status = 'available';

CREATE TABLE ticket_purchases (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ticket_id BIGINT NOT NULL UNIQUE REFERENCES tickets(id),
    buyer_pix_key TEXT NOT NULL REFERENCES pix_test_keys(pix_key),
    status TEXT NOT NULL DEFAULT 'confirmed'
        CHECK (status = 'confirmed'),
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ticket_purchases_by_time
    ON ticket_purchases (purchased_at);
```

O seed cria um evento e uma quantidade configurável de ingressos disponíveis, com códigos/preços determinísticos. O identificador do evento, códigos, preços, quantidade e estado inicial devem ser iguais entre frameworks e registrados com os resultados do benchmark. `ticket_purchases.ticket_id UNIQUE` é uma proteção adicional contra a venda duplicada do mesmo ingresso.

### Chaves PIX sintéticas do cenário
O conjunto terá exatamente 100 chaves PIX do tipo EVP em formato UUID, registradas na tabela `pix_test_keys`. São fixtures sintéticas e conhecidas pelo teste; não pertencem a pessoas ou contas reais, não devem ser apresentadas como chaves registradas no DICT e nunca devem ser usadas para pagamentos. A validação de formato não comprova titularidade nem existência de uma chave PIX real.

Lista determinística a inserir na migration:

```sql
INSERT INTO pix_test_keys (fixture_number, pix_key) VALUES
    (1,   '00000000-0000-4000-8000-000000000001'),
    (2,   '00000000-0000-4000-8000-000000000002'),
    (3,   '00000000-0000-4000-8000-000000000003'),
    (4,   '00000000-0000-4000-8000-000000000004'),
    (5,   '00000000-0000-4000-8000-000000000005'),
    (6,   '00000000-0000-4000-8000-000000000006'),
    (7,   '00000000-0000-4000-8000-000000000007'),
    (8,   '00000000-0000-4000-8000-000000000008'),
    (9,   '00000000-0000-4000-8000-000000000009'),
    (10,  '00000000-0000-4000-8000-000000000010'),
    (11,  '00000000-0000-4000-8000-000000000011'),
    (12,  '00000000-0000-4000-8000-000000000012'),
    (13,  '00000000-0000-4000-8000-000000000013'),
    (14,  '00000000-0000-4000-8000-000000000014'),
    (15,  '00000000-0000-4000-8000-000000000015'),
    (16,  '00000000-0000-4000-8000-000000000016'),
    (17,  '00000000-0000-4000-8000-000000000017'),
    (18,  '00000000-0000-4000-8000-000000000018'),
    (19,  '00000000-0000-4000-8000-000000000019'),
    (20,  '00000000-0000-4000-8000-000000000020'),
    (21,  '00000000-0000-4000-8000-000000000021'),
    (22,  '00000000-0000-4000-8000-000000000022'),
    (23,  '00000000-0000-4000-8000-000000000023'),
    (24,  '00000000-0000-4000-8000-000000000024'),
    (25,  '00000000-0000-4000-8000-000000000025'),
    (26,  '00000000-0000-4000-8000-000000000026'),
    (27,  '00000000-0000-4000-8000-000000000027'),
    (28,  '00000000-0000-4000-8000-000000000028'),
    (29,  '00000000-0000-4000-8000-000000000029'),
    (30,  '00000000-0000-4000-8000-000000000030'),
    (31,  '00000000-0000-4000-8000-000000000031'),
    (32,  '00000000-0000-4000-8000-000000000032'),
    (33,  '00000000-0000-4000-8000-000000000033'),
    (34,  '00000000-0000-4000-8000-000000000034'),
    (35,  '00000000-0000-4000-8000-000000000035'),
    (36,  '00000000-0000-4000-8000-000000000036'),
    (37,  '00000000-0000-4000-8000-000000000037'),
    (38,  '00000000-0000-4000-8000-000000000038'),
    (39,  '00000000-0000-4000-8000-000000000039'),
    (40,  '00000000-0000-4000-8000-000000000040'),
    (41,  '00000000-0000-4000-8000-000000000041'),
    (42,  '00000000-0000-4000-8000-000000000042'),
    (43,  '00000000-0000-4000-8000-000000000043'),
    (44,  '00000000-0000-4000-8000-000000000044'),
    (45,  '00000000-0000-4000-8000-000000000045'),
    (46,  '00000000-0000-4000-8000-000000000046'),
    (47,  '00000000-0000-4000-8000-000000000047'),
    (48,  '00000000-0000-4000-8000-000000000048'),
    (49,  '00000000-0000-4000-8000-000000000049'),
    (50,  '00000000-0000-4000-8000-000000000050'),
    (51,  '00000000-0000-4000-8000-000000000051'),
    (52,  '00000000-0000-4000-8000-000000000052'),
    (53,  '00000000-0000-4000-8000-000000000053'),
    (54,  '00000000-0000-4000-8000-000000000054'),
    (55,  '00000000-0000-4000-8000-000000000055'),
    (56,  '00000000-0000-4000-8000-000000000056'),
    (57,  '00000000-0000-4000-8000-000000000057'),
    (58,  '00000000-0000-4000-8000-000000000058'),
    (59,  '00000000-0000-4000-8000-000000000059'),
    (60,  '00000000-0000-4000-8000-000000000060'),
    (61,  '00000000-0000-4000-8000-000000000061'),
    (62,  '00000000-0000-4000-8000-000000000062'),
    (63,  '00000000-0000-4000-8000-000000000063'),
    (64,  '00000000-0000-4000-8000-000000000064'),
    (65,  '00000000-0000-4000-8000-000000000065'),
    (66,  '00000000-0000-4000-8000-000000000066'),
    (67,  '00000000-0000-4000-8000-000000000067'),
    (68,  '00000000-0000-4000-8000-000000000068'),
    (69,  '00000000-0000-4000-8000-000000000069'),
    (70,  '00000000-0000-4000-8000-000000000070'),
    (71,  '00000000-0000-4000-8000-000000000071'),
    (72,  '00000000-0000-4000-8000-000000000072'),
    (73,  '00000000-0000-4000-8000-000000000073'),
    (74,  '00000000-0000-4000-8000-000000000074'),
    (75,  '00000000-0000-4000-8000-000000000075'),
    (76,  '00000000-0000-4000-8000-000000000076'),
    (77,  '00000000-0000-4000-8000-000000000077'),
    (78,  '00000000-0000-4000-8000-000000000078'),
    (79,  '00000000-0000-4000-8000-000000000079'),
    (80,  '00000000-0000-4000-8000-000000000080'),
    (81,  '00000000-0000-4000-8000-000000000081'),
    (82,  '00000000-0000-4000-8000-000000000082'),
    (83,  '00000000-0000-4000-8000-000000000083'),
    (84,  '00000000-0000-4000-8000-000000000084'),
    (85,  '00000000-0000-4000-8000-000000000085'),
    (86,  '00000000-0000-4000-8000-000000000086'),
    (87,  '00000000-0000-4000-8000-000000000087'),
    (88,  '00000000-0000-4000-8000-000000000088'),
    (89,  '00000000-0000-4000-8000-000000000089'),
    (90,  '00000000-0000-4000-8000-000000000090'),
    (91,  '00000000-0000-4000-8000-000000000091'),
    (92,  '00000000-0000-4000-8000-000000000092'),
    (93,  '00000000-0000-4000-8000-000000000093'),
    (94,  '00000000-0000-4000-8000-000000000094'),
    (95,  '00000000-0000-4000-8000-000000000095'),
    (96,  '00000000-0000-4000-8000-000000000096'),
    (97,  '00000000-0000-4000-8000-000000000097'),
    (98,  '00000000-0000-4000-8000-000000000098'),
    (99,  '00000000-0000-4000-8000-000000000099'),
    (100, '00000000-0000-4000-8000-000000000100');
```

### Contrato e compra
- `POST /ticket-purchases` recebe `ticket_id` e `buyer_pix_key`.
- Validar formato do payload e confirmar que `buyer_pix_key` pertence ao conjunto de teste. Não aceitar chaves arbitrárias ou fazer consulta ao DICT. Payload malformado retorna HTTP 400; uma chave com formato válido que não pertença às fixtures retorna HTTP 422, sem iniciar compra.
- Para chamadas válidas, chamar o mock de elegibilidade antes de abrir a transação de compra. Essa chamada não movimenta dinheiro e não autoriza pagamento; ela existe para simular uma dependência com latência controlada.
- Elegibilidade negada retorna HTTP 422 e não cria compra.
- Em caso de elegibilidade aprovada, iniciar uma transação SQL, selecionar o ingresso com `SELECT ... FOR UPDATE`, confirmar que existe e está disponível, inserir a compra e marcar o ingresso como `sold`. Confirmar as alterações no mesmo commit; em erro, executar rollback e retornar erro explícito.
- Ingresso inexistente retorna HTTP 404. Ingresso já vendido retorna HTTP 409. Erro/timeout do serviço mock retorna HTTP 502 e não deve deixar compra confirmada.
- `GET /ticket-purchases/{id}` consulta o estado persistido e retorna HTTP 404 quando a compra não existe.
- A resposta de sucesso contém identificador da compra, ingresso, estado confirmado e timestamp; nunca expor detalhes internos do banco ou dados pessoais.

### Sincronização e concorrência
- A consistência do estoque é garantida pela transação e pelo bloqueio da linha no PostgreSQL, com a constraint `UNIQUE` em `ticket_purchases.ticket_id` como proteção adicional.
- Não envolver toda a rota em um `sync.Mutex` global: isso serializaria compras de ingressos diferentes, reduziria artificialmente o throughput e não coordenaria processos ou réplicas.
- Não manter mutex ou transação SQL abertos durante a chamada ao WireMock, cuja latência é intencionalmente de 100–120 ms.
- Usar primitivas de `sync` somente onde há estado compartilhado em memória que realmente exija coordenação. No teste concorrente, usar `sync.WaitGroup` para aguardar todas as goroutines e uma barreira (por exemplo, channel) para liberar requisições simultaneamente. Proteger contadores/mapas mutáveis compartilhados com `sync.Mutex` ou `sync/atomic`, conforme o tipo de dado.
- Testar simultaneamente compras do mesmo ingresso e de ingressos diferentes. Total de compras confirmadas nunca pode exceder estoque; um ingresso não pode aparecer em mais de uma compra.
- Configurar carga concorrente maior que o estoque e registrar quantas respostas foram sucesso, conflito, não encontrado e falha externa.

### WireMock e delays
Os mappings serão mantidos em `./mocks/mappings/` e servidos pelo WireMock já declarado no `docker-compose.yaml`. A chamada simulada será `POST /ticket-eligibility/check`; todos os frameworks devem usar o mesmo `WIREMOCK_URL`, stub e timeout.

Mapping de aprovação com atraso uniforme por chamada entre 100 e 120 ms:

```json
{
  "request": {
    "method": "POST",
    "urlPath": "/ticket-eligibility/check",
    "bodyPatterns": [
      { "matchesJsonPath": "$.buyer_pix_key" },
      { "matchesJsonPath": "$.ticket_id" }
    ]
  },
  "response": {
    "status": 200,
    "jsonBody": { "eligible": true },
    "headers": { "Content-Type": "application/json" },
    "delayDistribution": {
      "type": "uniform",
      "lower": 100,
      "upper": 120
    }
  }
}
```

Criar também um mapping controlado para elegibilidade negada, sem delay diferente dos 100–120 ms, para validar o caminho de recusa sem depender de um serviço real. O endpoint mock não deve armazenar dados pessoais nem executar cobrança. Os valores retornados devem ser estáveis, e os testes não devem exigir acesso à Internet.

O projeto já monta `./mocks` em `/home/wiremock`; a implementação deverá adicionar os mappings no layout esperado pelo WireMock. A configuração atual não garante montagem em modo read-only; para execução do benchmark, preferir volume somente leitura e não alterar mappings enquanto houver requisições em andamento.

### Testes, benchmark e preparação
- Testes unitários de regra e controller usam doubles/fakes injetados, sem banco ou WireMock real.
- Testes de integração usam PostgreSQL isolado e WireMock local; cobrir sucesso, elegibilidade negada, timeout/5xx do mock, ingresso inexistente, ingresso já vendido e rollback em falhas.
- Ter uma migration/seed reproduzível para teste e uma rotina que restaure o mesmo estado inicial entre execuções. A preparação e limpeza não entram na janela cronometrada.
- Confirmar antes e depois de cada rodada as invariantes: quantidade de ingressos vendidos = compras confirmadas, cada ingresso tem no máximo uma compra e saldo disponível + vendido = estoque inicial.
- Rodar o mesmo número de workers/requisições, estoque, chaves PIX, distribuição de IDs, delay e limites de CPU/memória em todos os frameworks.
- Reportar status HTTP, sucessos, conflitos, erros do mock/banco, latência média e p95, throughput e latência da dependência separadamente. Não classificar 4xx/5xx como sucesso.
- Como a chamada WireMock está no caminho medido, declarar que os números representam o fluxo completo (framework + aplicação + PostgreSQL + WireMock). Se o objetivo for isolar apenas o overhead do framework, rodar também um experimento sem a chamada externa, sem misturar os dois resultados.

## Consequências
- A migration `db/init.sql` deixa schema e fixtures iniciais explícitos e repetíveis; a execução precisa ser conectada corretamente ao ciclo de inicialização do PostgreSQL.
- As 100 chaves PIX de fixture são sintéticas, estáveis e sem associação com contas reais.
- O banco, e não um mutex local, coordena estoque entre requisições e réplicas.
- WireMock fornece latência controlada de 100–120 ms sem cobranças nem dependência de serviço financeiro real.
- O cenário permite comparar contenção e transações, desde que banco, mock, carga e recursos sejam iguais entre frameworks.
- Este cenário continua proposto; deve ser implementado, integrado e testado antes de compor resultados comparativos.

## Referências
1. [README — cenários de benchmark](../../readme.md)
2. [ADR-002: Cliente](002-cliente.md)
3. [ADR-003: Handler](003-handler.md)
4. [ADR-006: Estratégia de benchmarking e testes](006-benchmarking-e-testes.md)
5. [ADR-010: Injeção de Dependência (DI)](010-injecao-dependencia.md)
6. [ADR-013: Implementação dos frameworks web comparados](013-implementacao-frameworks-web.md)
7. [Go package sync](https://pkg.go.dev/sync)
8. [WireMock — delay distribution](https://wiremock.org/docs/simulating-faults/)
9. [PostgreSQL — database initialization scripts](https://hub.docker.com/_/postgres)
