# ADR-001b: Detalhamento da camada App

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [app, handlers, controllers, validation, clean-arch]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
Não existe um diretório `api/adapter` na implementação atual. A coordenação que este ADR originalmente atribuía a essa camada está em `api/app`; a tradução específica entre os frameworks e os contratos de aplicação está em `api/infra/frameworks`.

## Decisão
A camada `app` é responsável por:

- controllers para entrada das requisições;
- validação de dados de entrada;
- apresentação de resultados em formatos adequados;
- retorno de erros ao adapter de framework e integração com o handler padronizado; o handler atual aplica fallback 500 quando não há mapeamento;
- integração com o handler padronizado da aplicação.

Os adapters de `infra/frameworks/{beego,chi,echo,fiber,gin,nethttp}` fazem parsing do corpo e parâmetros HTTP, chamam controllers e escrevem responses. A interface `ports.FrameworkAdapter` declara `RegisterRoute`, `Use` e `Start`.

### Estrutura sugerida
- `app/controller`: controllers que recebem as chamadas da API.
  - Recebe `ports.Context` e `requests.Request[...]`.
- `app/handler`: resolução de respostas de sucesso e erro.
- `app/validation`: implementação `PlaygroundValidator`, baseada em `go-playground/validator`.
- `app/presenter`: transformação de resultado em resposta.
- `infra/frameworks/<nome>`: integração de entrada/saída específica de cada framework.

## Consequências
- Controllers e use cases evitam depender diretamente de tipos concretos de framework.
- Os adapters podem ser trocados sem reescrever as regras de negócio.
- A camada fica mais fácil de testar com mocks de use cases e portas.
