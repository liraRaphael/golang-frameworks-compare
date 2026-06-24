# ADR-002: Cliente

**Status**: Aceito

**Data**: 2026-06-22

**Tags**: [API, cliente, services, http, grpc]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
Para que o projeto seja capaz de se comunicar com serviços externos, é necessário definir um cliente que possa ser utilizado para realizar chamadas HTTP ou gRPC. Este cliente deve ser configurável, permitindo a definição de parâmetros como timeout, headers e autenticação.
A ideia é que seja implementado uma *classe abstrata* `Client` que defina os métodos necessários para realizar requisições, e implementações concretas que utilizem a bibliotecas como o padrão do Go (`net/http`), a `fasthttp` ou e a `github.com/grpc/grpc-go` (para testes com [*gRPC*](https://grpc.io/docs/languages/go/)), de tal forma que a lógica de comunicação seja isolada e reutilizável.

## Decisão
- Dada a estruturação do projeto definida na ADR-001, a decisão é criar um pacote `client` dentro da camada `infrastructure`, que conterá a interface `Client` e suas implementações concretas. A interface `Client` deve definir métodos para realizar requisições HTTP e gRPC, enquanto as implementações concretas devem encapsular a lógica de comunicação com os serviços externos.
- Para cada biblioteca de comunicação (HTTP, gRPC), será criada uma implementação específica do `Client`, permitindo que a aplicação utilize diferentes protocolos de comunicação conforme necessário. Além disso, o cliente deve ser configurável, permitindo a definição de parâmetros como timeout, headers, body e etc.
- Para facilitar a integração com os casos de uso do `core`, o cliente deve ser projetado para receber e retornar estruturas de dados definidas no `core`, garantindo que a lógica de negócio permaneça independente da implementação do cliente, além disso o *client* deverá receber o *client* correspondente como parâmetro de inicialização, por exemplo: `NewPixClient(client *client.Client)` ou `NewIngressClient(client *client.Client, param1 string, param2 any)`.
- Os erros de comunicação deverão ser encapsulados em tipos de erro específicos, que implementem a interface `error` (padrão do Go) sendo criado na camada `core/domain/errors` (definido na *ADR-001*) do Go e carreguem informações necessárias para tratamento e mapeamento para respostas **INDEPENDENTE DA BIBLIOTECA DE COMUNICAÇÃO**.
- Os erros gerados por *Status Code* diferentes de 2xx deverão ser tratados e encapsulados em tipos de erro específicos, que implementem a interface `error` (padrão do Go) sendo criado na camada `core/domain/errors` e carreguem informações necessárias para tratamento e mapeamento para respostas HTTP/grpc, tais como *Status Code Error*, *Body Response*, *Headers Response* e etc.
- A interface `Client` deve ser projetada para permitir a adição de novos métodos de comunicação no futuro, garantindo que a aplicação possa evoluir sem quebrar a compatibilidade com implementações existentes.
  - Deverá ser criado um método genérico `Do` que receba como parâmetro uma estrutura de requisição (`context.Context`, `endpoint string`, `httpMethod enums.HttpMethod`, `*Request`) e retorne uma estrutura de resposta (`*Response`, `error`), permitindo que diferentes tipos de requisições sejam tratados de forma uniforme.
  - Deverá ser criado um Enum `HttpMethod` que contenha os métodos HTTP suportados (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD) e que seja utilizado como parâmetro no método `Do`, garantindo que apenas métodos válidos sejam utilizados que ficará na camada `core/domain/enums`.