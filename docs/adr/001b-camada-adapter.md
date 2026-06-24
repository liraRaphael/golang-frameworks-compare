# ADR-001b: Detalhamento da camada Adapter

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [adapter, handlers, controllers, validation, clean-arch]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
A camada `adapter` precisa atuar como ponte entre o mundo externo e o domínio. Ela deve receber dados, traduzir para o formato esperado pelo `core` e devolver respostas padronizadas.

## Decisão
A camada `adapter` será responsável por:

- handlers/controllers para entrada de requisições;
- validação de dados de entrada e saída;
- apresentação de resultados em formatos adequados;
- tradução de erros do domínio para respostas HTTP ou gRPC;
- integração com o handler padronizado da aplicação.

### Estrutura sugerida
- `adapter/controller`: controllers que receberá as chamadas da API.
  - Recebe `context`, `*Request`
- `adapter/services`: responsável pela chamada de outros serviços.
- `adapter/validation`: validações reutilizáveis.
- `adapter/presenters`: transformação de resultado em resposta.

## Consequências
- O domínio permanece desacoplado de HTTP e transporte.
- Os adapters podem ser trocados sem reescrever as regras de negócio.
- A camada fica mais fácil de testar com mocks de use cases e portas.
