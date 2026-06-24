# ADR-003: Handler

**Status**: Aceito

**Data**: 2026-06-24

**Tags**: [handler, clean-arch, response, error, success, mapeamento]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)

## Contexto
Para que haja uma padronização de respostas e erros, é necessário definir um *handler* que seja responsável por devolver a resposta para o cliente, seja ela de sucesso ou erro, e também por logar as informações necessárias para monitoramento e depuração, além de tratar erros, mapear a mensagem de erro e propagá-los conforme necessário.
Nessa camada, o *handler* deve ser capaz de receber os erros gerados pelos *controllers* do `adapter` e mapeá-los para respostas apropriadas, garantindo que a aplicação seja consistente e previsível em suas respostas.

## Decisão
- A decisão é criar um pacote `handler` dentro da camada `adapter`, que será responsável por lidar com as respostas e erros da aplicação. Este pacote deve fornecer funções para mapear erros para respostas HTTP/grpc apropriadas, bem como para logar informações relevantes para monitoramento e depuração.
- Deverá ter uma função genérica `Handler` que receba como parâmetro `ctx context.Context` e `result any` e retorne uma estrutura de resposta (`*Response`), permitindo que diferentes tipos de erros sejam tratados de forma uniforme.
- A mensagem de erro padrão deverá obedecer a seguinte estrutura:
    ```json
    {
    "erro": {
        "codigo": "codigo de erro",
        "mensagem": "mensagem de erro explicando o por que do erro",
        "detalhes": [
        {
            "campo": "qual campo gerou o erro ou qual parte do sistema gerou o erro",
            "causa": "qual a causa do erro, caso seja possível detalhar melhor",
            "valor": any // aqui pode ser qualquer tipo de dado, como string, number, boolean, array ou objeto, dependendo do contexto do erro
        }
        ]
    }
    }
    ```
    - Deverá ser possível registrar os erros nessa camada, permitindo que informações relevantes sejam logadas para monitoramento e depuração.
    - O registro poderá ser feito por "texto" do erro, ou por "tipo" do erro, ou seja, o *handler* deverá ser capaz de identificar o tipo de erro e logar informações relevantes para cada tipo de erro.
    - Usará o desing partner *Strategy* para permitir que diferentes estratégias de mapeamento de erros possam ser implementadas e utilizadas conforme necessário, garantindo flexibilidade e extensibilidade na forma como os erros são tratados e mapeados para respostas apropriadas.
      - Crie um tipo genérico para a função anonima que será responsável por mapear o erro para a resposta apropriada em `core/domain/enums`, que seja `type ErrorFuncType func(contex.Context, error) *Response`.
      - Na camada de *Handler*, deverá ter 2 mapas de funções de mapeamento de erros: 
        1. onde a chave será o tipo do erro e o valor será a função de mapeamento correspondente.
        2. onde a chave será o texto do erro e o valor será a função de mapeamento correspondente.
      - nesse mapeamento, deverá ser possivel devolver a função anonima que será responsável por mapear o erro para a resposta apropriada, ou seja, o *handler* deverá ser capaz de identificar o tipo de erro e chamar a função de mapeamento correspondente, devolvendo o `*Response` de `core` adequado.
      - caso não haja uma função de mapeamento para o tipo de erro, o *handler* deverá devolver uma resposta padrão de erro, com código *"Internal Server Error"* e mensagem "Erro interno do servidor não mapeado".
- caso seja `any` o parâmetro de `result`, ou seja NÃO seja uma implementação de `error`, o *handler* deverá devolver uma resposta padrão de sucesso definida anteriormente na construção da rota
- Deverá ter um método genérico `Handle` que receba como parâmetro uma estrutura de requisição (`context.Context`, `result` (que poderá ser `any` ou `error`) ) e retorne uma estrutura de resposta (`*Response`), permitindo que diferentes tipos de requisições sejam tratados de forma uniforme. 
- *Handler* deverá receber como parâmetro de inicialização o *framework* a ser utilizado, e a partir disso, será necessário que o framework implemente um método dos *frameworks* que será responsável por: 
  - devolver a resposta para o cliente, seja ela de sucesso ou erro, e também por logar as informações necessárias para monitoramento e depuração, além de tratar erros, mapear a mensagem de erro e propagá-los conforme necessário.
  - a assinatura desse método deverá ser `func(ctx context.Context, statusCode int, body any) error`, onde `statusCode` será o código de status HTTP/grpc e `body` será a estrutura de resposta (`*Response`) que será devolvida para o cliente.
