# ADR-003: Handler

**Status**: Aceito

**Data**: 2026-07-19

**Tags**: [handler, clean-arch, response, error, success, mapeamento, json]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
Para garantir resposta consistente e rastreável, a aplicação precisa de uma camada responsável por padronizar sucesso, erro e logging. Essa camada não deve depender diretamente do framework web nem do protocolo de transporte, mas precisa ser capaz de transformar resultados de use cases em respostas previsíveis para o cliente.

## Decisão
Será criada uma camada de `handler` na pasta `adapter` com responsabilidade única de transformar resultado de uma operação em uma resposta padronizada.

### Biblioteca recomendada
A implementação inicial utilizará a biblioteca padrão `encoding/json` para serializar as respostas padronizadas, mantendo o fluxo simples e compatível com os frameworks comparados. A escolha foi feita para reduzir dependências extras e preservar a neutralidade da camada de handler.

### Comportamento esperado
- O handler receberá um `context.Context` e um `result any`.
- Se `result` for um `error`, o handler identificará o tipo e/ou mensagem do erro e aplicará a estratégia adequada de mapeamento.
- Se `result` não for um erro, o handler interpretará como sucesso e retornará o payload de resposta já preparado pelo caso de uso ou presenter.
- A implementação concreta do handler deve ser mantida em `adapter/handler` e consumida pela camada de infraestrutura por meio da interface `ports.Handler`.

### Resposta padrão
A resposta padrão de erro deve seguir o seguinte formato:

```json
{
  "erro": {
    "codigo": "codigo de erro",
    "mensagem": "mensagem de erro explicando o por que do erro",
    "detalhes": [
      {
        "campo": "campo relacionado ao erro",
        "causa": "causa do erro",
        "valor": "valor associado"
      }
    ]
  }
}
```

### Estratégia de mapeamento
- O handler usará o padrão Strategy para escolher como mapear um erro para uma resposta.
- A função de mapeamento será definida como um tipo em `responses` com a assinatura `ErrorResponseFuncType`.
- O contrato da função anônima será:

```go
type ErrorResponseFuncType func(ctx context.Context, err error) Response[any, any]
```

- O handler deverá manter dois mecanismos de lookup:
  1. para dar match com o texto do `error`;
  2. para dar match com o tipo do `error`.
- Os mapas devem ser definidos com a seguinte intenção:
  - `errorMessageHandlers map[string]ErrorResponseFuncType`
  - `errorTypeHandlers map[reflect.Type]ErrorResponseFuncType`
- Caso nenhum mapeamento seja encontrado, o handler deverá retornar uma resposta padrão de erro com status `500` e mensagem `Erro interno do servidor não mapeado`.

### Interface do handler
```go
type Handler interface {
    Handle(ctx context.Context, successStatusCode int, output any, headers domain.HttpParamsType, cookies domain.HttpParamsType) responses.Response[any, any]
    RegisterByMessage(message string, fn responses.ErrorResponseFuncType)
    RegisterByType(err error, fn responses.ErrorResponseFuncType)
    ResolveError(ctx context.Context, err error) responses.Response[any, any]
}
```

A interface permanece definida em `core/ports` para manter o contrato do domínio desacoplado do framework e da infraestrutura.

### Logging
- O handler deve registrar eventos de erro e sucesso com contexto de requisição.
- Logs devem incluir identificador de requisição, status, código de erro, mensagem e detalhes relevantes.
- O logging deve ser implementado por meio de abstração de observabilidade em `infra`.
- A implementação concreta da observabilidade fica em `infra/observability` e deve ser injetada quando necessário.

### Integração com o framework
- O handler não deve conhecer diretamente o framework escolhido.
- Ele deverá retornar uma estrutura de resposta pronta para o framework serializar.
- O framework, por sua vez, será responsável por escrever a resposta ao cliente com o status e o body corretos.

## Consequências
- As respostas da API passam a ser uniformes entre os frameworks comparados.
- Erros tornam-se mais fáceis de rastrear e mapear.
- O domínio permanece desacoplado de detalhes de serialização e transporte.

## Referências
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
