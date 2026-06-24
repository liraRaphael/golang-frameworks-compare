# ADR-001: Estrutura da aplicação

**Status**: Aceito

**Data**: 2026-06-22

**Tags**: [app, clean-arch, arquitetura]

## Pré-requisitos
- Ter o [Golang](https://go.dev/dl/) instalado
- Conhecimento básico de Clean Architecture

## Contexto
Para que haja um padrão de estruturação da aplicação, é necessário definir uma arquitetura que seja escalável, modular e de fácil manutenção. A *Clean Architecture* é uma abordagem que promove a separação de responsabilidades e a independência das camadas, permitindo que a lógica de negócio seja isolada das dependências externas.

*Clean Architecture* Segue a linha natural de cima para baixo ou de fora para dentro:

```mermaid
Usuário → API → Application → Domain → Infrastructure
```
Aponta estritamente para dentro. A camada de Application define apenas as interfaces/contratos (ex: IPedidoRepository), e a camada de Infrastructure é quem precisa conhecer a aplicação para implementar e injetar essa interface. \[[1](https://dev.to/yuripeixinho/clean-architecture-arquitetura-limpa-33e1)\]

Aqui está a tabela de matriz de dependência cruzada entre as três camadas, destacando quem pode conhecer quem, o nível de estabilidade e o papel de cada uma no fluxo de código.

| Camada Origem (Quem desenvolve) | ...Pode depender de **Core**? | ...Pode depender de **Adapter**? | ...Pode depender de **Infrastructure**? | Nível de Estabilidade | Papel Principal |
| --- | --- | --- | --- | --- | --- |
| **1. Core** | **Sim** *(Auto-contido)* | ❌ **Nunca** | ❌ **Nunca** | **Altíssima** (Raramente muda) | Contém as regras de negócio puras e define as interfaces (contratos). |
| **2. Adapter** | **Sim** | ❌ *Apenas interno* | ❌ **Nunca** | **Média** (Muda se a API/Interface mudar) | Converte e traduz os dados do mundo externo para o formato do Core. |
| **3. Infrastructure** | **Sim** | **Sim** | ❌ *Apenas interno* | **Baixa** (Altamente volátil/Plugável) | Implementa os contratos do Core (banco, envio de e-mail) e configura o framework. |



Assim:

* **O Core é Isolado:** Se você precisar importar uma biblioteca de banco de dados (como Entity Framework, Prisma, Hibernate) ou um framework web (como ASP.NET, Express, Spring Boot) dentro do **Core**, a regra de dependência foi quebrada.
* **Inversão de Dependência ($D$ do SOLID):** Para que a *Infrastructure* consiga conversar com o *Core* sem que o *Core* dependa dela, a *Infrastructure* **implementa as interfaces** que o *Core* criou.
  * Logo, para que um componente do **Adapter** (ou um Caso de Uso do *Core*) funcione utilizando um serviço externo da **Infrastructure** (como um cliente HTTP/SDK de Pix ou Ingressos), esse componente do **Adapter** deve receber a **interface definida pelo Core** como parâmetro de inicialização. A implementação concreta dessa interface (o *Client* real com a biblioteca selecionada) é desenvolvida na **Infrastructure** e injetada no **Adapter** (ou no *Core*) no momento da inicialização da aplicação."
  * O **Adapter** recebe a **interface do Core** em seu construtor (`NewPixService(client core.Client)` por exemplo), permitindo que qualquer *Client* concreto da **Infrastructure** (que implemente essa interface) seja passado para ele na inicialização.
Dessa forma, o código da Infraestrutura aponta para dentro em tempo de compilação, mas em tempo de execução o fluxo corre para fora para buscar os dados.

## Decisão
Esta ADR define a estrutura inicial do projeto baseada em princípios de Clean Architecture. O layout principal é organizado em três camadas: `adapter`, `core` e `infrastructure`. As camadas devem ser independentes entre si e seguir contratos (interfaces) definidos no `core`.

### Princípios gerais
- Apenas a camada `adapter` pode depender de bibliotecas e frameworks externos.
- Todas as funções e métodos devem receber `context.Context` como primeiro parâmetro para permitir cancelamento e propagação de deadlines.
- Recursos de observabilidade (ex.: *pprof*, métricas, tracing) devem ser implementados em `infrastructure`.

### Adapter
A camada `adapter` recebe requisições (HTTP, gRPC, CLI etc.) e as traduz para os casos de uso do `core`. Deve conter adaptadores concretos, como controllers/handlers, presenters, serviços externos e implementações de repositórios.
- Responsabilidades:
	- Adaptar entradas/saídas para o formato do `core`.
	- Implementar interfaces definidas no `core`.
	- Conter dependências de infraestrutura e terceiros (frameworks, clientes HTTP, etc.).

#### Controllers
Recebem o `context.Context` e os dados da requisição (headers, body), validam, transformam para objetos de domínio (`core/domains/requests`) e chamam os use cases do `core`.

#### Presenters
Transformam resultados do `core` em respostas apropriadas para o cliente (JSON, HTML, Protobuf, etc.).

#### Services
Adaptadores para serviços externos (APIs de terceiros, filas, provedores de e-mail), isolando dependências externas.

#### Handler
Camada que ficará responsável por devolver a resposta para o cliente, seja ela de sucesso ou erro, e também por logar as informações necessárias para monitoramento e depuração, além de tratar erros, mapear a mensagem de erro e propagá-los conforme necessário.

#### Validation
Camada responsável por validar os dados de entrada e/ou saída dos casos de uso, garantindo que eles estejam em conformidade com as regras de negócio. Nela é possivel criar validações genéricas que podem ser reutilizadas em diferentes casos de uso, como validação de campos obrigatórios, formatos de dados, limites de tamanho, entre outros.
Sua interface deverá estar em `core/validation` e suas implementações concretas em `adapter/validation`, abstraindo a lib `github.com/go-playground/validator/v10`


#### Repositories (implementações)
Implementações concretas de persistência que obedecem às interfaces declaradas no `core`.

### Core
Contém a lógica de negócio e os contratos (interfaces) que o resto do sistema deve respeitar.

- Estrutura típica: `usecases`, `domains` (com `entities`, `requests`, `responses`, `errors`, `enums`) e pacotes com interfaces (`controller`, `services`, `repositories`).

#### Use cases
Orquestram operações de negócio, coordenando entidades, validações e chamadas a repositórios/serviços.

#### Domains
- `entities`: modelos de negócio persistentes e regras associadas.
- `requests`: estruturas que representam entradas de casos de uso.
  - deverá ter uma struct genérica `Request` que contenha `Body` (generic), `Headers` (`HttpParam`), `QueryParams` (HttpParam) e `PathParams` (HttpParam) para facilitar a adaptação de diferentes tipos de entrada.
  - Deverá ficar as demais estruturas de entrada específicas de cada caso de uso em subpacotes, por exemplo: `domains/requests/user`, `domains/requests/order`, etc.
- `responses`: estruturas que representam saídas de casos de uso.
  - deverá ter uma struct genérica `Response` que contenha `Body` (any), `Headers` (HttpParam) e `Status Code` (int).
- `errors`: tipos de erro do domínio; devem implementar a interface `error` do Go e carregar informações necessárias para tratamento e mapeamento para respostas HTTP/grpc.
- `enums`: constantes, vars e tipos auxiliares para valores enumerados.
  - `HttpParam` deve ter os métodos: 
    - `Get(key string) ([]string, bool)` para recuperar valores de forma segura.
    - `GetFirst(key string) (string, bool)` para recuperar o primeiro valor de forma segura.
    - `GetFirstOrDefault(key string, defaultValue string) string` para recuperar o primeiro valor ou retornar um valor padrão.
    - `GetAll() map[string][]string` para recuperar todos os valores.
    - `Set(key string, value string)` para definir um valor.
    - `SetAll(params map[string][]string)` para definir múltiplos valores.
    - `Delete(key string)` para remover um valor.
    - `GetFirstInt(key string) (int, bool)` para recuperar o primeiro valor como inteiro de forma segura.
      - Caso de erro ao converter para inteiro, deve retornar `false` no segundo valor.
    - `GetFirstIntOrDefault(key string, defaultValue int) int` para recuperar o primeiro valor como inteiro ou retornar um valor padrão.
      - Caso de erro ao converter para inteiro, deve retornar o valor padrão.
    - `Clear()` para limpar todos os valores.

#### Interfaces
Definem os contratos que `adapter` e `infrastructure` devem implementar (controladores, repositórios, clientes, etc.). Mantenha-as pequenas e específicas para facilitar testes e substituições.

### Infrastructure
Camada responsável por integrar o sistema ao "mundo real": banco de dados, clientes externos, configuração, observabilidade, roteamento e documentação.

#### Database
Contém conexões, migrations, modelos e implementações de repositórios. Repositórios devem expor contratos em `core` e fornecer CRUD conforme necessário.

#### Observability
Implementa métricas, tracing, logging e ferramentas como *pprof*; expõe pontos de instrumentação e integra com provedores (Prometheus, OpenTelemetry).

#### Config
Carrega variáveis de ambiente e arquivos de configuração; centraliza parâmetros de inicialização da aplicação.
Essa deverá ter um arquivo *go* por recurso que exige configurações, por exemplo: `config/database.go`, `config/server.go`, `config/pprof.go`, `config/swagger.go`, etc. No entanto, terá um arquivo `config/config.go` que carregue todas as configurações de uma vez em `init()` em funções previamente registradas e salvas numa variavel global, permitindo que a aplicação inicialize todas as configurações de uma vez, sem precisar se preocupar com a ordem de inicialização dos recursos.
As structs de configuração deverão ser definidas em `core/config` e suas implementações concretas em `infrastructure/config`, abstraindo a lib `github.com/spf13/viper`.

#### Routes
Define a composição de rotas e middlewares. Recomendado expor uma `struct` configurável (builder) para registrar rotas e middlewares de forma agnóstica ao framework, permitindo suportar múltiplos adaptadores (Gin, Echo, Fiber, etc.), bem como o tipo do objeto de entrada e saída.

#### Docs
Gerar e servir especificações (ex.: Swagger/OpenAPI) e documentação estática (Redoc) a partir da composição da aplicação.
**OBS:** 
- Disponibilizar documentação de APIs, independente do framework usado, para facilitar integração com clientes e testes.

#### Clients
Implementações de clientes HTTP/gRPC que obedecem aos contratos do `core` e podem ser substituídos por mocks em testes.

#### Frameworks
Espaço para implementações específicas de frameworks em `infrastructure/frameworks/<framework>` (por exemplo: Gin, Fiber, Echo). Cada implementação deve aderir às interfaces usadas por `adapter` e `routes`.

### Observações
- Prefira projetar interfaces para facilitar testes e substituição de implementações.
- Mantenha as dependências direcionadas: 
  - `core` não deve depender de `adapter`.
- Documente convenções de pacotes e exemplos de uso para acelerar a on-boarding.
- Colocar logs, métricas e tracing em pontos estratégicos do código para facilitar monitoramento e depuração.
- Criar testes unitários, garantindo que a lógica de negócio seja testada isoladamente do framework ou infraestrutura.

## Referências
1. [Clean Architecture](https://dev.to/yuripeixinho/clean-architecture-arquitetura-limpa-33e1)
2. [Golang download](https://go.dev/dl/)