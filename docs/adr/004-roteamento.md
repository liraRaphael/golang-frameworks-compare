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
- O componente será implementado por meio de um `Router` abstrato, de um `RouterBuilder` e de um conjunto de `FrameworkAdapter`.
- Cada framework terá seu próprio inicializador em `infra/frameworks/<nome>`.
- Uma função de registro de rotas (ex: `RegisterRoutes(router any, ...handlers)`) será responsável por definir os endpoints de forma agnóstica. O `adapter` de cada framework fará a conversão do `router any` para o tipo específico (`*gin.Engine`, `*fiber.App`).
- Cada rota deve aceitar método HTTP, path, handler, middlewares e metadata de documentação.
- O padrão deve prever status code padrão por método, com `POST` usando `201 Created` e demais métodos usando `200 OK` por padrão.

### Contratos de interface
#### Infra / Frameworks
A interface que cada `adapter` de framework deve implementar ficará em `infra/frameworks` e atuará como a ponte entre o mundo do framework e o contrato padrão do projeto. O objetivo é adaptar uma requisição do framework para o `Request` do domínio e converter o `Response` do domínio para a resposta do framework.

```go
// api/infra/frameworks/adapter.go
package frameworks

import "seu-projeto/api/adapter/router"

// FrameworkAdapter é um exemplo de como um handler de framework (ex: gin.HandlerFunc)
// pode ser criado para envolver um controller do domínio.
type FrameworkAdapter func(controller YourControllerInterface) FrameworkHandler
    Start(addr string) error
}
```

### Wrapper de framework
- A implementação concreta de `FrameworkAdapter` deve ser extremamente leve e atuar apenas como um tradutor entre o ciclo de vida do framework e o contrato padrão do `Router`.
- Ela deve lidar com a conversão do contexto nativo do framework para `ports.Context`, além de payloads de request/response e eventos de execução do framework, sem conter lógica de negócio, regras de domínio ou persistência.
- Cada framework terá um subdiretório próprio em `infra/frameworks/<framework>` para isolar a integração e preservar a compatibilidade com as diferenças de implementação entre Gin, Fiber, Echo e outros.

### Fluxo de construção
- O `RouterBuilder` receberá um `FrameworkAdapter` e montará um `standardRouter` com comportamento uniforme entre os frameworks comparados.
- O `main.go` ou o bootstrap da aplicação passará apenas a escolha do framework e a construção do roteador, sem necessidade de reescrever a lógica de endpoints.
- Em testes de integração, um `MockAdapter` pode ser injetado para validar registro de rotas, middlewares e inicialização sem subir um servidor HTTP real.

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
