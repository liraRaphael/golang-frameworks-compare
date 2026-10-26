# ADR-002: Cliente

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [API, cliente, services, http, grpc, integração]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
Para permitir a comunicação com serviços externos sem acoplar o domínio aos detalhes de transporte, é necessário definir um contrato de cliente reutilizável. Esse contrato deve ser compatível com chamadas HTTP e gRPC, permitir configuração de timeout, headers, autenticação e tratamento de erros, e manter o domínio independente de bibliotecas específicas.

## Decisão
O contrato existente é `ports.Client` em `core/ports/client.go`; implementações ficam em `infra/clients`.

### Contrato
- O pacote `core/ports` deverá expor uma interface `Client` com operações genéricas para execução de chamadas externas.
- `Client.Do` recebe `ports.Context`, `enums.HttpMethod`, URL e `requests.Request[any, any]`, retornando `responses.Response[any, any]` e `error`.
- O uso de `core/domain/enums` deverá incluir um enum `HttpMethod` com os métodos suportados: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS` e `HEAD`.
- `Client.Do` permanece não genérico: mesmo no Go 1.27, métodos de interface não podem declarar seus próprios parâmetros de tipo. Métodos genéricos são permitidos em tipos concretos, mas não tornam a interface `Client` capaz de exigir uma implementação genérica.
- Quando um consumidor precisar de um corpo tipado, a conversão deve ser explícita e retornar erro em caso de falha. Ela pode ser oferecida por uma função genérica de pacote que recebe `Client`; não usar type assertion insegura nem assumir que o valor dinâmico retornado como `any` já tem o tipo de destino. O contrato e o helper devem preservar status, headers e demais metadados da resposta.
- A implementação da conversão deve evitar round-trip JSON desnecessário sempre que possível: preferir preservar o payload bruto (`json.RawMessage`) até a decodificação para o tipo esperado, sem acoplar o contrato a tipos específicos de serviço.

### Implementações
- `infra/clients/nethttp` usa `net/http`; `infra/clients/fasthttp` usa `fasthttp`.
- `infra/clients/grpc` existe, mas sua chamada genérica `Do` ainda não invoca um serviço: retorna `ClientError` com status 501. A criação de conexão usa credenciais inseguras.
- Os clientes HTTP serializam o body JSON, copiam headers e devolvem o status, headers e body na resposta genérica. Se o body não for JSON, o cliente HTTP devolve seu conteúdo como string.
- Não há evidência na implementação de testes comparativos específicos para HTTP/1.1 e HTTP/2 cleartext.

### Erros
- `core/domain/errors.ClientError` contém status code, body, mensagem e erro subjacente.
- Os clientes HTTP retornam erros de transporte em `ClientError`; as respostas HTTP, inclusive status não-2xx, são devolvidas como response, sem conversão automática para erro.
- O contrato não deve depender de erros específicos de `net/http`, `grpc` ou outra biblioteca.

### Injeção
- O contrato permite injeção por construtor. O `main.go` compõe um cliente HTTP, serviço e controller para a consulta de CEP; outros fluxos podem injetar o mesmo contrato conforme necessário.

### Regras
- O `core` não deve importar bibliotecas de transporte.
- O `app` pode orquestrar o contrato do cliente, mas não deve conhecer detalhes de `net/http`, `fasthttp` ou SDKs de transporte.
- O `infra` deve ser a única camada responsável pela comunicação real com o mundo externo.

## Consequências
- O domínio fica mais limpo e testável.
- O contrato único permite trocar implementações HTTP no ponto de composição; a implementação gRPC genérica ainda não oferece equivalência funcional.
- O tratamento de erro fica padronizado e mais simples de mapear em respostas da API.
- O tipo `any` mantém o contrato reutilizável, mas transfere ao consumidor a responsabilidade de decodificar o corpo com segurança. A tipagem do corpo não deve depender de uma type assertion que possa causar panic.

## Referências
- [gRPC](https://grpc.io/docs/languages/go/)
- [net/http](https://pkg.go.dev/net/http)