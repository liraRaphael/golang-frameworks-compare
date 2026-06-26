package errors

import (
	"fmt"
	"strings"
)

type FieldError struct {
	Field     string `json:"field"`
	Tag       string `json:"tag"`
	Value     any    `json:"value,omitempty"`
	Param     string `json:"param,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type FieldValidationError struct {
	Errors []FieldError `json:"errors"`
}

func (e *FieldValidationError) Error() string {
	var errs []string
	for _, err := range e.Errors {
		errs = append(errs, fmt.Sprintf("field: %s, tag: %s", err.Field, err.Tag))
	}
	return fmt.Sprintf("validation errors: %s", strings.Join(errs, "; "))
}
