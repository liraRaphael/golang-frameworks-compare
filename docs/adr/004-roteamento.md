# ADR-004: Roteamento

**Status**: Aceito

**Data**: 2026-06-26

**Tags**: [roteamento, router, response, status code, swagger, openapi]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-003: Handler](003-handler.md)

## Contexto
Tendo como base a ideia de testar *N* frameworks web do Go, é necessário definir um roteamento que seja capaz de lidar com diferentes tipos de requisições e respostas, além de permitir a integração com ferramentas de documentação como *Swagger* e *OpenAPI* e *pprof*. O roteamento deve ser configurável, permitindo a definição de parâmetros como métodos HTTP, paths, middlewares e handlers (pré estruturado na *ADR-001*).

## Decisão
Sabendo-se que a ideia dessa ferramenta é o comparativo de frameworks web do Go, a decisão é criar um pacote `router` dentro da camada `adapter`, que será responsável por lidar com o roteamento da aplicação. Este pacote deve fornecer funções para mapear rotas para handlers apropriados, bem como para logar informações relevantes para monitoramento e depuração sem que haja diferença entre os frameworks. Assim, o roteamento deve ser capaz de lidar com diferentes tipos de requisições e respostas, permitindo a integração com ferramentas de documentação como *Swagger* e *OpenAPI* e *pprof*. Além disso, o roteamento deve ser configurável, permitindo a definição de parâmetros como métodos HTTP, paths, middlewares e handlers (pré estruturado na *ADR-001*).
- Fazendo o uso de composição e desing partner *Factory*, será criado um *Router* genérico que receberá como parâmetro o *framework* a ser utilizado, e a partir disso, será possível criar rotas específicas para cada framework, sem que haja diferença entre eles. 
- O *Router* genérico será responsável por ter um método de cada tipo de método HTTP.
  - Deverá ter um método para cada tipo de método HTTP (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD) que receba como parâmetro o *path* da rota e o *handler* correspondente, permitindo que diferentes tipos de requisições sejam tratados de forma uniforme.
  - Deverá ter um método para adicionar *middlewares* que receba como parâmetro o *middleware* correspondente, permitindo que diferentes tipos de requisições sejam tratados de forma uniforme.
  - Deverá ter a opção de devolver um status code padrão para cada rota, sem o padrão *200 OK* com exceção do *POST* que deverá devolver *201 Created*.
  - Ter a opção de *tipo de request e response* usando *generics* (se possível)
  - Poder de ter a opção de adicionar *Swagger* e *OpenAPI* para cada rota, permitindo que diferentes tipos de requisições sejam tratados de forma uniforme.
    - Com isso, cadastrar os `errors` e usar o *handler* correspondente para gerar os exemplos, adicionando automaticamente a documentação *swagger* e *openapi*.
  - Quando possível, use *ponteiros* para não precisar ficar copiando structs de request e response, mas apenas referenciando-as.
- Deverá ser criado com o partner *Builder* um *RouterBuilder* que será responsável por construir o *Router* genérico, recebendo como parâmetro o *framework* a ser utilizado, e a partir disso, será possível criar rotas específicas para cada framework, sem que haja diferença entre eles.
  - É necessário criar uma interface em `infrastructure/framework` que defina os métodos necessários para criar rotas específicas para cada framework, permitindo que a aplicação utilize diferentes frameworks conforme necessário, fazendo um de para entre eles, além de um *start* para inicializar o framework.
  - Cada *framework* deverá ter sua própria implementação dessa interface, encapsulando a lógica de criação de rotas específicas para cada framework, permitindo que a aplicação utilize diferentes frameworks conforme necessário.
  - Cada *framework* deverá ter as suas proprias configurações em `infrastructure/config`, encapsulando a lógica de configuração específica para cada framework, permitindo que a aplicação utilize diferentes frameworks conforme necessário.
- Crie o *pprof* em um *server* separado (e nativo), para não atrapalhar o fluxo principal da aplicação, e que seja possível habilitar ou desabilitar o *pprof* conforme necessário.
  - por padrão na porta 8081, mas que seja possível alterar a porta conforme necessário (config).
  - o *pprof* deverá ser configurado para permitir a análise de desempenho da aplicação, permitindo que diferentes tipos de requisições sejam tratados de forma uniforme.
  - deverá ficar no pacote `infrastructure/pprof`, encapsulando a lógica de configuração específica para o *pprof*, permitindo que a aplicação utilize diferentes frameworks conforme necessário.
- Crie a rota de *swagger* e *openapi* em um *server* separado (e nativo), para não atrapalhar o fluxo principal da aplicação, e que seja possível habilitar ou desabilitar o *swagger* e *openapi* conforme necessário.
  - por padrão na porta 8082, mas que seja possível alterar a porta conforme necessário (config).
  - as rotas de *swagger* e *openapi* deverão ser geradas automaticamente a partir das rotas cadastradas no *Router*, permitindo que a documentação esteja sempre atualizada com as rotas disponíveis na aplicação.
  - deverá ficar no pacote `infrastructure/docs/swagger`, encapsulando a lógica de configuração específica para o *swagger* e *openapi*, permitindo que a aplicação utilize diferentes frameworks conforme necessário.
  - use a lib `github.com/swaggest/openapi-go` para gerar a documentação *swagger* e *openapi* automaticamente a partir das rotas cadastradas no *Router*, permitindo que a documentação esteja sempre atualizada com as rotas disponíveis na aplicação.
    - Use as tags de field como as descritas em: 
      - [These tags can be used](https://github.com/swaggest/jsonschema-go#field-tags):
        * [`title`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.6.1), string
        * [`description`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.6.1), string
        * [`default`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.6.2), can be scalar or JSON value
        * [`example`](https://json-schema.org/draft/2020-12/json-schema-validation.html#name-examples), a scalar value that matches type of parent property, for an array it is applied to items
        * [`examples`](https://json-schema.org/draft/2020-12/json-schema-validation.html#name-examples), a JSON array value
        * [`const`](https://json-schema.org/draft/2020-12/json-schema-validation.html#rfc.section.6.1.3), can be scalar or JSON value
        * [`deprecated`](https://json-schema.org/draft/2020-12/json-schema-validation#name-deprecated), boolean
        * [`readOnly`](https://json-schema.org/draft/2020-12/json-schema-validation#name-deprecated), boolean
        * [`writeOnly`](https://json-schema.org/draft/2020-12/json-schema-validation#name-deprecated), boolean
        * [`pattern`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.2.3), string
        * [`format`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.7), string
        * [`multipleOf`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.1.1), float > 0
        * [`maximum`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.1.2), float
        * [`minimum`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.1.3), float
        * [`maxLength`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.2.1), integer
        * [`minLength`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.2.2), integer
        * [`maxItems`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.3.2), integer
        * [`minItems`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.3.3), integer
        * [`maxProperties`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.4.1), integer
        * [`minProperties`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.4.2), integer
        * [`exclusiveMaximum`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.1.2), boolean
        * [`exclusiveMinimum`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.1.3), boolean
        * [`uniqueItems`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.3.4), boolean
        * [`enum`](https://json-schema.org/draft-04/json-schema-validation.html#rfc.section.5.5.1), tag value must be a JSON or comma-separated list of strings
        * `required`, boolean, marks property as required
        * `nullable`, boolean, overrides nullability of the property


## Referências
1. https://pkg.go.dev/github.com/swaggest/openapi-go#section-readme
