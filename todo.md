# Plano de próximos passos

## Fase 1 — Estrutura inicial
- [x] Revisar ADRs e alinhar a documentação com o README.
- [ ] Criar a estrutura inicial do módulo de API em `api/` com `core`, `adapter` e `infra`.
- [ ] Definir os pacotes base para domínio, requests, responses, errors e enums.
- [ ] Criar o primeiro endpoint de exemplo (Hello World) em um framework inicial.

## Fase 2 — Contratos e integração
- [ ] Implementar as interfaces de repositório, cliente externo e validação.
- [ ] Criar o handler padronizado para sucesso e erro.
- [ ] Definir o roteamento abstrato e a integração com um framework concreto.

## Fase 3 — Infraestrutura
- [ ] Configurar Docker Compose com PostgreSQL, observabilidade e ferramentas de benchmark.
- [ ] Implementar logging estruturado e tracing inicial.
- [ ] Expor métricas básicas para Prometheus/Grafana.

## Fase 4 — Benchmark
- [ ] Definir os cenários de teste e a carga inicial do JMeter.
- [ ] Rodar benchmark para os frameworks comparados.
- [ ] Consolidar resultados e atualizar o README com evidências.
