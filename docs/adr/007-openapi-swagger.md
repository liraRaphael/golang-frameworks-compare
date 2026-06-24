# ADR-007: Geração automática de documentação OpenAPI/Swagger

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [openapi, swagger, docs, generation, swaggest, jsonschema]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-004: Roteamento](004-roteamento.md)

## Contexto
Como o projeto compara vários frameworks web em Go, a documentação da API precisa ser gerada automaticamente a partir das rotas, dos tipos de request/response e dos erros esperados. Isso evita divergência entre contrato e implementação e reduz manutenção manual.

## Decisão
A documentação OpenAPI/Swagger será gerada automaticamente com a biblioteca `github.com/swaggest/openapi-go`, usando metadados declarados nos tipos e nas rotas.

### Como usar a biblioteca
1. Definir estruturas de request, response e erros com tags de schema.
2. Registrar cada rota com metadata de operação, path, método, summary, description e exemplos.
3. Gerar um documento OpenAPI a partir do registrador de rotas.
4. Expor o documento em um endpoint dedicado ou em um servidor separado para documentação.

### Tags de campo recomendadas
Para cada campo, o projeto deverá usar tags de schema que descrevam o contrato de forma explícita. Exemplos recomendados:

- `title`: título do campo
- `description`: descrição detalhada
- `example`: exemplo de valor
- `examples`: lista de exemplos
- `default`: valor padrão
- `deprecated`: marca o campo como obsoleto
- `readOnly`: campo somente leitura
- `writeOnly`: campo somente escrita
- `pattern`: regex de validação
- `format`: formato do valor, como `date-time`, `uuid`, `email`
- `enum`: lista de valores permitidos
- `required`: marca o campo como obrigatório
- `nullable`: indica que o valor pode ser nulo

### Regras de uso
- Os tipos do `core` devem ser ricos em metadados para que a geração de documentação seja automática.
- O `adapter` deve fornecer a descrição operacional da rota, como summary e tags.
- Os erros devem ter representação clara no schema para facilitar a documentação e o consumo por clientes.

## Consequências
- A documentação fica sincronizada com o código.
- A manutenção do contrato da API se torna mais simples.
- O projeto ganha uma base forte para testes, geração de SDKs e integração com clientes.

## Referências
- https://github.com/swaggest/openapi-go
- https://github.com/swaggest/jsonschema-go