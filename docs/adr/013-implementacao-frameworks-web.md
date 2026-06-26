# ADR-013: Implementação dos frameworks web comparados

**Status**: Aceito

**Data**: 2026-06-26

**Tags**: [frameworks, web, infra, benchmark, gin, fiber, echo, chi, beego]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)
- [ADR-004: Roteamento](004-roteamento.md)
- [ADR-010: Injeção de Dependência](010-injecao-dependencia.md)

## Contexto
O projeto já possui uma implementação inicial de adaptador para o pacote padrão do Go em [api/infra/frameworks/nethttp/adapter.go](../../api/infra/frameworks/nethttp/adapter.go), mas o [readme.md](../../readme.md) define uma lista explícita de frameworks a serem avaliados: Gin, Fiber, Echo, Chi e Beego.

Para que a comparação entre os frameworks seja justa e reproduzível, a aplicação precisa suportar cada um deles por meio de adaptadores concretos que exponham o mesmo contrato de roteamento, middleware e inicialização. Sem essa padronização, cada implementação acabaria se desviando em detalhes de integração, o que comprometeria a comparabilidade dos benchmarks e dificultaria a manutenção do código.

## Decisão
Serão implementados adaptadores concretos para os frameworks listados no README, cada um encapsulando a integração específica com o framework sem alterar a lógica de negócio da aplicação.

### Escopo da implementação
Os seguintes pacotes serão adicionados em [api/infra/frameworks](../../api/infra/frameworks):
- [api/infra/frameworks/gin](../../api/infra/frameworks/gin)
- [api/infra/frameworks/fiber](../../api/infra/frameworks/fiber)
- [api/infra/frameworks/echo](../../api/infra/frameworks/echo)
- [api/infra/frameworks/chi](../../api/infra/frameworks/chi)
- [api/infra/frameworks/beego](../../api/infra/frameworks/beego)

Cada adaptador deverá implementar a interface de contrato já definida em [api/core/ports/handler.go](../../api/core/ports/handler.go), com suporte mínimo para:
- registro de rotas por método HTTP e path;
- aplicação de middlewares;
- inicialização do servidor em um endereço configurável;
- propagação de contexto de requisição;
- serialização consistente de resposta e status code.

### Padrão de integração
A implementação seguirá o mesmo padrão já adotado pelo adaptador de net/http:
- o núcleo da aplicação permanece inalterado;
- os controllers e use cases continuam sendo consumidos via interfaces de porta;
- o framework concreto atua apenas como adaptador de entrada/saída;
- o roteador padrão continua sendo responsável por registrar as rotas e os middlewares de forma uniforme.

### Regras de implementação
1. Cada framework terá um pacote próprio isolado, com um arquivo principal de adaptação, por exemplo `adapter.go`.
2. A implementação deve respeitar o contrato de `ports.FrameworkAdapter` e não deve introduzir lógica de negócio.
3. O mesmo conjunto de rotas, handlers e middlewares deve poder ser usado independentemente do framework escolhido.
4. O bootstrap da aplicação deve permitir trocar de framework sem reescrever os controllers, use cases ou presenters.
5. A escolha do framework deve ser controlada por configuração ou por uma pequena camada de composição no ponto de entrada da aplicação.

### Exemplo de referência baseado em net/http
Embora o adaptador padrão do Go em [api/infra/frameworks/nethttp/adapter.go](../../api/infra/frameworks/nethttp/adapter.go) sirva como referência de composição e fluxo, ele não será mantido como implementação final do produto. O objetivo deste exemplo é ilustrar o padrão de adaptação a ser seguido pelos frameworks concretos.

Exemplo conceitual de estrutura de um adaptador:

```go
type adapter struct {
    mux    *http.ServeMux
    handle ports.Handler
}

func NewAdapter(handle ports.Handler) ports.FrameworkAdapter {
    return &adapter{mux: http.NewServeMux(), handle: handle}
}

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller) {
    a.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        result, err := ctrl.WrapperExecute(ctx, requests.NewRequest(...)) // extrair os parâmetros
        if err != nil {
            a.handle.ResolveError(ctx, err)
            return
        }

        w.WriteHeader(http.StatusOK)
        _ = json.NewEncoder(w).Encode(result)
    })
}

func (a *adapter) Use(middleware ports.Middleware) {
    ...
}

func (a *adapter) Start(addr string) error {
    return http.ListenAndServe(addr, a.mux)
}
```

### Ponto importante
- esse exemplo é apenas uma referência de fluxo de integração;
- a implementação final deverá ser substituída por adaptadores específicos para Gin, Fiber, Echo, Chi e Beego;
- o contrato da aplicação continua sendo o mesmo, independentemente do framework escolhido.

### Dependências recomendadas
As integrações devem usar as bibliotecas oficiais ou mais consolidadas do ecossistema Go para cada framework, por exemplo:
- Gin: `github.com/gin-gonic/gin`
- Fiber: `github.com/gofiber/fiber/v2`
- Echo: `github.com/labstack/echo/v4`
- Chi: `github.com/go-chi/chi/v5`
- Beego: `github.com/beego/beego/v2/server/web`

As versões devem ser fixadas no `go.mod` e acompanhadas de testes de integração básicos para garantir compatibilidade com o contrato atual.

## Consequências
- A aplicação passa a suportar os frameworks listados no README de forma consistente e comparável.
- Os benchmarks deixam de depender de implementações ad hoc e passam a refletir o comportamento real de cada framework.
- A camada de infraestrutura fica mais modular e facilitada para futuras extensões.
- A troca de framework torna-se um esforço localizado, reduzindo o risco de regressão.

## Referências
1. [readme.md](../../readme.md)
2. [api/infra/frameworks/nethttp/adapter.go](../../api/infra/frameworks/nethttp/adapter.go)
3. [api/core/ports/handler.go](../../api/core/ports/handler.go)
4. [docs/adr/004-roteamento.md](004-roteamento.md)
