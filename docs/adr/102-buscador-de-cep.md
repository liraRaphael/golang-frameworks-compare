# ADR-102: Cenário básico — Busca de CEP

**Status**: Aceito — implementação parcial; conversão tipada e tratamento de CEP inexistente pendentes

**Data**: 2026-10-25

**Tags**: [benchmark, cep, viacep, validacao, roteamento, service, client]

## Pré-requisito
- [ADR-001: Estrutura da aplicação](001-estrutura-app.md)
- [ADR-001a: Detalhamento da camada Core](001a-camada-core.md)
- [ADR-001b: Detalhamento da camada App](001b-camada-adapter.md)
- [ADR-001c: Detalhamento da camada Infrastructure](001c-camada-infraestrutura.md)
- [ADR-002: Cliente](002-cliente.md)
- [ADR-003: Handler](003-handler.md)
- [ADR-004: Roteamento](004-roteamento.md)
- [ADR-005: Observability](005-observability.md)
- [ADR-006: Estratégia de benchmarking e testes](006-benchmarking-e-testes.md)
- [ADR-007: Geração automática de documentação OpenAPI/Swagger](007-openapi-swagger.md)
- [ADR-009: Padrões de Request, Response e HttpParam no Core](009-request-response-core.md)
- [ADR-010: Injeção de Dependência (DI)](010-injecao-dependencia.md)
- [ADR-011: Gestão de Configuração](011-gestao-configuracao.md)
- [ADR-012: Mecanismo de Validação de Dados](012-validacao-campo.md)
- [ADR-013: Implementação dos frameworks web comparados](013-implementacao-frameworks-web.md)

## Contexto
Este ADR define o cenário de benchmark de Busca de CEP listado no [README](../../readme.md), o comportamento esperado e o estado conhecido da implementação em `api`.

O cenário mede roteamento e extração de parâmetro de rota, validação, chamada HTTP ao ViaCEP e serialização da resposta. A aplicação não usará cache: cada consulta válida deve chegar ao serviço externo, tornando a latência e a disponibilidade do ViaCEP parte do resultado medido.

As ADRs de arquitetura funcionam como referência de projeto e não como prova de implementação. No estado atual:
- `api/core/ports/service.go` declara a interface `CepService`, e `api/app/service/cep.go` implementa a consulta via `ports.Client`.
- `api/core/domain/requests/cep.go`, `api/core/domain/responses/cep.go` e `api/app/controller/cep.go` definem o parâmetro, o endereço e o controller do fluxo.
- `api/main.go` compõe o cliente HTTP com timeout de cinco segundos, o serviço, o controller e os mapeamentos de erros; registra a rota `/cep/{cep}`.
- O serviço valida o CEP, consulta o ViaCEP e verifica o status HTTP. No entanto, tenta converter o corpo com `res.Body().(responses.CepAddress)`. As implementações HTTP retornam um body decodificado em `any`, normalmente um `map[string]interface{}`, portanto essa asserção não é confiável e pode falhar sem retornar erro.
- O serviço não trata `res == nil` quando o cliente retorna `nil, nil`, o que pode causar panic.
- `responses.CepAddress` não contém o campo `erro` do ViaCEP; a resposta HTTP 200 de CEP inexistente ainda não é convertida em `CEPNotFoundError`, apesar de o tipo de erro e o mapeamento HTTP 404 existirem.
- Ainda não há testes unitários para o serviço de CEP.
- Clientes HTTP devolvem respostas não-2xx como response, sem convertê-las automaticamente em erro (ADR-002).
- Erros de CEP inválido, não encontrado e upstream estão registrados em `api/main.go`; os demais erros sem mapeamento explícito seguem o fallback HTTP 500 (ADR-003).
- Os adapters não garantem equivalência total de extração de path params, binding ou escrita da resposta; `Use` ainda não aplica os middlewares previstos (ADR-013).
- A infraestrutura de benchmarks e a instrumentação de métricas previstas na ADR-006 ainda não estão integradas ao projeto.

## Decisão
O cenário consulta o ViaCEP diretamente, sem cache. A chamada ao serviço fica na camada `api/app/service`: controllers não chamam diretamente o cliente HTTP nem a API externa. A interface declarada em `api/core/ports/service.go` é injetada no controller.

### Fonte de dados
- A origem é `GET https://viacep.com.br/ws/{cep}/json/`.
- O CEP enviado à origem contém oito dígitos ASCII, sem hífen ou pontuação. Tratar o identificador como string para preservar zeros à esquerda.
- Consultas válidas devem chamar o ViaCEP; não armazenar nem reutilizar respostas entre requisições.
- Consultas inválidas devem ser rejeitadas antes de qualquer chamada externa.
- Não interpretar como endereço uma resposta de CEP inexistente (`{"erro": true}`), uma resposta HTTP malsucedida ou um payload inválido.
- Usar a mesma lista e distribuição de CEPs em todos os frameworks. Registrar os CEPs usados e o período da execução, pois os dados e a disponibilidade da fonte podem mudar.

Exemplo representativo de resposta do ViaCEP:

```json
{
  "cep": "01001-000",
  "logradouro": "Praça da Sé",
  "complemento": "lado ímpar",
  "unidade": "",
  "bairro": "Sé",
  "localidade": "São Paulo",
  "uf": "SP",
  "estado": "São Paulo",
  "regiao": "Sudeste",
  "ibge": "3550308",
  "gia": "1004",
  "ddd": "11",
  "siafi": "7107"
}
```

### Responsabilidade e contrato do serviço
O contrato é declarado em `api/core/ports/service.go`, sem dependência de framework ou cliente HTTP. A implementação fica em `api/app/service` e é responsável por consultar a origem externa via `ports.Client` e interpretar sua resposta.

Contrato conceitual declarado em `api/core/ports/service.go`:

```go
package ports

import "github.com/liraraphael/go-framework-bench/api/core/domain/responses"

type CepService interface {
    Find(ctx Context, cep string) (responses.CepAddress, error)
}
```

`responses.CepAddress` representa o endereço encontrado e não deve representar um CEP inexistente como endereço vazio. A interface e o tipo já existem nos caminhos descritos acima. O serviço deve converter o payload externo para o tipo de endereço com erro explícito, sem type assertion insegura; deve também detectar `erro: true` do ViaCEP e retornar `CEPNotFoundError`.

Responsabilidades:
- **Controller (`api/app/controller`)**: receber o CEP já convertido e validado, chamar `CepService.Find` e criar `responses.Response` tipada com status apropriado para resultados conhecidos. Não deve validar DTOs de transporte, serializar HTTP/JSON, nem conhecer ViaCEP, URL externa ou detalhes de `ports.Client`.
- **Porta (`api/core/ports/service.go`)**: declarar o contrato consumido pelo controller.
- **Serviço (`api/app/service`)**: implementar `CepService`; chamar `ports.Client` em cada consulta válida, verificar status e conteúdo da resposta e converter o payload remoto para o tipo de endereço.
- **Cliente (`api/infra/clients`)**: executar a comunicação HTTP real por meio de `ports.Client`; a camada `app` não deve importar `net/http`, `fasthttp` ou SDK de framework.
- **Composição (`api/main.go`)**: construir cliente, serviço e controller e injetar dependências manualmente, conforme ADR-010.

O serviço recebe `ports.Client` por construtor. Para testes determinísticos, injetar um client fake; testes de integração podem usar um servidor HTTP local. Não adicionar cache local, distribuído ou em memória, nem cachear respostas no serviço.

O serviço deve configurar um timeout finito para chamadas externas. O mesmo timeout e a mesma configuração devem ser usados em todos os frameworks. Status HTTP, erros de transporte e erros de decodificação devem ser tratados explicitamente; não interpretar response não-2xx como sucesso apenas porque `ports.Client.Do` não retornou erro.

### Contrato HTTP e validação
- Método e rota: `GET /cep/{cep}`.
- O CEP é exclusivamente um parâmetro de path; não aceitar CEP por query ou body.
- Validar o formato `^[0-9]{8}$` antes da consulta ao serviço. Não remover hífen, espaços ou outros caracteres automaticamente.
- CEP deve ser representado como string em todo o fluxo.
- CEP malformado: HTTP 400 e nenhuma consulta ao ViaCEP.
- CEP válido encontrado: HTTP 200 e endereço em JSON com os campos da resposta do ViaCEP.
- CEP válido inexistente: HTTP 404. O campo `erro: true` do ViaCEP deve ser interpretado como ausência de resultado, nunca como endereço válido.
- Timeout, falha de transporte, status externo inesperado ou payload inválido: HTTP 502 para erro de dependência externa. Erros internos inesperados: HTTP 500. Nenhum desses casos pode produzir resposta de sucesso vazia.
- Respostas de erro devem usar `responses.ErrorResponse` com códigos estáveis, por exemplo `invalid_cep`, `cep_not_found`, `upstream_error` e `internal_error`. O mapeamento explícito deve ser registrado ou implementado no fluxo de resposta; não depender do fallback HTTP 500 para erros de entrada ou ausência de recurso.

O DTO dos parâmetros de path declara a obrigatoriedade e o formato do CEP em tags `validate` (por exemplo, `validate:"required,cep"`). A rota tipada converte o path para o DTO e executa `ports.Validator` antes do controller; a regra customizada de oito dígitos ASCII fica em `api/app/validation/validation.go`. O controller não valida nem normaliza o CEP. A existência do CEP e a interpretação de `erro: true` dependem da consulta e pertencem ao serviço.

O contrato de sucesso deve representar os campos efetivamente retornados pelo ViaCEP. Exemplos de tipos conceituais:

```go
type CepAddress struct {
    CEP         string `json:"cep"`
    Logradouro  string `json:"logradouro"`
    Complemento string `json:"complemento"`
    Unidade     string `json:"unidade"`
    Bairro      string `json:"bairro"`
    Localidade  string `json:"localidade"`
    UF          string `json:"uf"`
    Estado      string `json:"estado"`
    Regiao      string `json:"regiao"`
    IBGE        string `json:"ibge"`
    GIA         string `json:"gia"`
    DDD         string `json:"ddd"`
    SIAFI       string `json:"siafi"`
}
```

Os campos numéricos da fonte devem permanecer strings quando fizerem parte do identificador ou puderem conter zeros à esquerda. O formato de sucesso preserva o CEP formatado que a fonte retorna; a entrada permanece no path sem hífen.

### Organização por responsabilidade
Os caminhos abaixo são os contratos e componentes previstos para o fluxo:

#### Core
- `api/core/domain/requests/cep.go`: parâmetros tipados da rota.
- `api/core/domain/responses/cep.go`: `CepAddress` e os tipos de saída do domínio.
- `api/core/ports/service.go`: interface `CepService`, consumida pelo controller.
- `api/core/ports/client.go`: reutilizar o contrato genérico `ports.Client`; não criar chamada HTTP de ViaCEP diretamente no domínio.

#### App
- `api/app/controller/cep.go`: receber os path params convertidos/validados, chamar `CepService` e construir `responses.Response` tipada, definindo o status HTTP para resultados conhecidos.
- `api/app/service/cep.go`: implementar o serviço de CEP, incluindo chamada ao client e interpretação do resultado externo; retornar endereço tipado e `error`, sem construir response HTTP, definir status ou serializar. A implementação atual ainda deve substituir a type assertion por conversão tipada explícita e tratar resposta nula.
- `api/app/validation/validation.go`: registrar a regra customizada para oito dígitos ASCII; as tags `validate` ficam no DTO de path em `core/domain/requests`.
- `api/app/handler`: mapear erros de validação, não encontrado, dependência externa e falha interna para status/payload correspondentes.

#### Infra e composição
- `api/infra/clients`: reutilizar uma implementação HTTP existente para ViaCEP, configurada com timeout; não introduzir chamada externa dentro dos adapters de framework.
- `api/infra/router` e adapters: registrar `GET /cep/{cep}` com os mesmos metadados em cada framework.
- `api/main.go`: compor manualmente client, serviço, validator e controller; a composição e a rota já estão registradas.
- Registrar metadados OpenAPI de summary, description, tags, parâmetros de path e resposta de sucesso. Documentar separadamente os erros, pois o registry atual não registra automaticamente o schema de erros (ADR-007).

### Fluxo esperado
1. O adapter recebe `GET /cep/{cep}` e encaminha o parâmetro à rota tipada.
2. A rota tipada converte o path para o DTO e valida as tags `validate` pelo `ports.Validator`; CEP malformado resulta em HTTP 400 sem chamar controller ou ViaCEP.
3. O controller chama `CepService.Find`; o serviço consulta ViaCEP via `ports.Client`, converte o payload para tipo conhecido e interpreta `erro: true` como `CEPNotFoundError`, sem construir response HTTP.
4. O controller cria `responses.Response` com status 200, 404 ou 502 para resultados conhecidos; erros inesperados seguem para o handler e resultam em HTTP 500.
5. O adapter serializa e escreve a response do controller ou a response de erro resolvida pelo handler. Controller e serviço não serializam nem escrevem HTTP/JSON.

### Testes e comparação
Antes de considerar o cenário pronto para benchmark, cobrir:
- chamada ao ViaCEP para cada consulta válida;
- preservação de zeros à esquerda;
- CEP malformado sem chamada ao client;
- CEP válido inexistente (`erro: true`);
- status externo não-2xx, timeout/erro de transporte, JSON malformado e payload incompleto;
- chamadas simultâneas para CEPs iguais e diferentes;
- status, payload e headers do endpoint em cada adapter.

Testes unitários do serviço devem injetar um client fake, sem acesso real à rede. Testes de integração do serviço podem usar servidor HTTP de teste; não depender do ViaCEP ao executar a suíte regular. Testes HTTP devem confirmar a extração do parâmetro de path em cada adapter, pois o contrato comum não garante equivalência comportamental.

Para o benchmark:
- Configurar `N` CEPs válidos distintos e distribuir as consultas da mesma forma em todos os frameworks.
- Não incluir aquecimento ou pré-carregamento da aplicação com endereços, pois não há cache.
- Separar respostas 400/404/502 e falhas de dependência; não contar resposta de erro como sucesso.
- Reportar latência, throughput, status de resposta e configuração/limites do ambiente conforme ADR-006.
- Como a chamada externa está dentro da janela medida, registrar a latência do ViaCEP e executar todos os frameworks em condições tão próximas quanto possível. Executar rodadas comparativas em períodos próximos e reportar indisponibilidades ou variações da dependência. Resultados medem o sistema completo (framework, aplicação, rede e ViaCEP), não isoladamente o custo do framework.
- Se o objetivo for comparar apenas o custo dos frameworks, usar um servidor stub controlado em experimento separado, mantendo o mesmo contrato e latência configurada para todas as implementações; não introduzir cache como substituto.

## Consequências
- A separação explícita entre controller, porta, serviço de aplicação e client evita chamadas externas acopladas a framework ou controller.
- Sem cache, cada consulta válida depende da disponibilidade, latência e resposta do ViaCEP.
- A distinção entre CEP inválido, inexistente e falha externa define status e permite interpretar os resultados sem falsos sucessos.
- O benchmark com ViaCEP mede o caminho completo e não isola o overhead do framework; para isso, é necessário experimento separado com dependência simulada e controlada.
- A estrutura principal (serviço, interface, controller, composição e mapeamentos básicos de erro) já existe, mas o cenário não está pronto para benchmark até que a conversão tipada seja confiável, CEP inexistente seja detectado e os testes previstos sejam implementados.

## Referências
1. [README — cenários de benchmark](../../readme.md)
2. [ADR-002: Cliente](002-cliente.md)
3. [ADR-003: Handler](003-handler.md)
4. [ADR-004: Roteamento](004-roteamento.md)
5. [ADR-006: Estratégia de benchmarking e testes](006-benchmarking-e-testes.md)
6. [ADR-007: Geração automática de documentação OpenAPI/Swagger](007-openapi-swagger.md)
7. [ADR-009: Padrões de Request, Response e HttpParam no Core](009-request-response-core.md)
8. [ADR-010: Injeção de Dependência (DI)](010-injecao-dependencia.md)
9. [ADR-011: Gestão de Configuração](011-gestao-configuracao.md)
10. [ADR-012: Mecanismo de Validação de Dados](012-validacao-campo.md)
11. [ADR-013: Implementação dos frameworks web comparados](013-implementacao-frameworks-web.md)
12. [ViaCEP — Webservice gratuito para consulta de CEP](https://viacep.com.br/)
