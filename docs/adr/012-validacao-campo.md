# ADR-012: Mecanismo de Validação de Dados

**Status**: Aceito

**Data**: 2026-06-26

**Tags**: [validation, validator, core, ports, infra, clean-arch]

## Pré-requisito

* [ADR-001: Estrutura da aplicação](./001-estrutura-app.md)
* [ADR-001a: Detalhamento da camada Core](./001a-camada-core.md)
* [ADR-001b: Detalhamento da camada Adapter](./001b-camada-adapter.md)

## Contexto

Para garantir a integridade dos dados que entram na aplicação e permitir uma comparação justa entre os diferentes frameworks (Gin, Fiber, Echo, etc.), precisamos de um mecanismo de validação centralizado e uniforme.

Contudo, acoplar as estruturas de dados ou os casos de uso diretamente a uma biblioteca de validação específica viola as regras da Clean Architecture estabelecidas na [ADR-001](./001-estrutura-app.md). Cada framework possui sua própria engine interna ou convenção de validação, mas para fins de consistência e isolamento do domínio, a validação de payloads de entrada deve ser agnóstica e componentizada.

## Decisão

Será criado um contrato genérico de validação na camada `core/ports` e uma implementação concreta baseada na biblioteca `go-playground/validator/v10` na camada de infraestrutura.

### Contrato (Interface)

A interface de validação será definida em `api/core/ports/validator.go` (ou pacote equivalente dentro de `core/ports`) com a assinatura genérica solicitada:

```go
package ports

import "context"

// Validator define o contrato genérico para validação de dados na aplicação.
type Validator interface {
	Validate(ctx context.Context, input any) (error)
}

```

*Nota sobre o retorno `error`:* O retorno de erro deverá ser criado em `api\core\domain\errors` de nome `FieldValidationError` com informações relevantes tais como: tag, nome que campo, pacote completo (veja o restante em https://pkg.go.dev/github.com/go-playground/validator/v10#FieldError).

### Implementação Concreta

A implementação concreta utilizará o `go-playground/validator` e residirá em `api/adapter/validation/playground.go`:

```go
package validation

import (
	"context"
	"github.com/go-playground/validator/v10"
)

type PlaygroundValidator struct {
	validate *validator.Validate
}

func NewPlaygroundValidator() ports.Validator {
	return &PlaygroundValidator{
		validate: validator.New(),
	}
}

func (pv *PlaygroundValidator) Validate(ctx context.Context, input any) (error) {
	err := pv.validate.StructCtx(ctx, input)
	if err != nil {
		// Retorna o erro original ou uma estrutura mapeada de erros de validação (ex: erros.FieldValidationError)
		return fieldValidationError
	}
	return nil
}

```

### Regras de Uso e Acoplamento

* **Camada Core**: As structs de Request em `core/domain/requests` podem conter tags de validação (ex: `validate:"required"`), visto que representam restrições do próprio caso de uso, mas o `core` **nunca** importará ou executará a biblioteca de validação diretamente.
* **Camada Adapter**: Os controllers receberão a interface `ports.Validator` por injeção de dependência. Após o binding do JSON/Payload específico do framework, o controller invocará o método `Validate`.
* **Tratamento de Erros**: Se o retorno do validador indicar falhas, o `adapter/handler` mapeará esses erros para o formato de resposta padronizado da API (conforme [ADR-003](https://www.google.com/search?q=003-handler.md)).

## Consequências

* **Isolamento Total**: O domínio e os casos de uso não conhecem o motor de validação.
* **Uniformidade no Benchmark**: Todos os frameworks comparados utilizarão exatamente a mesma engine de validação, removendo discrepâncias de performance que surgiriam se usássemos o validador nativo de cada um deles (ex: o validador embutido do Echo vs. o do Gin).
* **Substituibilidade**: Caso seja necessário trocar para uma biblioteca mais performática (como `go-andiamo/validator` ou geração de código estático), a mudança impactará apenas o bootstrap da aplicação e a camada `infra`.