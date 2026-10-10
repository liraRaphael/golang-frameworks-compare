# ADR-001c: Detalhamento da camada Infrastructure

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [infrastructure, infra, config, observability, routers]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
A camada `infra` concentra as implementações concretas existentes: clientes externos, configuração, contexto, diagnóstico, documentação, observabilidade, roteamento e adapters de frameworks. Não há implementação de banco de dados no pacote `api`.

## Decisão
A camada `infra` será organizada em subpacotes com responsabilidade específica:

- `infra/config`: carregamento de variáveis de ambiente e configuração da aplicação.
- `infra/clients/nethttp` e `infra/clients/fasthttp`: clientes HTTP.
- `infra/clients/grpc`: cliente gRPC cuja operação genérica `Do` retorna erro 501, pois requer invocação específica de serviço.
- `infra/context`: wrapper de `context.Context` com logger e tracing.
- `infra/diagnostic`: inicia o servidor `pprof`.
- `infra/docs/swagger`: registra rotas OpenAPI e sobe Swagger UI.
- `infra/observability`: logs, métricas e tracing.
- `infra/router`: implementação concreta do roteamento para o framework escolhido.
- `infra/frameworks`: adaptações específicas para Beego, Chi, Echo, Fiber, Gin e net/http, além de utilitários HTTP compartilhados.

### Regras
- Adapters e clientes implementam contratos definidos em `core/ports` quando disponíveis.
- Detalhes de framework, SDK e banco não devem aparecer no domínio.
- `AppConfig` concentra as configurações da aplicação e é carregada por `envconfig` durante `init`.

## Consequências
- A aplicação pode trocar biblioteca ou framework sem reescrever o domínio.
- A infraestrutura fica isolada e mais fácil de manter.
- A comparação entre frameworks passa a ser mais objetiva.
