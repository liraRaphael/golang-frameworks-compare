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

O contrato `ports.Validator` está em `api/core/ports/handler.go`; a implementação concreta `PlaygroundValidator` está em `api/app/validation/playground.go` e usa `go-playground/validator/v10`.

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
	return &PlaygroundValidator{
		validate: validator.New(),
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
* **App**: `HelloController` recebe o validator por construtor e valida o request antes de invocar o use case. O framework faz o binding do JSON antes de chamar o controller.
* **Tratamento de Erros**: O erro é devolvido ao adapter; sem registro de tipo/mensagem no `standardHandler`, o fallback atual gera resposta 500, não detalhes de validação.

## Consequências

* **Isolamento do motor**: O contrato impede que controllers dependam diretamente da engine, embora a implementação viva em `app`.
* **Uniformidade entre adapters**: Os adapters usam o mesmo controller/engine nos endpoints que validam; atualmente `/hello` utiliza validação, enquanto `/health` não valida payload.
* **Tratamento atual**: As falhas de validação não têm um mapper registrado no bootstrap e são devolvidas pelo fallback 500 do handler padrão.