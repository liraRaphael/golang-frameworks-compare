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
- `usecases`: casos de uso, que recebem requests do domínio e retornam o corpo tipado da resposta e um erro, sem construir `responses.Response` nem definir status HTTP.
- `ports`: contratos de controller, handler, middleware, presenter, validator, cliente, contexto, router e adapters de framework.

### Regras
- A implementação atual não está totalmente isolada de `infra`: `ports.Context` referencia `infra/observability/logger` e OpenTelemetry; `ports.Router` referencia `infra/docs/swagger`.
- Use cases implementam regras da aplicação sem depender diretamente de um framework web; não serializam HTTP/JSON nem escolhem status HTTP.
- `CurrencyConversionRequest` concentra as operações puras específicas do request de conversão (`CurrencyRate` e `CurrencyQuote`); `CurrencyQuote` calcula o valor convertido e arredondado, enquanto `CurrencyRate` fornece o multiplicador da resposta.
- Interfaces devem ser pequenas, específicas e fáceis de mockar.

### Implementação presente
O use case de Hello delega a produção da saída à porta `HelloPresenter`. O conversor de moedas retorna `ConvertCurrencyResponse` como corpo tipado e erro; o controller o envolve em `responses.Response` e escolhe o status HTTP. Os tipos `HelloRequest`, `HelloOutput`, `ErrorResponse`, `ClientError` e `FieldValidationError` vivem nos pacotes existentes de `core/domain`.

## Consequências
- A lógica de negócio fica centralizada e reutilizável.
- Testes unitários se tornam mais simples e menos frágeis.
- Mudanças externas não impactam diretamente o domínio.
