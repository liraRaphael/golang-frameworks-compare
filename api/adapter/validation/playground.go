package validation

import (
	"context"

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

func (pv *PlaygroundValidator) Validate(ctx context.Context, input any) error {
	err := pv.validate.StructCtx(ctx, input)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fieldErrors := make([]errors.FieldError, 0, len(validationErrors))
			for _, fe := range validationErrors {
				fieldErrors = append(fieldErrors, errors.FieldError{
					Field:     fe.Field(),
					Tag:       fe.Tag(),
					Value:     fe.Value(),
					Param:     fe.Param(),
					Namespace: fe.Namespace(),
				})
			}
			return &errors.FieldValidationError{Errors: fieldErrors}
		}
		return err
	}
	return nil
}
