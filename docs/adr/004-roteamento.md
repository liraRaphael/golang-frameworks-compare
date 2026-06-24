# ADR-004: Roteamento

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [roteamento, router, response, status code, swagger, openapi, pprof]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-003: Handler](003-handler.md)

## Contexto
Como o projeto compara vários frameworks web em Go, o roteamento precisa ser abstraído para que as rotas e os comportamentos sejam consistentes independentemente do framework escolhido. Além disso, é importante garantir documentação automática, observabilidade e isolamento de infraestrutura.

## Decisão
Será criado um componente de roteamento abstrato em `adapter` que permita registrar rotas, middlewares e handlers de forma uniforme, independentemente do framework concreto usado em `infra`.

### Objetivos
- Permitir que a mesma definição de rota funcione em Gin, Echo, Fiber, Chi ou outro framework.
- Centralizar regras de status code, middlewares e mapeamento de resposta.
- Habilitar documentação automática via OpenAPI/Swagger.
- Expor endpoints de diagnóstico como `pprof` em um servidor separado.

### Componente de roteamento
- O componente será implementado por meio de um `Router` genérico e de um `RouterBuilder`.
- O roteamento deve usar composição e o padrão Factory para abstrair a implementação concreta do framework.
- A ideia é que o `RouterBuilder` monte um `Router` com comportamento comum, enquanto cada framework em `infra/frameworks` fornece a adaptação específica para registrar rotas e middlewares.
- Cada rota deve aceitar método HTTP, path, handler, middlewares e metadata de documentação.
- O padrão deve prever status code padrão por método, com `POST` usando `201 Created` e demais métodos usando `200 OK` por padrão.

### Interface genérica para frameworks
A interface genérica deverá ficar em `infra/frameworks` e deverá obedecer ao fluxo do handler, garantindo que o framework receba uma função de execução compatível com a resposta padronizada.

```go
type FrameworkAdapter interface {
    RegisterRoute(method string, path string, handler func(ctx context.Context, req any) (any, error)) error
    Use(middleware func(ctx context.Context, req any) (any, error))
    Start(addr string) error
}

type Router interface {
    AddRoute(method string, path string, handler func(ctx context.Context, req any) (any, error)) error
    AddMiddleware(middleware func(ctx context.Context, req any) (any, error))
}

type RouterBuilder interface {
    SetFramework(adapter FrameworkAdapter) RouterBuilder
    Build() Router
}
```

### Integração com frameworks
- Cada framework terá uma implementação concreta na camada `infra`.
- A interface de adaptação do framework deve expor métodos para registrar rota, iniciar o servidor e configurar middlewares.
- O Builder deverá receber um adaptador concreto e construir o `Router` com comportamento uniforme entre os frameworks comparados.

### Documentação
- A documentação OpenAPI/Swagger deverá ser gerada automaticamente a partir das rotas cadastradas.
- A configuração de documentação ficará em `infra/docs/swagger`.
- A geração deve aproveitar os metadados de rotas e os tipos de request/response.

### Observabilidade e diagnóstico
- O `pprof` deverá ficar em um servidor separado, por padrão na porta `8081`, configurável via `infra/config`.
- O Swagger/OpenAPI poderá ficar em um servidor separado, por padrão na porta `8082`, também configurável.

## Consequências
- O mesmo conjunto de rotas pode ser usado para comparar o comportamento dos frameworks sem reescrever o contrato da aplicação.
- A documentação fica mais fácil de manter e sempre sincronizada com as rotas implementadas.
- A infraestrutura de diagnóstico fica isolada do fluxo principal da API.

## Referências
1. https://pkg.go.dev/github.com/swaggest/openapi-go#section-readme
