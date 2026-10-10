# ADR-001a: Detalhamento da camada Core

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [core, domain, usecases, portas, clean-arch]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
A camada `core` precisa ser bem definida para manter as regras de negócio isoladas e fáceis de testar. Ela deve concentrar o domínio da aplicação sem depender de frameworks, banco de dados ou transporte HTTP.

## Decisão
A camada `core` será organizada em subpacotes com responsabilidades claras:

- `domain`: tipos HTTP, requests, responses, enums e erros; atualmente não há pacote `domain/entities`.
- `usecases`: implementação do caso de uso `HelloUseCase`.
- `ports`: contratos de controller, handler, middleware, presenter, validator, cliente, contexto, router e adapters de framework.

### Regras
- A implementação atual não está totalmente isolada de `infra`: `ports.Context` referencia `infra/observability/logger` e OpenTelemetry; `ports.Router` referencia `infra/docs/swagger`.
- `HelloUseCase.Execute` recebe `ports.Context` e um tipo de request de domínio; não depende diretamente de um framework web.
- Interfaces devem ser pequenas, específicas e fáceis de mockar.

### Implementação presente
O use case de Hello delega a produção da saída à porta `HelloPresenter`. Os tipos `HelloRequest`, `HelloOutput`, `ErrorResponse`, `ClientError` e `FieldValidationError` vivem nos pacotes existentes de `core/domain`. Não há conversor de moeda implementado em `core`.

## Consequências
- A lógica de negócio fica centralizada e reutilizável.
- Testes unitários se tornam mais simples e menos frágeis.
- Mudanças externas não impactam diretamente o domínio.
