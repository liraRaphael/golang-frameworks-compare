# ADR-010: Injeção de Dependência (DI)

**Status**: Aceito

**Data**: 2026-06-25

**Tags**: [di, infra, desacoplamento]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)

## Contexto
Para manter o desacoplamento proposto na ADR-001 e facilitar a substituição de dependências em testes, é necessário um padrão de injeção de dependência consistente. Sem isso, a composição da aplicação fica acoplada a detalhes de implementação e a manutenção dos componentes se torna mais difícil.

## Decisão
Utilizaremos Injeção de Dependência via Construtor (Constructor Injection) de forma manual. Não utilizaremos containers de DI, como `uber-dig` ou `google/wire`, para preservar a simplicidade e reduzir o overhead no startup da aplicação, especialmente em cenários de benchmark de cold-start.

### Padrão adotado
- Componentes como `HelloController`, `HelloUseCase` e clientes possuem construtores `New...`; nem todo componente tem dependências ou um construtor que receba interfaces.
- Atualmente a composição ocorre em `api/main.go`; não há pacote separado de bootstrap.
- Os controllers/use cases recebem algumas dependências por interfaces de `core/ports`; o bootstrap instancia implementações de `app` e `infra`. A separação não é completa: alguns contratos de `core/ports` importam tipos de `infra`.
- Os componentes devem permanecer fáceis de instanciar em testes, permitindo a substituição de dependências sem alterações no fluxo principal.

## Consequências
- A composição manual torna explícitas as dependências, embora as referências de `core/ports` à infraestrutura limitem o desacoplamento estrito.
- A aplicação se torna mais testável e mais previsível em cenários de benchmark.
- O projeto evita a complexidade e o overhead de containers de DI, preservando performance e simplicidade.

## Referências
1. https://martinfowler.com/articles/injection.html
2. https://go.dev/doc/effective_go
