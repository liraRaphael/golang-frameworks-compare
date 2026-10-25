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
`api/infra/frameworks` contém adapters para Gin, Fiber, Echo, Chi, Beego e também net/http. O bootstrap seleciona um deles conforme `FRAMEWORK`, usando net/http para valor ausente ou não reconhecido.

Os adapters implementam um contrato comum de registro de rota e inicialização, mas não uniformizam todos os comportamentos. As diferenças efetivas estão registradas nesta ADR e devem ser consideradas em qualquer comparação entre frameworks.

## Decisão
Estão implementados adapters concretos para os frameworks listados nesta ADR, cada um encapsulando a integração específica sem alterar a lógica de negócio.

### Escopo da implementação
Os seguintes pacotes existem em [api/infra/frameworks](../../api/infra/frameworks):
- [api/infra/frameworks/gin](../../api/infra/frameworks/gin)
- [api/infra/frameworks/fiber](../../api/infra/frameworks/fiber)
- [api/infra/frameworks/echo](../../api/infra/frameworks/echo)
- [api/infra/frameworks/chi](../../api/infra/frameworks/chi)
- [api/infra/frameworks/beego](../../api/infra/frameworks/beego)
- [api/infra/frameworks/nethttp](../../api/infra/frameworks/nethttp)

Cada adapter implementa `ports.FrameworkAdapter`, declarado em [api/core/ports/presenter.go](../../api/core/ports/presenter.go), com estas operações:
- registro de rotas por método HTTP e path;
- encaminhamento de middlewares (a aplicação dos middlewares ainda é no-op ou middleware vazio nos adapters atuais);
- inicialização do servidor em um endereço configurável;
- adaptação do contexto nativo do framework para `ports.Context`;
- parsing/binding do request e validação do content type JSON quando há body;
- serialização e escrita HTTP da response tipada produzida pelo controller.

### Padrão de integração
As implementações seguem um padrão semelhante ao adapter net/http:
- o núcleo da aplicação permanece inalterado;
- os controllers e use cases continuam sendo consumidos via interfaces de porta;
- o framework concreto atua apenas como adaptador de entrada/saída;
- a rota tipada compartilhada valida o DTO convertido antes de invocar o controller;
- o controller define o status e envolve o resultado do caso de uso em `responses.Response`; o adapter serializa e escreve essa response;
- `StandardRouter` registra rotas e encaminha middlewares a cada adapter;
- o request wrapper fornece body e mapas HTTP a controllers genéricos. Em `nethttp`, os parâmetros de path não são extraídos.

### Regras de implementação
1. Cada framework terá um pacote próprio isolado, com um arquivo principal de adaptação, por exemplo `adapter.go`.
2. A implementação deve respeitar o contrato de `ports.FrameworkAdapter` e não deve introduzir lógica de negócio.
3. O mesmo conjunto de rotas, handlers e middlewares deve poder ser usado independentemente do framework escolhido.
4. O bootstrap da aplicação deve permitir trocar de framework sem reescrever os controllers, use cases ou presenters.
5. `api/main.go` seleciona o adapter por `AppConfig.Framework`.

### Diferenças conhecidas
- O suporte a métodos HTTP varia: Beego tem casos explícitos para GET/POST/PUT/DELETE e usa `Any` nos demais; os outros registram o método fornecido ao framework, enquanto net/http valida o método dentro do handler.
- `Use` não executa o middleware fornecido nos adapters existentes.
- Os adapters diferem em extração de path params, binding, cookies e tratamento de falhas de escrita; portanto, o contrato comum não significa comportamento completamente equivalente.
- Em rotas tipadas, o body é convertido e validado antes da chamada ao controller. O controller cria a response tipada e define o status, mas não serializa JSON nem escreve HTTP.

### Exemplo de referência baseado em net/http
O adapter padrão do Go em [api/infra/frameworks/nethttp/adapter.go](../../api/infra/frameworks/nethttp/adapter.go) continua sendo uma opção selecionável em runtime e o fallback do bootstrap.

### Ponto importante
- net/http é uma implementação mantida, não apenas uma referência;
- as cinco integrações de terceiros também estão presentes;
- os adapters compartilham a interface, mas há diferenças comportamentais concretas, listadas acima.

### Dependências recomendadas
As dependências atualmente usadas incluem:
- Gin: `github.com/gin-gonic/gin`
- Fiber: `github.com/gofiber/fiber/v2`
- Echo: `github.com/labstack/echo/v4`
- Chi: `github.com/go-chi/chi/v5`
- Beego: `github.com/beego/beego/v2/server/web`

As versões estão fixadas no `api/go.mod`; existem testes para alguns adapters, mas o escopo não garante equivalência comportamental completa.

## Consequências
- A aplicação contém adapters para os frameworks listados nesta ADR e para net/http.
- A interface comum fornece uma base para comparação; as diferenças conhecidas devem ser consideradas ao interpretar resultados de benchmark.
- A camada de infraestrutura fica mais modular e facilitada para futuras extensões.
- A troca de framework torna-se um esforço localizado, reduzindo o risco de regressão.

## Referências
1. [readme.md](../../readme.md)
2. [api/infra/frameworks/nethttp/adapter.go](../../api/infra/frameworks/nethttp/adapter.go)
3. [api/core/ports/presenter.go](../../api/core/ports/presenter.go)
4. [docs/adr/004-roteamento.md](004-roteamento.md)
