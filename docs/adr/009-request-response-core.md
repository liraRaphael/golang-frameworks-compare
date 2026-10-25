# ADR-009: Padrões de Request, Response e HttpParam no Core

**Status**: Aceito

**Data**: 2026-07-19

**Tags**: [core, request, response, httpparam, domain, interfaces, generics]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001a: Detalhamento da camada Core](001a-camada-core.md)

## Contexto
`core/domain/requests` e `core/domain/responses` definem interfaces e structs genéricas para transporte de dados. `infra/router/typed.go` fornece um adapter que converte controllers genéricos para controllers tipados.

## Decisão
O `core` define as seguintes interfaces genéricas e structs para transporte de dados:

### Request padrão (Interface Genérica)
O contrato de request possui `BodyType` e `ParamType`. `Requester` expõe body, params e os mapas HTTP. Os nomes de tags aceitos são `header`, `query`, `path` e `cookie`; a extração de índices usa reflexão e cacheia os índices por tipo:

```go
type Request[BodyType any, ParamType any] interface {
    Body() BodyType
    Params() ParamType
    Headers() domain.HttpParamsType
    QueryParams() domain.HttpParamsType
    PathParams() domain.HttpParamsType
    Cookies() domain.HttpParamsType
}
```

### Response padrão (Interface Genérica)
O contrato de response possui body, params, headers, cookies e status code:

```go
type Response[BodyType any, ParamType any] interface {
    Body() BodyType
    Params() ParamType
    Headers() domain.HttpParamsType
    Cookies() domain.HttpParamsType
    StatusCode() int
}
```

### Responsabilidade de controller e use case
O controller é responsável por definir o status HTTP e criar a response tipada `responses.Response[ResponseBody, ParamType]`. O use case retorna apenas o corpo tipado (`ResponseBody`) e `error`; não constrói `responses.Response`, não define status HTTP e não serializa JSON. O controller não serializa nem escreve HTTP.

O adapter de framework recebe a response criada pelo controller, preserva status e headers e realiza a serialização/escrita HTTP. Para erros de parsing, validação ou execução, o fluxo comum resolve o erro e o adapter serializa/escreve a response de erro.

### HttpParamsType e HttpParam Type
Parâmetros HTTP de entrada e saída são mapeados por `HttpParamsType`, que é um mapa de listas de strings, facilitando a manipulação segura de chaves com múltiplos valores:

```go
type HttpParamsType map[string]HttpParamType

type HttpParamType []string
```

`HttpParamType` expõe helpers como `GetFirstInt` e `GetFirstBoolOrDefault`. O comportamento de fallback é implementado pelo respectivo helper; não deve ser tratado como validação dos dados.

### Instanciação de Requests e Responses
Os construtores geram `Requester`/`Responser` e, quando fornecido um `ParamType` struct compatível, extraem ou populam campos a partir das tags:
- `requests.NewRequest[BodyType, ParamType](body, params)`
- `requests.NewRequestFromParams[BodyType, ParamType](body, headers, query, path, cookies)`
- `responses.NewResponse[BodyType, ParamType](body, params, statusCode)`
- `responses.NewResponseFromParams[BodyType, ParamType](body, statusCode, headers, cookies)`

`PopulateParams` aceita struct ou ponteiro para struct e converte o primeiro valor encontrado por tag para tipos string, int, uint e bool; tipos incompatíveis retornam erro. `NewRequestFromParams` e `NewResponseFromParams` transformam esse erro em panic. `ExtractParamsFromTag` ignora tipos que não sejam struct e mantém cache de índices por tipo.

### Integração atual com rotas
- Os adapters de framework constroem requests como `Request[any, any]`, preenchendo os mapas HTTP em wrappers concretos.
- `AddTypedRoute` converte body e `Params()` via JSON marshal/unmarshal antes de chamar o controller tipado. Ele delega os acessores Headers/QueryParams/PathParams/Cookies ao request original.
- O wrapper tipado não repopula `InputParams` a partir dos mapas HTTP; portanto, a extração tipada dos parâmetros não está garantida no fluxo das rotas atuais.
- `Responser.MarshalJSON` emite `Body`, `StatusCode` e `Headers` (com esses nomes de chave), não inclui cookies nem `Params`. Os adapters também usam o status e headers da response HTTP diretamente.

### Regras
- O `Request` e o `Response` permanecem em `core/domain` e não dependem de bibliotecas de terceiros ou frameworks.
- Adapters de infraestrutura traduzem requisições HTTP para `Request[any, any]`; `AddTypedRoute` é a ponte disponível para controllers tipados.
- Controllers definem status e envelope tipado; use cases retornam somente o corpo e erro; adapters são responsáveis por serializar e escrever a response.

## Consequências
- Contratos genéricos estão disponíveis; as rotas existentes usam interfaces genéricas e a conversão tipada inclui serialização JSON intermediária.
- Reflexão robusta e encapsulada para extração automática de dados HTTP via struct tags.
- Os índices das tags são cacheados por tipo; `NewRequestFromParams`/`NewResponseFromParams` ainda podem gerar panic se a conversão de um parâmetro tagueado falhar.
