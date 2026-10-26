package validation

import (
	"math"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
)

var cepValidatorPattern = regexp.MustCompile(`^[0-9]{8}$`)

func registerValidationRules(validate *validator.Validate) error {
	if err := validate.RegisterValidation("finite", func(field validator.FieldLevel) bool {
		value := field.Field()
		if value.Kind() != reflect.Float32 && value.Kind() != reflect.Float64 {
			return false
		}
		return !math.IsNaN(value.Float()) && !math.IsInf(value.Float(), 0)
	}); err != nil {
		return err
	}

	if err := validate.RegisterValidation("cep", func(field validator.FieldLevel) bool {
		if field.Field().Kind() != reflect.String {
			return false
		}
		value := field.Field().String()
		if value == "" {
			return false
		}
		return cepValidatorPattern.MatchString(value)
	}); err != nil {
		return err
	}

	return validate.RegisterValidation("finite_conversion", func(field validator.FieldLevel) bool {
		request, ok := field.Parent().Interface().(requests.CurrencyConversionRequest)
		if !ok || request.Value == nil {
			return false
		}

		result := request.CurrencyQuote()
		return !math.IsNaN(result) && !math.IsInf(result, 0)
	})
}
