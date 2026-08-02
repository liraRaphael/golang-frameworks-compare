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

- `domain/entities`: entidades de negócio e regras internas.
- `domain/requests`: tipos de entrada para os casos de uso.
- `domain/responses`: tipos de saída para uso externo.
- `domain/errors`: erros de domínio e erros de negócio.
- `domain/enums`: constantes e tipos enumerados relevantes.
- `usecases`: implementação dos fluxos principais da aplicação.
- `ports`: interfaces para repositórios, clientes externos e validações.

### Regras
- Nenhuma dependência de `adapter` ou `infra` é permitida.
- Todas as operações de negócio devem receber `ports.Context` como contrato de borda, com o `Context` em `core/ports` compondo `context.Context`.
- Interfaces devem ser pequenas, específicas e fáceis de mockar.

## Consequências
- A lógica de negócio fica centralizada e reutilizável.
- Testes unitários se tornam mais simples e menos frágeis.
- Mudanças externas não impactam diretamente o domínio.
