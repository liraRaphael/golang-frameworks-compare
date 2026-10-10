# ADR-001: Estrutura da aplicação

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [app, clean-arch, arquitetura, estrutura]

## Pré-requisito
- Ter o [Golang](https://go.dev/dl/) instalado
- Conhecimento básico de Clean Architecture
- Visão geral do projeto em [readme.md](../../readme.md)

## Contexto
O projeto precisa de uma estrutura consistente para permitir a comparação entre frameworks web sem misturar regras de negócio, adaptação de entrada/saída e infraestrutura. A arquitetura precisa facilitar testes, substituição de dependências e observabilidade, além de evitar que a lógica de negócio dependa de detalhes de HTTP, banco de dados ou framework escolhido.

A estrutura descrita neste ADR acompanha o código presente em [api](../../api). A implementação está organizada em `core`, `app` e `infra`; não existe um diretório `api/adapter`.

## Decisão
A aplicação será organizada nos pacotes que existem hoje:

- `api/core`: contratos e lógica do domínio, com `domain`, `ports` e `usecases`.
- `api/app`: controllers, handler, presenters e validação que coordenam os contratos de `core`.
- `api/infra`: configuração, contexto, diagnóstico, documentação Swagger, adapters de frameworks, clientes, observabilidade e roteador.

O ponto de composição é `api/main.go`: carrega a configuração, inicializa observabilidade, monta os componentes manualmente, seleciona o framework e registra as rotas `/health` e `/hello`.

### Regras de dependência
1. Use `core/ports` para contratos consumidos pela aplicação e infraestrutura; a implementação atual não aplica isolamento estrito, pois `ports.Context` importa logger e tracing de `infra`, e `ports.Router` importa o registro Swagger de `infra`.
2. Controllers da aplicação dependem dos contratos de `core` e também de utilitários de `infra` quando necessário.
3. O bootstrap instancia implementações concretas em `main.go`; a injeção é manual e feita por construtores.
4. Operações que recebem contexto usam `ports.Context`, que estende `context.Context`.

### Estrutura esperada
- `api/core/domain/...`
- `api/core/usecases/...`
- `api/core/ports/...`
- `api/app/controller/...`
- `api/app/handler/...`
- `api/app/validation/...`
- `api/app/presenter/...`
- `api/infra/config/...`
- `api/infra/context/...`
- `api/infra/clients/...`
- `api/infra/diagnostic/...`
- `api/infra/docs/swagger/...`
- `api/infra/observability/...`
- `api/infra/router/...`
- `api/infra/frameworks/gin/...`
- `api/infra/frameworks/fiber/...`
- `api/infra/frameworks/echo/...`
- `api/infra/frameworks/chi/...`
- `api/infra/frameworks/beego/...`
- `api/infra/frameworks/nethttp/...`

### Responsabilidades por camada
#### Core
- Definir regras de negócio puras.
- Declarar as portas implementadas para clientes, validadores e outros contratos; atualmente não há contrato de repositório.
- Manter tipos de erro de domínio e contratos de entrada/saída.

#### App
- Implementar controllers, handler, presenters e validação.
- Orquestrar os casos de uso e mapear seus resultados para contratos de resposta.
- Manter a integração HTTP específica nos adapters de `infra/frameworks`.

#### Infra
- Implementar os adapters concretos de framework e clientes externos, além de configuração, tracing, métricas, roteamento e documentação.
- Garantir que a escolha de framework ou biblioteca não vaze para o domínio.
- Manter cada implementação concreta de framework em um subdiretório isolado, por exemplo `api/infra/frameworks/gin` e `api/infra/frameworks/fiber`, para evitar que uma integração contamine a outra.

### Limites conhecidos da implementação
- `api/core/domain/entities`, `api/core/domain/services`, `api/infra/database`, `api/adapter` e `api/adapter/services` não existem atualmente.
- Os contratos de `core/ports` ainda referenciam alguns tipos de `infra`; portanto, o desacoplamento de camadas descrito como objetivo não está integralmente realizado.

### Padrões recomendados
- Interfaces pequenas e específicas.
- Injeção de dependência via construtores.
- Erros e contratos de erro ficam em `core/domain`; `app/handler` resolve erros para responses e adapters de `infra/frameworks` escrevem a resposta HTTP.
- Não colocar código de framework, banco ou SDK no `core`.

## Consequências
- O projeto fica mais previsível para novos contribuidores.
- Mudanças em framework, banco ou cliente externo ficam isoladas em `infra`.
- Casos de uso podem ser testados sem depender do servidor web ou da infraestrutura real.

## ADRs complementares
- [ADR-001a: Detalhamento da camada Core](001a-camada-core.md)
- [ADR-001b: Detalhamento da camada App](001b-camada-adapter.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)
- [ADR-007: Geração automática de documentação OpenAPI/Swagger](007-openapi-swagger.md)
- [ADR-008: Pprof e diagnóstico de performance](008-pprof.md)
- [ADR-009: Padrões de Request, Response e HttpParam no Core](009-request-response-core.md)
- [ADR-010: Injeção de Dependência (DI)](010-injecao-dependencia.md)
- [ADR-011: Gestão de Configuração](011-gestao-configuracao.md)
- [ADR-012: Mecanismo de Validação de Dados](012-validacao-campo.md)

## Referências
1. [Clean Architecture](https://dev.to/yuripeixinho/clean-architecture-arquitetura-limpa-33e1)
2. [Golang download](https://go.dev/dl/)