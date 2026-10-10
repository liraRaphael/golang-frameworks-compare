# ADR-005: Observability

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [logs, otel, observability, metrics, tracing]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](./001-estrutura-app.md)

## Contexto
`api/main.go` inicializa logger, tracing e métricas antes de montar o router. A infraestrutura exporta traces e métricas por OTLP HTTP; não há instrumentação de métricas de requisição, CPU ou memória implementada nos adapters atuais.

## Decisão
Será criada uma camada de observabilidade em `infra` para coleta de logs, métricas e traces, com foco em comparação entre frameworks.

### Logs
- Uso de `go.uber.org/zap` para geração de logs estruturados.
- O logger concreto usa `zap.NewProduction` e oferece níveis Info, Error, Fatal, Warn e Debug.
- Os adapters registram method, path e trace ID em `incoming request`. Não há request ID ou duração registrados por padrão.
- `ports.Context` oferece um logger e acesso ao trace ID via `infra/context`.

### Tracing
- Uso de `github.com/open-telemetry/opentelemetry-go`.
- Os adapters HTTP criam um span `http.request`; o controller de Hello cria `hello-controller.execute`.
- O tracer configura OTLP HTTP, propagação TraceContext/Baggage e sampler `AlwaysSample`. Chamadas externas e persistência não estão instrumentadas no fluxo presente.

### Métricas
- `infra/observability/metrics` configura um exporter OTLP HTTP e um MeterProvider com exportação periódica.
- Atualmente não há instrumentos nem middleware de métricas definidos; não há endpoint Prometheus/Grafana no código da API.

### Regras
- Logger, tracing e métricas são inicializados no bootstrap. `core/ports.Context` referencia diretamente o logger da infraestrutura e o tipo `trace.Span`.

## Consequências
- Facilita análise de desempenho e diagnósticos.
- Fornece dados comparáveis entre os frameworks analisados.
- Torna o sistema mais fácil de operar em ambiente de benchmark.

## Referências
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Zap](https://github.com/uber-go/zap)