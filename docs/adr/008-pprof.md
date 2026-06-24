# ADR-008: Pprof e diagnóstico de performance

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [pprof, performance, profiling, debug, observability]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-005: Observability](005-observability.md)

## Contexto
Para comparar frameworks web de forma justa, é necessário analisar o comportamento em CPU, alocação de memória e gargalos de execução. O `pprof` é uma ferramenta padrão do ecossistema Go para coletar esse tipo de informação.

## Decisão
O projeto deve expor um servidor separado para `pprof`, com configuração própria e isolada do fluxo principal da API.

### Regras
- O `pprof` deve ficar em `infra/pprof` ou equivalente.
- Deve ser iniciado em um porta separada, por padrão `8081`, configurável por ambiente.
- Deve ser possível habilitar ou desabilitar o endpoint sem alterar o código principal.
- O fluxo da aplicação principal não deve ser impactado pela presença do profiling.

### Endpoints esperados
- `/debug/pprof/`
- `/debug/pprof/cmdline`
- `/debug/pprof/profile`
- `/debug/pprof/symbol`
- `/debug/pprof/trace`

### Integração
- O `pprof` poderá ser ativado em modo de desenvolvimento ou em execuções de benchmark.
- As métricas obtidas devem ser correlacionadas com os resultados de JMeter, Prometheus e traces do OpenTelemetry.

## Consequências
- O projeto passa a ter uma visão mais detalhada de consumo de CPU e memória.
- A análise de desempenho se torna mais objetiva e reproduzível.
- Os benchmarks ganham maior valor técnico e confiabilidade.
