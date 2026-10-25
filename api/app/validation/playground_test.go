package validation

import (
	"math"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/infra/context"

	"github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validateTestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"email"`
}

func TestPlaygroundValidator_ValidatesCurrencyFields(t *testing.T) {
	validator := NewPlaygroundValidator()
	ctx := context.Background()
	validValue := 10.0
	zero := 0.0
	negative := -1.0
	nan := math.NaN()
	infinity := math.Inf(1)
	negativeInfinity := math.Inf(-1)
	tests := []struct {
		name      string
		request   requests.CurrencyConversionRequest
		field     string
		tag       string
		wantValid bool
	}{
		{
			name:      "valid request",
			request:   requests.CurrencyConversionRequest{From: "USD", To: "BRL", Value: &validValue},
			wantValid: true,
		},
		{
			name:      "zero value is valid",
			request:   requests.CurrencyConversionRequest{From: "USD", To: "BRL", Value: &zero},
			wantValid: true,
		},
		{
			name:    "missing source currency",
			request: requests.CurrencyConversionRequest{To: "BRL", Value: &validValue},
			field:   "From",
			tag:     "required",
		},
		{
			name:    "unsupported source currency",
			request: requests.CurrencyConversionRequest{From: "EUR", To: "BRL", Value: &validValue},
			field:   "From",
			tag:     "oneof",
		},
		{
			name:    "missing target currency",
			request: requests.CurrencyConversionRequest{From: "USD", Value: &validValue},
			field:   "To",
			tag:     "required",
		},
		{
			name:    "unsupported target currency",
			request: requests.CurrencyConversionRequest{From: "USD", To: "EUR", Value: &validValue},
			field:   "To",
			tag:     "oneof",
		},
		{
			name:    "missing value",
			request: requests.CurrencyConversionRequest{From: "USD", To: "BRL"},
			field:   "Value",
			tag:     "required",
		},
		{
			name:    "negative value",
			request: requests.CurrencyConversionRequest{From: "USD", To: "BRL", Value: &negative},
			field:   "Value",
			tag:     "gte",
		},
		{
			name:    "NaN value",
			request: requests.CurrencyConversionRequest{From: "USD", To: "BRL", Value: &nan},
			field:   "Value",
			tag:     "finite",
		},
		{
			name:    "positive infinity",
			request: requests.CurrencyConversionRequest{From: "USD", To: "BRL", Value: &infinity},
			field:   "Value",
			tag:     "finite",
		},
		{
			name:    "negative infinity",
			request: requests.CurrencyConversionRequest{From: "USD", To: "BRL", Value: &negativeInfinity},
			field:   "Value",
			tag:     "finite",
		},
		{
			name:    "conversion result overflows",
			request: requests.CurrencyConversionRequest{From: "USD", To: "BRL", Value: floatPointer(math.MaxFloat64)},
			field:   "Value",
			tag:     "finite_conversion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(ctx, tt.request)
			if tt.wantValid {
				assert.NoError(t, err)
				return
			}

			var validationErr *errors.FieldValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Condition(t, func() bool {
				for _, fieldErr := range validationErr.Errors {
					if fieldErr.Field == tt.field && fieldErr.Tag == tt.tag {
						return true
					}
				}
				return false
			}, "expected validation error for %s with tag %s, got %+v", tt.field, tt.tag, validationErr.Errors)
		})
	}
}

func floatPointer(value float64) *float64 {
	return &value
}

func TestPlaygroundValidator_Validate(t *testing.T) {
	v := NewPlaygroundValidator()
	ctx := context.Background()

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
