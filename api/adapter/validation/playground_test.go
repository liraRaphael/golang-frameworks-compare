package validation

import (
	"context"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/stretchr/testify/assert"
)

type validateTestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"email"`
}

func TestPlaygroundValidator_Validate(t *testing.T) {
	v := NewPlaygroundValidator()
	ctx := ports.NewContext(context.Background())

	t.Run("valid struct", func(t *testing.T) {
		input := validateTestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		err := v.Validate(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("invalid struct", func(t *testing.T) {
		input := validateTestStruct{
			Name:  "",
			Email: "invalid-email",
		}
		err := v.Validate(ctx, input)
		assert.Error(t, err)

		validationErr, ok := err.(*errors.FieldValidationError)
		assert.True(t, ok)
		assert.Len(t, validationErr.Errors, 2)

		assert.Equal(t, "Name", validationErr.Errors[0].Field)
		assert.Equal(t, "required", validationErr.Errors[0].Tag)

		assert.Equal(t, "Email", validationErr.Errors[1].Field)
		assert.Equal(t, "email", validationErr.Errors[1].Tag)
	})

	t.Run("non-struct input", func(t *testing.T) {
		// go-playground validator returns a non-validation error for primitive types
		err := v.Validate(ctx, "just a string")
		assert.Error(t, err)
		_, ok := err.(*errors.FieldValidationError)
		assert.False(t, ok) // should be a standard validator error, not FieldValidationError
	})
}
