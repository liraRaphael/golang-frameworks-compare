# ADR-003: Handler

**Status**: Aceito

**Data**: 2026-07-19

**Tags**: [handler, clean-arch, response, error, success, mapeamento, json]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
Para garantir resposta consistente e rastreável, a aplicação precisa de uma camada responsável por padronizar sucesso, erro e logging. Essa camada não deve depender diretamente do framework web nem do protocolo de transporte, mas precisa ser capaz de transformar resultados de use cases em respostas previsíveis para o cliente.

## Decisão
O handler está implementado em `api/app/handler` como `standardHandler`, exposto pela interface `ports.Handler`.

### Biblioteca recomendada
A implementação inicial utilizará a biblioteca padrão `encoding/json` para serializar as respostas padronizadas, mantendo o fluxo simples e compatível com os frameworks comparados. A escolha foi feita para reduzir dependências extras e preservar a neutralidade da camada de handler.

### Comportamento esperado
- `Handle` recebe `ports.Context`, status de sucesso, saída e headers/cookies.
- O `ports.Context` é a abstração de borda do domínio e compõe o `context.Context` do Go, permitindo que adaptadores de framework convertam o contexto nativo da requisição em um contrato neutro para a camada de portas.
- Se `result` for um `error`, o handler identificará o tipo e/ou mensagem do erro e aplicará a estratégia adequada de mapeamento.
- Se `result` não for um erro, o handler interpretará como sucesso e retornará o payload de resposta já preparado pelo caso de uso ou presenter.
- Os adapters de framework usam `ResolveError` quando parsing ou controller retornam erro. Os controllers atuais devolvem `responses.Response` diretamente em caso de sucesso; `Handle` existe no contrato, mas não é usado nesse fluxo.

### Resposta padrão
`core/domain/responses.ErrorResponse` define o payload de erro:

```json
{
  "Body": {
    "codigo": "internal_error",
    "mensagem": "Erro interno do servidor não mapeado"
  },
  "StatusCode": 500,
  "Headers": null
}
```

### Estratégia de mapeamento implementada
- `ErrorResponseFuncType` está definido em `core/ports` e retorna `responses.Response[any, any]`.
- A assinatura do callback é:

```go
type ErrorResponseFuncType func(ctx ports.Context, err error) responses.Response[any, any]
```

- O handler consulta primeiro a mensagem exata e depois o tipo concreto (`reflect.Type`).
- Sem correspondência, devolve status 500, código `internal_error` e mensagem `Erro interno do servidor não mapeado`.
- `main.go` não registra mapeamentos. Erros de parsing/validação/controller resultam, portanto, no fallback 500 neste handler.

### Interface do handler
```go
type Handler interface {
    Handle(ctx ports.Context, successStatusCode int, output any, headers domain.HttpParamsType, cookies domain.HttpParamsType) responses.Response[any, any]
    RegisterByMessage(message string, fn ports.ErrorResponseFuncType)
    RegisterByType(err error, fn ports.ErrorResponseFuncType)
    ResolveError(ctx ports.Context, err error) responses.Response[any, any]
}
```

A interface está definida em `core/ports`. O contrato permanece independente dos tipos concretos de framework, embora `ports` tenha outras dependências diretas de `infra`.

### Logging
- O handler atual não registra logs. Os adapters de framework fazem log de requisições recebidas e falhas de escrita/serialização por meio do logger do contexto.

### Integração com o framework
- O handler não deve conhecer diretamente o framework escolhido.
- Ele deverá retornar uma estrutura de resposta pronta para o framework serializar.
- O framework, por sua vez, será responsável por escrever a resposta ao cliente com o status e o body corretos.

## Consequências
- As respostas da API passam a ser uniformes entre os frameworks comparados.
- Erros tornam-se mais fáceis de rastrear e mapear.
- Os adapters convertem a resposta padronizada para JSON e a escrevem com status, headers e cookies.

## Referências
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
