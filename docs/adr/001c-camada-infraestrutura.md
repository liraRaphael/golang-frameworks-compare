# ADR-001c: Detalhamento da camada Infrastructure

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [infrastructure, infra, config, database, observability, routers]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
A camada `infra` concentra as implementações concretas que permitem que a aplicação funcione no mundo real: banco de dados, clientes externos, configuração, observabilidade e frameworks.

## Decisão
A camada `infra` será organizada em subpacotes com responsabilidade específica:

- `infra/config`: carregamento de variáveis de ambiente e configuração da aplicação.
- `infra/database`: conexões, migrations e repositórios concretos.
- `infra/clients`: clientes HTTP/gRPC para serviços externos.
- `infra/observability`: logs, métricas e tracing.
- `infra/router`: implementação concreta do roteamento para o framework escolhido.
- `infra/frameworks`: adaptações específicas para Gin, Echo, Fiber, Chi e outros.

### Regras
- `infra` deve implementar as interfaces definidas em `core`.
- Detalhes de framework, SDK e banco não devem aparecer no domínio.
- A configuração deve ser centralizada e inicializada de forma previsível.

## Consequências
- A aplicação pode trocar biblioteca ou framework sem reescrever o domínio.
- A infraestrutura fica isolada e mais fácil de manter.
- A comparação entre frameworks passa a ser mais objetiva.
