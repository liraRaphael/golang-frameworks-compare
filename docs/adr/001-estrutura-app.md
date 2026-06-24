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

A estrutura proposta aqui foi alinhada ao layout atual do repositório, em que o código da API fica sob [api](../../api) e é organizado em `adapter`, `core` e `infra`.

## Decisão
A aplicação será organizada em três camadas principais:

- `api/core`: núcleo do domínio. Contém entidades, use cases, enums, errors, contratos de request/response e portas (interfaces).
- `api/adapter`: adaptadores de entrada/saída. Contém handlers, controllers, presenters, validação e tradução entre o mundo externo e o domínio.
- `api/infra`: implementações concretas. Contém clientes HTTP/gRPC, repositórios, configuração, observabilidade, roteamento e integração com frameworks.

### Regras de dependência
1. `core` não pode depender de `adapter` nem de `infra`.
2. `adapter` pode depender de `core` e, quando necessário, de `infra`.
3. `infra` pode depender de `core` e, em alguns casos, de `adapter` para composição de infraestrutura específica.
4. Toda comunicação entre camadas deve ocorrer por interfaces definidas em `core`.
5. Funções que realizam I/O, uso de caso de negócio, chamadas externas ou processamento de requisições devem receber `context.Context` como primeiro parâmetro.

### Estrutura esperada
- `api/core/domain/...`
- `api/core/usecases/...`
- `api/core/ports/...`
- `api/adapter/controller/...`
- `api/adapter/services/...`
- `api/adapter/validation/...`
- `api/adapter/presenter/...`
- `api/infra/config/...`
- `api/infra/clients/...`
- `api/infra/observability/...`

### Responsabilidades por camada
#### Core
- Definir regras de negócio puras.
- Declarar interfaces e contratos para repositórios, clientes, validadores e demais dependências.
- Manter tipos de erro de domínio e contratos de entrada/saída.

#### Adapter
- Traduzir requisições externas para objetos de domínio.
- Aplicar validação, mapeamento de resposta e tratamento de erro.
- Encapsular o uso de frameworks web e demais mecanismos de transporte.
- Organizar controllers e services para separar entrada HTTP de implementação de integração.

#### Infra
- Implementar as interfaces definidas em `core`.
- Lidar com banco de dados, configuração, clientes externos, tracing, métricas, routers e documentação.
- Garantir que a escolha de framework ou biblioteca não vaze para o domínio.

### Padrões recomendados
- Interfaces pequenas e específicas.
- Injeção de dependência via construtores.
- Erros de domínio permanecem em `core`; mapas para HTTP/gRPC acontecem em `adapter` ou `infra`.
- Não colocar código de framework, banco ou SDK no `core`.

## Consequências
- O projeto fica mais previsível para novos contribuidores.
- Mudanças em framework, banco ou cliente externo ficam isoladas em `infra`.
- Casos de uso podem ser testados sem depender do servidor web ou da infraestrutura real.

## ADRs complementares
- [ADR-001a: Detalhamento da camada Core](001a-camada-core.md)
- [ADR-001b: Detalhamento da camada Adapter](001b-camada-adapter.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)
- [ADR-007: Geração automática de documentação OpenAPI/Swagger](007-openapi-swagger.md)
- [ADR-008: Pprof e diagnóstico de performance](008-pprof.md)
- [ADR-009: Padrões de Request, Response e HttpParam no Core](009-request-response-core.md)

## Referências
1. [Clean Architecture](https://dev.to/yuripeixinho/clean-architecture-arquitetura-limpa-33e1)
2. [Golang download](https://go.dev/dl/)