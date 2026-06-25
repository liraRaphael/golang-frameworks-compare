# ADR-010: Injeção de Dependência (DI)

**Status**: Aceito

**Data**: 2026-06-25

**Tags**: [di, infra, desacoplamento]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)

## Contexto
Para manter o desacoplamento proposto na ADR-001 e facilitar a troca de implementações concretas, como substituir um repositório real por um mock em testes, é necessário um padrão de injeção de dependência consistente. Sem isso, a composição da aplicação fica acoplada a detalhes de implementação e a manutenção dos componentes se torna mais difícil.

## Decisão
Utilizaremos Injeção de Dependência via Construtor (Constructor Injection) de forma manual. Não utilizaremos containers de DI, como `uber-dig` ou `google/wire`, para preservar a simplicidade e reduzir o overhead no startup da aplicação, especialmente em cenários de benchmark de cold-start.

### Padrão adotado
- Cada componente principal, como controller, use case e repository, deve expor uma função `New...` que receba suas dependências por interfaces.
- A composição da aplicação deve ocorrer no `main.go` ou em um pacote dedicado de bootstrap, como `api/cmd/bootstrap` ou `api/app/bootstrap`.
- As dependências devem ser definidas em termos de interfaces pertencentes à camada `core` ou `adapter`, evitando acoplamento a implementações concretas.
- Os componentes devem permanecer fáceis de instanciar em testes, permitindo a substituição de dependências sem alterações no fluxo principal.

## Consequências
- O desacoplamento entre camadas aumenta e a troca de implementação fica mais simples.
- A aplicação se torna mais testável e mais previsível em cenários de benchmark.
- O projeto evita a complexidade e o overhead de containers de DI, preservando performance e simplicidade.

## Referências
1. https://martinfowler.com/articles/injection.html
2. https://go.dev/doc/effective_go
