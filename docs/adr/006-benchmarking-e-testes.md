# ADR-006: Estratégia de benchmarking e testes

**Status**: Proposto

**Data**: 2026-06-24

**Tags**: [benchmark, testes, jmeter, performance, docker]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [README](../../readme.md)

## Contexto
O projeto pretende comparar frameworks web Go em termos de desempenho, consumo de recursos, complexidade de integração e consistência. Para isso, é necessário definir uma estratégia de testes reutilizável, com cenários claros, métricas observáveis e execução reproduzível.

## Decisão
A estratégia de validação do projeto será composta por três camadas:

1. Testes unitários do domínio e dos casos de uso.
2. Testes de integração de adapter e infraestrutura.
3. Benchmarks end-to-end com JMeter e observabilidade integrada.

### Testes unitários
- Cobrir regras de negócio em `core`.
- Evitar dependência de framework, banco ou cliente externo.
- Priorizar cenários de sucesso, falha e edge cases.

### Testes de integração
- Validar roteamento, parsing de request, handler e serialização de response.
- Usar mocks ou stubs para clientes externos.
- Manter o teste focado no comportamento da camada de adapter/infra.

### Benchmarks
- Executar cenários básicos e complexos com JMeter.
- Medir latência, throughput, uso de CPU, memória e erro por requisição.
- Rodar os testes em containers Docker para reduzir variação de ambiente.

### Cenários mínimos
- Hello World
- Conversor de Moedas
- Busca de CEP
- Ingresso Esgotado
- Pix Black Friday

### Métricas obrigatórias
- Latência média e p95
- Throughput por segundo
- Uso de CPU
- Uso de memória heap
- Erros HTTP e status distribuídos
- Tempo de resposta do banco e de clientes externos

## Consequências
- O processo de comparação entre frameworks fica mais confiável.
- As decisões de arquitetura passam a ser apoiadas por evidências reais.
- A evolução do projeto ganha um padrão claro de validação contínua.

## Referências
- [JMeter](https://jmeter.apache.org/)
- [Go testing](https://pkg.go.dev/testing)
