# ADR-002: Cliente

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [API, cliente, services, http, grpc, integração]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
Para permitir a comunicação com serviços externos sem acoplar o domínio aos detalhes de transporte, é necessário definir um contrato de cliente reutilizável. Esse contrato deve ser compatível com chamadas HTTP e gRPC, permitir configuração de timeout, headers, autenticação e tratamento de erros, e manter o domínio independente de bibliotecas específicas.

## Decisão
Será criado um contrato de cliente na camada `core` e sua implementação concreta em `infra`.

### Contrato
- O pacote `core/ports` deverá expor uma interface `Client` com operações genéricas para execução de chamadas externas.
- A interface deverá aceitar `context.Context` e um payload de requisição tipado, retornando um payload de resposta tipado e `error`.
- O uso de `core/domain/enums` deverá incluir um enum `HttpMethod` com os métodos suportados: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS` e `HEAD`.

### Implementações
- `infra/clients/http` para chamadas HTTP usando `net/http` ou `fasthttp`.
- `infra/clients/grpc` para chamadas gRPC, se necessário no futuro.
- Cada implementação deve assumir a responsabilidade de serialização, configuração e tratamento de transporte, sem expor detalhes do framework para o domínio.
- a alteração entre grpc ou http não deverá ser sentido
- para http será testados os cenários com http1.1 e http2c

### Erros
- Erros de comunicação devem ser representados por tipos específicos em `core/domain/errors`.
- Erros de resposta com status inesperado devem preservar informações como status code, corpo e headers, permitindo mapeamento posterior em `adapter` ou `infra`.
- O contrato não deve depender de erros específicos de `net/http`, `grpc` ou outra biblioteca.

### Injeção
- O cliente será injetado em serviços e use cases por meio de construtor.
- Exemplos de uso: `NewPixService(client ports.Client)` ou `NewIngressClient(client ports.Client, cfg config.ExternalServiceConfig)`.

### Regras
- O `core` não deve importar bibliotecas de transporte.
- O `adapter` deve traduzir as entradas do usuário para o formato esperado pelo cliente, mas não implementar a comunicação diretamente.
- O `infra` deve ser a única camada responsável pela comunicação real com o mundo externo.

## Consequências
- O domínio fica mais limpo e testável.
- Alterar de HTTP para gRPC ou trocar bibliotecas se torna uma mudança localizada em `infra`.
- O tratamento de erro fica padronizado e mais simples de mapear em respostas da API.

## Referências
- [gRPC](https://grpc.io/docs/languages/go/)
- [net/http](https://pkg.go.dev/net/http)