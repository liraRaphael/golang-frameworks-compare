# ADR-012: Mecanismo de Validação de Dados

**Status**: Aceito

**Data**: 2026-06-26

**Tags**: [validation, validator, core, ports, app]

## Pré-requisito

* [ADR-001: Estrutura da aplicação](./001-estrutura-app.md)
* [ADR-001a: Detalhamento da camada Core](./001a-camada-core.md)
* [ADR-001b: Detalhamento da camada App](./001b-camada-adapter.md)

## Contexto

Para garantir a integridade dos dados que entram na aplicação e permitir uma comparação justa entre os diferentes frameworks (Gin, Fiber, Echo, etc.), precisamos de um mecanismo de validação centralizado e uniforme.

Contudo, acoplar as estruturas de dados ou os casos de uso diretamente a uma biblioteca de validação específica viola as regras da Clean Architecture estabelecidas na [ADR-001](./001-estrutura-app.md). Cada framework possui sua própria engine interna ou convenção de validação, mas para fins de consistência e isolamento do domínio, a validação de payloads de entrada deve ser agnóstica e componentizada.

## Decisão

O contrato `ports.Validator` está em `api/core/ports/validator.go`; a implementação concreta `PlaygroundValidator` está em `api/app/validation/playground.go` e usa `go-playground/validator/v10`.
Regras de validação customizadas compartilhadas ficam em `api/app/validation/validation.go` e são registradas pelo `PlaygroundValidator`; os campos declaram essas regras por tags no DTO. A regra `finite_conversion` acessa os demais campos do DTO para impedir resultado aritmético não finito antes do use case.

No cenário de conversão de moedas, `From` e `To` declaram `validate:"required,oneof=USD BRL"`; `Value` declara `required`, `finite`, `gte=0` e `finite_conversion`. A regra customizada `finite_conversion` usa `CurrencyConversionRequest.CurrencyQuote()` para verificar o resultado calculado e arredondado. As regras customizadas e a biblioteca validator permanecem na camada `app`; o request no Core contém apenas tags e operações puras do domínio.

### Contrato (Interface)

O contrato atual recebe `ports.Context` e devolve um `error`:

```go
package ports

type Validator interface {
	Validate(ctx Context, input any) error
}

```

`core/domain/errors.FieldValidationError` contém uma lista de `FieldError` com `Field`, `Tag`, `Value`, `Param` e `Namespace`. `PlaygroundValidator` converte `validator.ValidationErrors` para esse tipo; outros erros são devolvidos sem transformação.

### Implementação Concreta

A implementação concreta utiliza `go-playground/validator` e reside em `api/app/validation/playground.go`:

```go
package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type PlaygroundValidator struct {
	validate *validator.Validate
}

func NewPlaygroundValidator() ports.Validator {
	validate := validator.New()
	if err := registerValidationRules(validate); err != nil {
		panic(err)
	}
	return &PlaygroundValidator{
		validate: validate,
	}
}

func (pv *PlaygroundValidator) Validate(ctx ports.Context, input any) error {
	err := pv.validate.StructCtx(ctx, input)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fieldErrors := make([]errors.FieldError, 0, len(validationErrors))
			for _, fe := range validationErrors {
				fieldErrors = append(fieldErrors, errors.FieldError{
					Field: fe.Field(), Tag: fe.Tag(), Value: fe.Value(),
					Param: fe.Param(), Namespace: fe.Namespace(),
				})
			}
			return &errors.FieldValidationError{Errors: fieldErrors}
		}
		return err
	}
	return nil
}

```

### Regras de Uso e Acoplamento

* **Camada Core**: As structs de Request em `core/domain/requests` podem conter tags de validação (ex: `validate:"required"`), visto que representam restrições do próprio caso de uso, mas o `core` **nunca** importará ou executará a biblioteca de validação diretamente.
* **Framework/roteamento tipado**: o adapter faz o binding do JSON e verifica o content type; `api/infra/router` converte o body para o DTO da rota e executa o `ports.Validator` antes de chamar o controller.
* **Controller**: recebe o request já convertido e validado, executa/orquestra o caso de uso, define o status e cria `responses.Response` tipada. Não deve repetir validação do DTO nem serializar HTTP/JSON.
* **Use case**: recebe o request validado e retorna apenas o corpo tipado da resposta e `error`; não cria `responses.Response`, não define status HTTP e não serializa.
* **Tratamento de Erros**: o erro é devolvido ao adapter e `FieldValidationError` é mapeado pelo `standardHandler` para HTTP 400; detalhes internos do motor de validação não são enviados ao cliente.

## Consequências

* **Isolamento do motor**: O contrato impede que controllers dependam diretamente da engine, embora a implementação viva em `app`.
* **Uniformidade entre adapters**: os adapters encaminham o payload à mesma validação compartilhada na rota tipada antes de executar o controller; o controller constrói a response tipada e o adapter serializa/escreve a resposta. Rotas sem body tipado, como `/health`, não executam validação de DTO.
* **Tratamento atual**: falhas `FieldValidationError` são convertidas em HTTP 400 com `validation_error` pelo handler padrão; outros erros sem mapper continuam usando o fallback HTTP 500.