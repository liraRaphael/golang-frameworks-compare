# ADR-104: Cenário complexo — Pix Black Friday

**Status**: Proposto — cenário não implementado em `api`

**Data**: 2026-10-25

**Tags**: [benchmark, pix, concorrencia, postgres, goroutines, channels, wiremock, outbox]

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
Este ADR especifica o cenário de benchmark Pix Black Friday listado no [README](../../readme.md); não descreve funcionalidade já presente. A API registra atualmente somente `/health` e `/hello`, e ainda não implementa banco de dados, fila, workers ou fluxo de transações PIX. `db/init.sql` está vazio e `mocks` ainda não contém mappings WireMock.

O cenário avalia validação de chave e valor, persistência transacional, aceitação rápida de trabalho assíncrono, processamento concorrente e consulta do resultado. O fluxo é uma simulação de benchmark: não consulta o DICT, não inicia pagamentos reais e não deve usar chaves ou dados pessoais. Os identificadores PIX sintéticos de teste serão os definidos na [ADR-103](103-ingresso-esgotado.md), sem apresentá-los como chaves registradas.

A dependência do provedor será simulada pelo WireMock já declarado no `docker-compose.yaml`. Com delay de 100–120 ms no processamento assíncrono, a resposta inicial HTTP não deve aguardar essa dependência. Os resultados devem separar latência de aceitação e latência de processamento; a segunda inclui aplicação, PostgreSQL, rede local e WireMock, não apenas o framework.

## Decisão

### Schema e persistência
O schema e as fixtures de PIX devem ser provisionados por migration antes de executar API ou benchmark. Reutilizar o conjunto determinístico de 100 chaves EVP sintéticas descrito na ADR-103; se os cenários tiverem bancos isolados, carregar os mesmos valores e a mesma estrutura de chaves. Não duplicar ou gerar chaves reais.

Schema conceitual:

```sql
CREATE TABLE pix_transactions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pix_key TEXT NOT NULL REFERENCES pix_test_keys(pix_key),
    idempotency_key TEXT NOT NULL UNIQUE,
    amount NUMERIC(18, 2) NOT NULL CHECK (amount > 0),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    provider_reference TEXT,
    failure_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    CHECK ((status IN ('pending', 'processing') AND processed_at IS NULL)
        OR (status IN ('completed', 'failed') AND processed_at IS NOT NULL))
);

CREATE INDEX pix_transactions_by_status_created
    ON pix_transactions (status, created_at);

CREATE TABLE pix_queue_guard (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO pix_queue_guard (id) VALUES (TRUE);

CREATE TABLE pix_outbox (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    transaction_id BIGINT NOT NULL UNIQUE
        REFERENCES pix_transactions(id),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'completed')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_until TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX pix_outbox_pending
    ON pix_outbox (available_at, id, locked_until)
    WHERE status IN ('pending', 'processing');
```

Inserir a transação em estado `pending` e seu registro de outbox na mesma transação SQL. O outbox permite que um dispatcher retome trabalho após reinicialização, sem perder um request já aceito entre o commit no banco e o envio ao channel. O dispatcher reivindica cada item com lease (`locked_until`); o worker marca-o `completed` apenas após persistir o estado terminal da transação. Se o processo falhar antes disso, o lease expira e o item pode ser tentado novamente. Atualizações de status, referência do provedor e timestamp final devem ser atômicas e persistidas; guardar mensagens de erro internas apenas em logs/campos controlados, nunca devolvê-las ao cliente.

### Contrato HTTP
- `POST /pix-transactions` recebe JSON com `pix_key` e `amount`, além do header obrigatório `Idempotency-Key`.
- O DTO de body e o DTO de headers declaram campos obrigatórios e restrições estruturais em tags `validate`; a rota tipada converte os dados e executa `ports.Validator` antes do controller. `pix_key` deve ter formato sintético EVP UUID; `amount` deve ser numérico, finito, maior que zero, limitado a duas casas decimais e dentro do limite definido. Tags padrão devem ser usadas quando suficientes e regras adicionais registradas em `api/app/validation/validation.go`.
- O header `Idempotency-Key` deve ser obrigatório e não vazio, sendo validado antes do controller. Formato estrutural inválido retorna HTTP 400 sem criar trabalho.
- Repetir a mesma chave de idempotência com o mesmo payload deve devolver a transação já criada, sem criar novo trabalho; a unicidade é persistida no banco. Reutilizar a chave com payload diferente retorna HTTP 409. Em corridas concorrentes pela mesma chave, tratar a violação da constraint consultando a transação vencedora e comparando o payload antes de responder.
- `pix_key` deve ser uma das chaves sintéticas ativas carregadas como fixture; rejeitar chaves válidas em formato, porém desconhecidas/inativas, antes de admitir trabalho. Essa verificação depende das fixtures persistidas e pertence ao serviço, não ao validador estrutural. Não consultar DICT nem validar titularidade real.
- JSON malformado ou valor inválido retorna HTTP 400. Chave bem formada mas ausente/inativa retorna HTTP 422. Esses casos não criam transação nem outbox.
- Em sucesso de admissão, persistir `pending` e outbox e retornar HTTP 202 com `id`, `status` e `created_at`. HTTP 202 significa aceito para processamento, não pagamento concluído.
- O use case/serviço retorna somente resultado tipado e `error`; não cria `responses.Response`, define status HTTP ou serializa JSON. O controller traduz resultados de negócio conhecidos (admissão 202, conflito de idempotência 409, chave inativa 422, capacidade esgotada 503 e consulta inexistente 404) em `responses.Response` com status; falhas de validação estrutural e erros técnicos são encaminhados ao handler conforme seus mapeamentos. O adapter serializa e escreve a response, sem lógica de status no caso de sucesso.
- Quando a capacidade de processamento estiver esgotada, retornar HTTP 503 sem criar transação. Falhas na persistência antes da confirmação retornam erro explícito e não devem produzir resposta de aceitação.
- `GET /pix-transactions/{id}` retorna HTTP 200 com `id`, `pix_key` mascarada, `amount`, `status` e timestamps pertinentes; retorna HTTP 404 se a transação não existir. Não incluir dados sensíveis ou falhas internas detalhadas.
- Erros inesperados retornam HTTP 500 sem expor detalhes internos. Registrar mapeamentos explícitos para erros de entrada, chave ausente, fila saturada e falha interna; não depender do fallback HTTP 500 do handler atual (ADR-003).

### Organização na aplicação
Os diretórios listados são propostas para implementação futura, não arquivos já presentes:

- `api/core/domain/requests/pix.go`: DTOs de body e parâmetros HTTP com tags JSON e `validate` para campos/restrições estruturais.
- `api/core/domain/responses/pix.go`: resposta de admissão e consulta, com estados permitidos.
- `api/core/ports/service.go`: interfaces consumidas pelos controllers para submissão e consulta.
- `api/core/ports`: contratos pequenos para persistência e cliente do provedor, sem tipos de framework ou SQL concreto.
- `api/infra/router` e `api/app/validation`: converter body/headers e executar as tags `validate` e regras customizadas registradas em `validation.go`, antes do controller.
- `api/app/controller/pix.go`: chamar o serviço e criar `responses.Response` com status apropriado para resultados de negócio conhecidos; não validar DTOs nem serializar/escrever HTTP. Erros de validação estrutural e falhas técnicas são resolvidos pelo handler por mapeamentos explícitos.
- `api/app/service/pix.go`: orquestrar validações dependentes de estado, admissão transacional, leitura do status e ciclo de vida do worker/dispatcher conforme os contratos; retornar corpo tipado e `error`, sem construir response HTTP nem escolher status.
- `api/infra/database`: implementar repositório PostgreSQL, transações e persistência do outbox.
- `api/infra/clients`: implementar o client HTTP do provedor mock usando o contrato `ports.Client` e timeout finito.
- `api/infra/workers` (ou pacote de infraestrutura equivalente): executar dispatcher, fila e workers, sem colocar lógica em adapters de framework.
- `api/main.go`: compor manualmente pool SQL, repositório, client, serviço, workers e controller por construtores, seguindo ADR-010. Acrescentar configuração de URL/timeout do provedor e de número de workers/capacidade da fila; o `AppConfig` atual não as declara. Como a configuração existente processa o prefixo `APP_`, alinhar o Compose para definir `APP_WIREMOCK_URL` ou declarar explicitamente outro prefixo; hoje o Compose define `WIREMOCK_URL`, que não será carregado automaticamente pelo `envconfig.Process("APP", ...)`.

O controller não deve acessar o banco, chamar WireMock nem criar uma goroutine por request diretamente. A API deve ter um ciclo de vida do worker gerido pelo bootstrap; não iniciar workers novos a cada chamada HTTP.

Validação estrutural de DTO (presença, formato e limites representáveis por tags/regras de campo) acontece uma única vez na rota tipada antes do controller. Verificações de idempotência, existência de chave ativa, capacidade, transições e estado da transação são regras de negócio/persistência do serviço e não devem ser movidas para o controller nem confundidas com validação de campos.

### Goroutines, fila e consistência
- Executar um número configurável e fixo de goroutines workers, cada uma consumindo identificadores de transação de um channel limitado. Usar o mesmo número de workers e capacidade em todos os frameworks comparados.
- Um dispatcher reivindica entradas `pending` ou com lease expirado do outbox e as envia ao channel limitado. Usar locking transacional (`FOR UPDATE SKIP LOCKED` ou mecanismo equivalente) e `locked_until` para evitar reivindicações concorrentes; o worker conclui o outbox somente depois de persistir o estado terminal. Entradas com lease expirado devem poder ser reenviadas após falha/reinício.
- Limitar a admissão com uma capacidade explícita para trabalho pendente/em processamento. Na mesma transação de admissão, bloquear a linha única de `pix_queue_guard`, contar transações `pending` e `processing` e comparar com a capacidade configurada; sob capacidade, inserir transação e outbox, e então fazer commit. Se estiver cheio, retornar 503 sem persistir. Esse lock de banco é curto e serializa apenas a decisão de admissão; channel limitado e mutex local sozinhos não impõem limite distribuído.
- Persistir transação e outbox numa única transação SQL antes de responder 202. Aplicar idempotência antes de consumir capacidade; repetição com a mesma chave e payload retorna a transação original. Comparar payload mesmo em concorrência e rejeitar reutilização incompatível. Se não houver capacidade, responder 503 sem persistência; se a gravação falhar, não retornar 202.
- A transição de `pending` para `processing` e para estado terminal deve ser condicionada ao estado esperado (compare-and-set/atualização condicional) para evitar processamento duplicado por workers concorrentes.
- O worker chama o provedor mock fora de qualquer transação SQL e sem manter mutex global; depois persiste `completed` ou `failed`. Recusa de negócio do provedor é terminal `failed`; falhas transitórias podem ser repetidas com backoff limitado, incremento de `attempts` e `available_at`. Após o máximo configurado, persistir falha terminal para não manter trabalho preso indefinidamente. A chamada deve ter timeout e usar o `transaction_id` como chave de idempotência.
- A entrega pode ser repetida após falha/reinício. Enviar `transaction_id` como chave idempotente ao provedor simulado para que retries representem o mesmo processamento. Resultados incertos de rede devem ser registrados e reconciliados; não presumir sucesso silenciosamente. Um mapping estático WireMock não comprova idempotência por si só: validar que a aplicação reutiliza a chave em todas as tentativas e separar essa garantia do comportamento do mock. Como a requisição já recebeu 202, falha posterior é observada como estado `failed` na consulta, não como uma nova resposta HTTP 502.
- Usar `sync.WaitGroup` para acompanhar a vida dos workers e aguardar encerramento; `context.Context` de aplicação controla dispatcher/worker e `context` por chamada limita cada operação externa. O contexto da requisição HTTP não pode cancelar trabalho já aceito após retornar 202.
- Criar spans/logs do processamento em background vinculados ao identificador da transação (e, quando disponível, ao contexto de trace propagado/persistido), sem reutilizar o contexto cancelável do request. A instrumentação atual não mede ainda chamadas externas nem métricas de fila (ADR-005).
- Usar mutex/atomic apenas para estado em memória que precise de coordenação. A correção e a recuperação do estado persistido não podem depender exclusivamente de primitivas `sync`.
- No shutdown, parar novas admissões, cancelar o dispatcher de modo que não envie para channel fechado, aguardar ou cancelar workers até um timeout de encerramento. Trabalho ainda pendente deve permanecer no banco/outbox para recuperação na próxima inicialização.

### WireMock e mocks
Guardar mappings em `./mocks/mappings/`; o Compose atual monta `./mocks` em `/home/wiremock` e publica WireMock na rede do Compose. A aplicação deve ler a URL (`http://wiremock:8080` no Compose) e o timeout da configuração injetada; não codificar URL/timeout no controller.

O worker chama `POST /pix-provider/transfers` com `transaction_id`, chave sintética e valor. O mapping de sucesso deve responder deterministicamente e simular latência uniforme de 100–120 ms:

```json
{
  "request": {
    "method": "POST",
    "urlPath": "/pix-provider/transfers",
    "bodyPatterns": [
      { "matchesJsonPath": "$.transaction_id" },
      { "matchesJsonPath": "$.pix_key" },
      { "matchesJsonPath": "$.amount" }
    ]
  },
  "response": {
    "status": 200,
    "jsonBody": {
      "status": "completed",
      "provider_reference": "wiremock-transaction"
    },
    "headers": { "Content-Type": "application/json" },
    "delayDistribution": {
      "type": "uniform"
    }
  }
}
```

Adicionar mappings específicos de teste para recusa do provedor e erro/timeout simulado, mantendo a distribuição de atraso documentada e igual entre frameworks. O mock não deve realizar chamadas externas, movimentar dinheiro, armazenar segredos nem usar dados pessoais. Em testes unitários, usar fakes injetados; em testes de integração e benchmark, usar WireMock local. A versão e os mappings do mock devem permanecer fixos durante uma rodada.

### Testes, benchmark e observabilidade
- Testes unitários cobrem validação de request, admissão e mapeamento de estados sem PostgreSQL ou WireMock.
- Testes de integração usam PostgreSQL isolado e WireMock local; cobrir sucesso, recusa, timeout/erro do provedor, banco indisponível, reentrega do outbox, idempotência, reinício com pendências e saturação da capacidade.
- Testes concorrentes devem iniciar várias submissões em paralelo, sincronizadas por uma barreira, e usar `sync.WaitGroup` para coletar resultados. Verificar que cada HTTP 202 corresponde a uma transação persistida e consultável, que toda transação eventualmente chega a estado terminal ou permanece recuperável, e que retries não duplicam o efeito.
- Repor as mesmas fixtures e estado inicial entre execuções; preparação e limpeza não entram na janela cronometrada.
- Comparar frameworks com a mesma versão/configuração de PostgreSQL, quantidade de workers, capacidade de fila, pool SQL, timeout, mappings, delay, carga e limites de CPU/memória.
- Separar p95/latência do POST de admissão, tempo de espera em fila, duração do processamento, latência do mock e duração SQL. Reportar throughput aceito, taxas de 202/400/404/422/503/5xx, estados finais, falhas e idade das mensagens pendentes. Respostas 202 não são pagamentos concluídos.
- A ADR-005 registra que a configuração atual não instrumenta métricas de requisição nem chamadas externas. Implementar instrumentos antes de reportar métricas de fila/worker; logs e traces não substituem contadores e histogramas.
- A rodada com WireMock mede o caminho completo (framework + aplicação + PostgreSQL + WireMock); se houver uma rodada sem mock para isolar o custo de framework, reportá-la separadamente.

## Consequências
- A resposta HTTP rápida fica desacoplada da duração do provedor simulado, e o estado persistido permite consultar o resultado assíncrono.
- Goroutines e channel limitado exercitam processamento concorrente, enquanto PostgreSQL/outbox fornecem durabilidade e recuperação.
- O padrão outbox reduz o risco de aceitar uma transação no banco e perder o evento antes de enfileirar.
- A capacidade distribuída e a idempotência precisam ser especificadas além de um mutex local para evitar aceitação excessiva ou duplicação em reinícios/replicas.
- WireMock fornece dependência controlada sem pagamentos reais; o delay deve ser considerado ao interpretar o tempo total de processamento.
- O cenário continua proposto e só deve compor resultados comparativos depois que schema/migrations, workers, mocks, métricas e testes estiverem implementados.

## Referências
1. [README — cenários de benchmark](../../readme.md)
2. [ADR-002: Cliente](002-cliente.md)
3. [ADR-003: Handler](003-handler.md)
4. [ADR-005: Observability](005-observability.md)
5. [ADR-006: Estratégia de benchmarking e testes](006-benchmarking-e-testes.md)
6. [ADR-010: Injeção de Dependência (DI)](010-injecao-dependencia.md)
7. [ADR-011: Gestão de Configuração](011-gestao-configuracao.md)
8. [ADR-013: Implementação dos frameworks web comparados](013-implementacao-frameworks-web.md)
9. [ADR-103: Ingresso Esgotado e fixtures PIX sintéticas](103-ingresso-esgotado.md)
10. [Go package sync](https://pkg.go.dev/sync)
11. [Go package context](https://pkg.go.dev/context)
12. [WireMock — fault simulation and delays](https://wiremock.org/docs/simulating-faults/)
