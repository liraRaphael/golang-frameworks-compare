# ADR-101: Implementação do cenário Conversor de Moedas

**Status**: Proposto — cenário não implementado em `api`

**Data**: 2026-08-09

**Tags**: [benchmark, conversor, moedas, validacao, serializacao, request, response, core, app]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001a: Detalhamento da camada Core](001a-camada-core.md)
- [ADR-001b: Detalhamento da camada App](001b-camada-adapter.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)
- [ADR-009: Padrões de Request, Response e HttpParam no Core](009-request-response-core.md)
- [ADR-010: Injeção de Dependência (DI)](010-injecao-dependencia.md)
- [ADR-013: Implementação dos frameworks web comparados](013-implementacao-frameworks-web.md)

## Contexto
Este ADR descreve uma proposta de cenário, não uma funcionalidade presente. A implementação em [api](../../api) contém atualmente apenas os endpoints `/health` e `/hello`; não existem request, response, use case, controller ou rota de conversão de moedas. Portanto, os tipos, a cotação e as regras abaixo são requisitos propostos e não devem ser tratados como comportamento executável.

Esse cenário é importante porque permite medir, em um mesmo contrato de API, a sobrecarga de:
- parsing de JSON de entrada;
- validação de campos obrigatórios e regras de domínio;
- conversão de tipos e serialização de resposta;
- tratamento de erros de request inválido.

Como não há integração externa, a comparação fica mais justa e o custo medido reflete a infraestrutura web e o processamento de transformação do payload, e não a latência de um cliente remoto.

## Decisão
Se o cenário for implementado futuramente, a proposta é fazê-lo como API de conversão local determinística e sem dependência de serviços externos.

### Contrato da entrada
O request terá os campos abaixo:
- `from`: moeda de origem; aceita apenas `USD` ou `BRL`.
- `to`: moeda de destino; aceita apenas `USD` ou `BRL`.
- `value`: valor numérico em ponto flutuante para conversão.

O contrato será modelado em `core` para manter o domínio agnóstico de framework e transporte.

```go
type ConvertCurrencyRequest struct {
    From  string  `json:"from" validate:"required,oneof=USD BRL"`
    To    string  `json:"to" validate:"required,oneof=USD BRL"`
    Value float64 `json:"value" validate:"required,gte=0"`
}
```

### Regra de cotação
A cotação será fixa e determinística para toda a execução do benchmark:
- 1 USD = 6.31 BRL
- a conversão inversa será calculada como USD/BRL, isto é, 1 BRL = 1 / 6.31 USD

A equivalência será aplicada como regra de negócio do caso de uso, sem depender de API externa, cache ou configuração dinâmica.

### Regras de conversão
- Se `from == to`, o valor deve ser retornado sem alteração.
- Se `from == "USD" && to == "BRL"`, a conversão será `value * 6.31`.
- Se `from == "BRL" && to == "USD"`, a conversão será `value / 6.31`.
- Qualquer combinação inválida será rejeitada pela validação de domínio e pelo validador da camada adapter.

### Contrato da resposta
A resposta será padronizada com os campos:
- `value`: valor convertido em float;
- `from`: moeda de origem do request;
- `to`: moeda final do request;
- `quote`: valor da cotação utilizada na conversão.

```go
type ConvertCurrencyResponse struct {
    Value float64 `json:"value"`
    From  string  `json:"from"`
    To    string  `json:"to"`
    Quote float64 `json:"quote"`
}
```

### Modelo de implementação
Como proposta futura, os tipos e componentes seriam organizados por responsabilidade. Os caminhos abaixo não existem na API atual:

#### Core
- `api/core/domain/requests/currency.go`
  - define `ConvertCurrencyRequest`
- `api/core/domain/responses/currency.go`
  - define `ConvertCurrencyResponse`
- `api/core/domain/enums/currency.go`
  - define moedas permitidas e constantes de quote
- `api/core/usecases/currency.go`
  - implementa a lógica de conversão em um `UseCase` dedicado
- `api/core/ports/currency.go`
  - expõe contratos de entrada/saída para o caso de uso, se necessário

#### App
- `api/app/validation/currency.go`
  - responsável por validar as regras de domínio e valores permitidos
- `api/app/controller/currency.go`
  - recebe a requisição HTTP, converte para o request do domínio e invoca o use case
- `api/app/presenter/currency.go`
  - transforma o resultado do caso de uso em resposta serializável

#### Infra
- A infraestrutura para esse cenário não deve incluir cliente HTTP, banco de dados, fila ou chamada externa.
- O cenário é intencionalmente “local-only”, então não haverá pacote `api/infra/clients` para essa rota.
- O roteamento e a serialização continuam sendo responsabilidade dos adaptadores de framework em `api/infra/frameworks/*`.

### Padrão de execução do fluxo
1. o framework recebe o payload HTTP;
2. o controller converte a requisição para `ConvertCurrencyRequest`;
3. a validação da aplicação verifica `from`, `to` e `value`;
4. o use case calcula a conversão com a cotação fixa;
5. o presenter prepara a resposta com `value`, `from`, `to` e `quote`;
6. o adapter de framework serializa e escreve a resposta JSON.

### Exemplo conceitual do caso de uso
```go
const (
    USDToBRL = 6.31
)

func ConvertCurrency(value float64, from, to string) (float64, float64) {
    if from == to {
        return value, USDToBRL
    }

    if from == "USD" && to == "BRL" {
        return value * USDToBRL, USDToBRL
    }

    if from == "BRL" && to == "USD" {
        return value / USDToBRL, USDToBRL
    }

    return 0, USDToBRL
}
```

## Consequências
- O benchmark fica mais simples e mais estável, pois remove ruído de rede e latência externa.
- O custo medido passa a refletir a sobrecarga real de binding, validação e serialização por framework.
- A regra de negócio fica clara e reproduzível em todos os frameworks comparados.
- A estrutura de arquivos se torna consistente com a arquitetura em camadas já definida pelas ADRs anteriores.
- A implementação permanece fácil de testar em unitários sem exigir infraestrutura adicional.

## Referências
1. [readme.md](../../readme.md)
2. [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
3. [ADR-009: Padrões de Request, Response e HttpParam no Core](009-request-response-core.md)
4. [ADR-013: Implementação dos frameworks web comparados](013-implementacao-frameworks-web.md)
