# ADR-009: Padrões de Request, Response e HttpParam no Core

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [core, request, response, httpparam, domain, structs]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001a: Detalhamento da camada Core](001a-camada-core.md)

## Contexto
Para padronizar a entrada e saída entre os casos de uso e os adapters, o projeto precisa definir estruturas reutilizáveis de request e response no `core`, além de uma estrutura auxiliar para manipular parâmetros HTTP de forma segura.

## Decisão
O `core` deverá definir os seguintes tipos padrão:

### Request padrão
A estrutura base de request deverá ser:

```go
type Request[BodyType any] struct {
    Body        BodyType
    Headers     map[string][]string
    QueryParams HttpParam
    PathParams  map[string]string
}
```

### Response padrão
A estrutura base de response deverá ser:

```go
type Response struct {
    Body       any
    Headers    HttpParam
    StatusCode int
}
```

### HttpParam
`HttpParam` será um tipo auxiliar para armazenar parâmetros HTTP de forma consistente e segura.

#### Interface esperada
```go
type HttpParam interface {
    Get(key string) ([]string, bool)
    GetFirst(key string) (string, bool)
    GetFirstOrDefault(key string, defaultValue string) string
    GetAll() map[string][]string
    Set(key string, value string)
    SetAll(params map[string][]string)
    Delete(key string)
    GetFirstInt(key string) (int, bool)
    GetFirstIntOrDefault(key string, defaultValue int) int
    Clear()
}
```

### Regras
- O `Request` e o `Response` devem permanecer no `core` e não depender de bibliotecas de framework.
- Estruturas específicas de cada caso de uso devem ficar em subpacotes, por exemplo `domain/requests/user`.
- O `adapter` é responsável por transformar dados da requisição HTTP em `Request` e por converter `Response` para o formato esperado pelo framework.

## Consequências
- A camada de domínio ganha um contrato comum para entrada e saída.
- O código fica mais consistente entre casos de uso e adapters.
- A manipulação de headers, query params e path params torna-se previsível.
