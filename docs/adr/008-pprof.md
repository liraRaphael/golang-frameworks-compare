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
`api/infra/diagnostic.StartPprof` inicia `net/http/pprof` em goroutine usando `http.ListenAndServe` e o mux padrão. `main.go` chama essa função sempre; o endereço usa a variável `APP_PPROF_PORT` (padrão `8081`).

### Regras
- O código está em `infra/diagnostic`, não em `infra/pprof`.
- O servidor é separado da API principal e a porta vem de `AppConfig.PprofPort`.
- Não existe opção para desabilitar o servidor na configuração atual.
- Falhas de `ListenAndServe` são registradas pelo logger dentro da goroutine; `StartPprof` não retorna erro ao bootstrap.

### Endpoints esperados
- `/debug/pprof/`
- `/debug/pprof/cmdline`
- `/debug/pprof/profile`
- `/debug/pprof/symbol`
- `/debug/pprof/trace`

### Integração
- O `pprof` é iniciado incondicionalmente, inclusive fora de benchmarks.
- Não há integração automática dos perfis com JMeter, Prometheus ou tracing.

## Consequências
- O projeto passa a ter uma visão mais detalhada de consumo de CPU e memória.
- A análise de desempenho se torna mais objetiva e reproduzível.
- Os benchmarks ganham maior valor técnico e confiabilidade.
