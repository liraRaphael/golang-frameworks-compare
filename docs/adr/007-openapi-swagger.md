# ADR-007: Geração automática de documentação OpenAPI/Swagger

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [openapi, swagger, docs, generation, swaggest, jsonschema]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-004: Roteamento](004-roteamento.md)

## Contexto
O projeto compara vários frameworks web em Go; o registry OpenAPI usa metadados de rota e tipos de entrada/saída. O schema de erro não é registrado automaticamente.

## Decisão
`api/infra/docs/swagger` gera OpenAPI com `github.com/swaggest/openapi-go` e serve Swagger UI por meio de `github.com/swaggest/swgui/v5`. O registro é composto no `main.go` e recebe metadados registrados com as rotas.

### Como usar a biblioteca
1. Definir os tipos e tags usados como entrada/saída da rota.
2. Registrar rotas com `swagger.RouteMetadata` (summary, description, tags, input, output e params).
3. `StandardRouter.AddRoute` encaminha os metadados ao registry OpenAPI.
4. `StartSwaggerUI` publica `/openapi.json` e a UI em `/docs/`, no servidor separado configurado pela variável `APP_SWAGGER_PORT` (padrão `8082`).

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
- Os tipos atuais `HelloRequest` e `HelloOutput` usam tags `title`, `description` e `example`; não há exigência implementada de todas as tags listadas acima.
- O bootstrap fornece summary, description, tags, input e output para `/health` e `/hello`.
- O registry atual não registra automaticamente o schema de erros.

### Interface
O código usa `swaggest/openapi-go` e `swaggest/swgui`; não usa `swaggo/swag`.

## Consequências
- A documentação de rotas é gerada dos metadados definidos no bootstrap; é necessário manter esses metadados e o contrato de execução sincronizados.
- A manutenção do contrato da API se torna mais simples.
- O projeto ganha uma base forte para testes, geração de SDKs e integração com clientes.

## Referências
- https://github.com/swaggest/openapi-go
- https://github.com/swaggest/jsonschema-go