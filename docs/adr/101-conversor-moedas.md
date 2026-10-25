# ADR-101: Implementação do cenário Conversor de Moedas

**Status**: Proposto — cenário não implementado em `api`

**Data**: 2026-08-09

**Tags**: [benchmark, conversor, moedas, validacao, serializacao, request, response, core, app]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001a: Detalhamento da camada Core](001a-camada-core.md)
- [ADR-001b: Detalhamento da camada App](001b-camada-adapter.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)
- [ADR-003: Handler](003-handler.md)
- [ADR-004: Roteamento](004-roteamento.md)
- [ADR-006: Estratégia de benchmarking e testes](006-benchmarking-e-testes.md)
- [ADR-007: Geração automática de documentação OpenAPI/Swagger](007-openapi-swagger.md)
- [ADR-009: Padrões de Request, Response e HttpParam no Core](009-request-response-core.md)
- [ADR-010: Injeção de Dependência (DI)](010-injecao-dependencia.md)
- [ADR-012: Mecanismo de Validação de Dados](012-validacao-campo.md)
- [ADR-013: Implementação dos frameworks web comparados](013-implementacao-frameworks-web.md)

## Contexto
Este ADR especifica um cenário futuro listado no [README](../../readme.md); não descreve uma funcionalidade presente. A API registra atualmente apenas `/health` e `/hello`. Não existem rota, request, response, use case, presenter ou testes de conversão de moedas em `api`.

O cenário básico deve permitir comparar o custo de roteamento, binding de JSON, validação, processamento local e serialização de resposta, sem banco de dados ou dependência externa. Uma taxa fixa mantém os resultados determinísticos e evita que disponibilidade e latência de terceiros contaminem a comparação.

As ADRs arquiteturais são uma referência de implementação, não garantia de comportamento já disponível. Em particular:
- erros sem mapeamento explícito no handler atual resultam no fallback HTTP 500 (ADR-003);
- os adapters não têm comportamento completamente equivalente para binding, path params, cookies e escrita de resposta, e `Use` ainda não executa middlewares (ADR-013);
- `AddTypedRoute` pode converter body e parâmetros por marshal/unmarshal JSON (ADR-009);
- a infraestrutura de benchmarks e a coleta de métricas descritas na ADR-006 não estão integralmente implementadas.

## Decisão
Se implementado, o cenário será uma API local, determinística e sem chamadas externas. As regras abaixo definem o contrato esperado e deverão ser cobertas por testes antes de qualquer resultado ser comparado.

### Contrato HTTP
- Método e rota: `POST /currency/convert`.
- Content-Type: `application/json`.
- O corpo contém `from`, `to` e `value`; não há parâmetros de rota, query ou headers de negócio.
- `from` e `to` aceitam exclusivamente os códigos `USD` e `BRL`, em letras maiúsculas. Não normalizar espaços ou caixa automaticamente.
- `value` é um número JSON não negativo, obrigatório; zero é um valor válido.
- Campos ausentes, tipos JSON incorretos, códigos desconhecidos, números negativos ou valores cujo resultado não seja finito devem ser rejeitados antes da conversão.

Exemplo:

```json
{
  "from": "USD",
  "to": "BRL",
  "value": 10.00
}
```

O request de domínio deverá preservar a distinção entre campo ausente e zero. Portanto, não usar somente `validate:"required"` em um `float64` não ponteiro: para a validator, o valor zero é o valor vazio. Usar uma representação que detecte presença (por exemplo, `*float64`) ou validação equivalente e validar o valor numérico explicitamente. As tags de validação descrevem restrições, mas sua execução permanece na camada `app`, conforme a ADR-012.

### Regra de cotação e precisão
A cotação é constante durante toda a execução:
- `USD` para `BRL`: multiplicador `6.31` (1 USD = 6.31 BRL).
- `BRL` para `USD`: multiplicador `1 / 6.31`.
- Mesma moeda: multiplicador `1`.

O `quote` da resposta representa o multiplicador usado para aquela direção; portanto, para moedas iguais seu valor é `1`, não a cotação USD/BRL. Aplicar `value * quote` e arredondar o resultado para duas casas decimais, ao centavo mais próximo (empates para longe de zero, conforme `math.Round`). O valor recebido pode ter mais de duas casas; o valor numérico da resposta é quantizado para centavos. Como o JSON representa números, zeros finais (por exemplo, `63.10` versus `63.1`) não são garantidos na serialização; clientes que precisem exibir duas casas devem formatar o valor. Rejeitar o resultado caso a operação produza NaN ou infinito.

O contrato usa números JSON e `float64` para manter o cenário simples e focado no custo dos frameworks. Esse arredondamento é suficiente para o benchmark proposto, mas ponto flutuante não é apropriado para contabilização financeira de produção. Uma aplicação financeira real deverá adotar uma representação decimal ou unidades monetárias inteiras e revisar esta decisão.

### Contrato da resposta
Em caso de sucesso, devolver HTTP 200 e JSON com:
- `value`: valor convertido, arredondado para duas casas;
- `from`: código de origem recebido;
- `to`: código de destino recebido;
- `quote`: multiplicador aplicado.

```go
type ConvertCurrencyResponse struct {
    Value float64 `json:"value" title:"Valor convertido" description:"Valor arredondado para duas casas decimais" example:"63.10"`
    From  string  `json:"from" title:"Moeda de origem" example:"USD"`
    To    string  `json:"to" title:"Moeda de destino" example:"BRL"`
    Quote float64 `json:"quote" title:"Cotação aplicada" description:"Multiplicador usado na conversão" example:"6.31"`
}
```

Exemplos de resultados esperados:

| Entrada (`from`, `to`, `value`) | Resposta (`value`, `quote`) |
|---|---|
| `USD`, `BRL`, `10` | `63.10`, `6.31` |
| `BRL`, `USD`, `63.10` | `10.00`, `1 / 6.31` |
| `USD`, `USD`, `10` | `10.00`, `1` |
| `BRL`, `BRL`, `0` | `0.00`, `1` |

### Erros HTTP
- JSON malformado, corpo ausente, campo obrigatório ausente ou inválido e moedas não suportadas: HTTP 400.
- Falhas internas inesperadas: HTTP 500, sem expor detalhes internos.
- Respostas de erro devem usar `responses.ErrorResponse` e os campos existentes `codigo`, `mensagem` e, quando aplicável, `detalhes`.
- A implementação deve registrar um mapeamento explícito para validação e parsing (ou produzir a resposta equivalente no fluxo do adapter); não depender do fallback 500 do handler atual para erros de entrada. Erros desconhecidos continuam sendo 500.

### Organização por responsabilidade
Os caminhos são uma proposta para implementação futura, não arquivos existentes:

#### Core
- `api/core/domain/requests/currency.go`: request tipado com tags JSON e metadados de schema.
- `api/core/domain/responses/currency.go`: resposta tipada da conversão.
- `api/core/domain/enums/currency.go` (ou constantes de domínio equivalentes): códigos de moeda suportados.
- `api/core/usecases/currency.go`: regra de cotação, cálculo e arredondamento, sem dependência de HTTP ou framework.
- `api/core/ports`: declarar somente as portas específicas que forem necessárias; seguir o padrão de construtores e injeção manual da ADR-010.

#### App
- `api/app/controller/currency.go`: validar a entrada e orquestrar o use case.
- `api/app/validation`: reutilizar `ports.Validator` e `PlaygroundValidator`; não implementar um motor de validação específico por framework.
- `api/app/presenter/currency.go`: adaptar a saída do caso de uso ao contrato de response, se essa separação for necessária no fluxo escolhido.
- `api/app/handler`: mapear erros de entrada para respostas HTTP 400 padronizadas.

#### Infra e composição
- Não criar cliente HTTP, banco, cache ou configuração de cotação para este cenário.
- Registrar `POST /currency/convert` no roteador comum, de modo que a mesma rota e os mesmos metadados sejam usados por Gin, Fiber, Echo, Chi, Beego e net/http.
- Compor validator, controller, use case e presenter manualmente em `api/main.go`, seguindo a injeção por construtor.
- Registrar metadados OpenAPI de summary, description, tags, input e output; a documentação não registra automaticamente o schema de erros (ADR-007).

### Fluxo esperado
1. O adapter recebe o POST, faz o binding do JSON e constrói o request genérico da aplicação.
2. A rota tipada, se usada, converte a entrada ao tipo de domínio; qualquer marshal/unmarshal intermediário deve ser considerado parte do caminho medido.
3. O controller valida presença, moedas e valor antes de chamar o caso de uso.
4. O use case aplica a taxa fixa, calcula e arredonda o resultado.
5. O presenter prepara o contrato de resposta e o adapter serializa o JSON com status 200.
6. Erros de parsing ou validação são mapeados para 400; falhas desconhecidas são mapeadas para 500.

### Testes e comparação
Antes de usar o cenário em benchmark, cobrir:
- conversão USD→BRL, BRL→USD e conversão para a mesma moeda;
- zero, valores com mais de duas casas e casos de arredondamento;
- campo ausente, JSON malformado, tipo incorreto, moeda inválida, valor negativo e resultado não finito;
- status, payload e headers de sucesso/erro em cada adapter.

Os testes unitários do use case não devem depender de servidor web. Testes de integração devem exercer o contrato HTTP em cada adapter. Como os adapters diferem em binding e tratamento de erros, compartilhar a regra de domínio não substitui a verificação de equivalência observável.

No benchmark, enviar os mesmos corpos e a mesma distribuição de entradas válidas e inválidas a todos os frameworks. Separar resultados por status, não contar respostas de erro como sucesso e registrar configuração e recursos do ambiente conforme ADR-006. A taxa fixa elimina latência externa, mas não elimina diferenças no binding, na validação, no adaptador tipado ou na serialização. Não atribuir resultados a métricas de aplicação (por exemplo, CPU/memória por requisição) sem instrumentação efetivamente configurada.

## Consequências
- O cenário permite comparar processamento local sem variabilidade de rede externa.
- Taxa, arredondamento, códigos aceitos e comportamento para zero ficam definidos e reproduzíveis.
- O uso de `float64` simplifica o benchmark, mas limita a adequação do exemplo a aplicações financeiras reais.
- A paridade entre frameworks depende de testes HTTP explícitos; a interface comum de adapter não garante comportamento idêntico.
- O cenário só deve ser considerado implementado após o registro da rota, dos mapeamentos de erro, da documentação, dos testes e do benchmark descritos nesta decisão.

## Referências
1. [README — cenários de benchmark](../../readme.md)
2. [ADR-003: Handler](003-handler.md)
3. [ADR-004: Roteamento](004-roteamento.md)
4. [ADR-006: Estratégia de benchmarking e testes](006-benchmarking-e-testes.md)
5. [ADR-007: Geração automática de documentação OpenAPI/Swagger](007-openapi-swagger.md)
6. [ADR-009: Padrões de Request, Response e HttpParam no Core](009-request-response-core.md)
7. [ADR-010: Injeção de Dependência (DI)](010-injecao-dependencia.md)
8. [ADR-012: Mecanismo de Validação de Dados](012-validacao-campo.md)
9. [ADR-013: Implementação dos frameworks web comparados](013-implementacao-frameworks-web.md)
