# ADR-005: Observability

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [logs, otel, observability, metrics, tracing, jaeger, prometheus, grafana]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](./001-estrutura-app.md)

## Contexto
Para comparar frameworks com dados confiáveis, é necessário instrumentar a aplicação com logs, métricas e tracing. Essa camada deve permitir identificar gargalos de CPU, latência, uso de memória e erros de execução, sem depender do framework web escolhido.

## Decisão
Será criada uma camada de observabilidade em `infra` para coleta de logs, métricas e traces, com foco em comparação entre frameworks.

### Logs
- Uso de `go.uber.org/zap` para geração de logs estruturados.
- Cada log deve incluir timestamp, nível, mensagem, request ID, trace ID, nome do componente e duração da operação.
- O contexto da requisição deve carregar identificadores de trace e duração inicial.

### Tracing
- Uso de `github.com/open-telemetry/opentelemetry-go`.
- A aplicação deve criar spans para operação principal, chamadas externas, persistência e tratamento de erros.
- Atributos-chave devem ser registrados para facilitar análise em Jaeger/Tempo/OTLP.

### Métricas
- Exposição de métricas de requisição, latência, erros e throughput.
- Integração com Prometheus/Grafana deve ser prevista.
- Métricas devem ser coletadas em pontos estratégicos do fluxo HTTP e das operações de negócios.

### Regras
- A observabilidade deve ser implementada em `infra` e consumida por `adapter` e `core` por meio de abstrações.
- O `core` não deve depender diretamente de bibliotecas de logging ou tracing.

## Consequências
- Facilita análise de desempenho e diagnósticos.
- Fornece dados comparáveis entre os frameworks analisados.
- Torna o sistema mais fácil de operar em ambiente de benchmark.

## Referências
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Zap](https://github.com/uber-go/zap)