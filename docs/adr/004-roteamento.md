# ADR-004: Roteamento

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [roteamento, router, response, status code, swagger, openapi, pprof]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-003: Handler](003-handler.md)

## Contexto
Como o projeto compara vários frameworks web em Go, o roteamento usa um contrato abstrato para registrar as mesmas rotas em adapters distintos. A implementação atual ainda apresenta diferenças de comportamento entre frameworks.

## Decisão
O roteador abstrato existente está em `api/infra/router`. `StandardRouter` registra rotas no adapter selecionado e, quando configurado, adiciona metadados ao registry Swagger.

### Objetivos
- Permitir que a mesma definição de rota funcione em Gin, Echo, Fiber, Chi ou outro framework.
- Centralizar registro de rotas e encaminhamento de middlewares ao adapter. A aplicação dos middlewares ainda não está implementada nos adapters de framework existentes.
- Habilitar documentação automática via OpenAPI/Swagger.
- Iniciar o servidor `pprof` por meio do pacote `infra/diagnostic`, em separado do servidor de rotas.

### Componente de roteamento
- O contrato `ports.Router` declara `WithSwagger`, `AddRoute`, `Use` e `Start`; o contrato `ports.FrameworkAdapter` declara `RegisterRoute`, `Use` e `Start`.
- `router.NewStandardRouter` recebe um `FrameworkAdapter`. Não há `RouterBuilder` nem função `RegisterRoutes` implementados.
- `AddRoute` guarda método, path, controller e `swagger.RouteMetadata`; `router.AddTypedRoute` adapta controllers tipados à interface de rota genérica.
- Se a inclusão da rota no registry Swagger falhar, `StandardRouter.AddRoute` atualmente descarta o erro.
- O status é definido na resposta do controller/handler; o router não aplica default por método HTTP.

### Contratos de interface
#### Infra / Frameworks
`FrameworkAdapter` atua como ponte entre o framework concreto e as portas da aplicação, adaptando requests/responses e iniciando o servidor.

Os contratos estão em `api/core/ports/presenter.go`:

```go
type FrameworkAdapter interface {
    RegisterRoute(method, path string, ctrl Controller[any, any, any, any])
    Use(middleware Middleware)
    Start(addr string) error
}
```

### Wrapper de framework
- A implementação concreta de `FrameworkAdapter` traduz o ciclo de vida do framework para o contrato padrão do `Router`.
- Os adapters concretos convertem contextos, corpo e parâmetros, chamam o controller e escrevem a resposta JSON. Atualmente `Use` ignora o middleware recebido em Gin, Chi, Echo, Beego e net/http; Fiber instala middleware vazio. Portanto, não se deve considerar middlewares funcionais.
- Cada framework terá um subdiretório próprio em `infra/frameworks/<framework>` para isolar a integração e preservar a compatibilidade com as diferenças de implementação entre Gin, Fiber, Echo e outros.

### Fluxo de construção
- `main.go` escolhe entre Gin, Fiber, Echo, Chi, Beego e net/http conforme `FRAMEWORK` e constrói `StandardRouter`. Valor não reconhecido usa net/http.
- Em testes de integração, um `MockAdapter` pode ser injetado para validar registro de rotas, middlewares e inicialização sem subir um servidor HTTP real.

### Documentação
- Quando o registry é configurado, `AddRoute` registra summary, description, tags, tipos de entrada/saída e parâmetros a partir de `RouteMetadata`.
- O pacote `infra/docs/swagger` usa `swaggest/openapi-go`, publica `/openapi.json` e serve Swagger UI em `/docs/`, num servidor separado. A geração define a resposta de saída como HTTP 200.

### Observabilidade e diagnóstico
- `main.go` inicia `pprof` em `APP_PPROF_PORT` (padrão `8081`) e Swagger UI em `APP_SWAGGER_PORT` (padrão `8082`), em goroutines separadas.

## Consequências
- O mesmo conjunto de rotas pode ser usado para comparar o comportamento dos frameworks sem reescrever o contrato da aplicação.
- A definição de registro de rotas é compartilhada; o comportamento dos adapters ainda apresenta diferenças, descritas na ADR-013.
- A infraestrutura de diagnóstico fica isolada do fluxo principal da API.

## Referências
1. https://pkg.go.dev/github.com/swaggest/openapi-go#section-readme
