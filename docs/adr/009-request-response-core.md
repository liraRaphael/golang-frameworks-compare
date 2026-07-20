# ADR-009: Padrões de Request, Response e HttpParam no Core

**Status**: Aceito

**Data**: 2026-07-19

**Tags**: [core, request, response, httpparam, domain, interfaces, generics]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001a: Detalhamento da camada Core](001a-camada-core.md)

## Contexto
Para padronizar e flexibilizar a entrada e saída de dados entre casos de uso, controllers e adapters de frameworks web de forma agnóstica e tipada, foram introduzidos contratos genéricos com suporte a múltiplos parâmetros HTTP por meio de reflexão (Struct Tags) e tipos auxiliares de alta performance.

## Decisão
O `core` define as seguintes interfaces genéricas e structs imutáveis para transporte de dados:

### Request padrão (Interface Genérica)
O contrato de request padrão possui dois parâmetros de tipo: `BodyType` e `ParamType`. Ele extrai cabeçalhos, query strings, cookies e parâmetros de path dinamicamente por reflexão baseando-se em tags de struct (`header`, `query`, `path`, `cookie`):

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
O contrato de response também é genérico e permite expor o corpo, parâmetros, cabeçalhos, cookies e status code de forma tipada e consistente:

```go
type Response[BodyType any, ParamType any] interface {
    Body() BodyType
    Params() ParamType
    Headers() domain.HttpParamsType
    Cookies() domain.HttpParamsType
    StatusCode() int
}
```

### HttpParamsType e HttpParam Type
Parâmetros HTTP de entrada e saída são mapeados por `HttpParamsType`, que é um mapa de listas de strings, facilitando a manipulação segura de chaves com múltiplos valores:

```go
type HttpParamsType map[string]HttpParamType

type HttpParamType []string
```

O tipo `HttpParamType` expõe métodos utilitários robustos para conversão e fallback seguro (ex: `GetFirstIntOrDefault`, `GetFirstBoolOrDefault`).

### Instanciação de Requests e Responses
Os construtores no Core geram automaticamente as estruturas concretas (`Requester` e `Responser`) e realizam a injeção/extração baseada em tags:
- `requests.NewRequest[BodyType, ParamType](body, params)`
- `requests.NewRequestFromParams[BodyType, ParamType](body, headers, query, path, cookies)`
- `responses.NewResponse[BodyType, ParamType](body, params, statusCode)`
- `responses.NewResponseFromParams[BodyType, ParamType](body, statusCode, headers, cookies)`

Se `ParamType` for fornecido como um ponteiro de struct, os valores de cabeçalhos, query e path params com as struct tags correspondentes serão populados de forma automática no campo `params`. Caso o `ParamType` seja especificado como `any`, o parser ignora a reflexão e assume um valor nulo/seguro sem causar panics.

### Regras
- O `Request` e o `Response` permanecem em `core/domain` e não dependem de bibliotecas de terceiros ou frameworks.
- Adapters de infraestrutura devem traduzir as requisições HTTP em `Request[any, any]` usando as funções construtoras e o wrapper de requisição para compatibilidade genérica nos controllers.

## Consequências
- Total tipagem e segurança de dados de ponta a ponta sem vazamento de detalhes de framework.
- Reflexão robusta e encapsulada para extração automática de dados HTTP via struct tags.
- Desempenho preservado e prevenção de panics de ponteiro nulo em requisições genéricas.
