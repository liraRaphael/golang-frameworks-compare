package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientError(t *testing.T) {
	t.Run("without internal error", func(t *testing.T) {
		err := &ClientError{
			StatusCode: 400,
			Message:    "bad request",
		}
		assert.Equal(t, "client error: bad request (status: 400)", err.Error())
	})

	t.Run("with internal error", func(t *testing.T) {
		internalErr := errors.New("connection timeout")
		err := &ClientError{
			StatusCode: 504,
			Message:    "gateway timeout",
			Err:        internalErr,
		}
		assert.Equal(t, "client error: gateway timeout (status: 504, err: connection timeout)", err.Error())
	})
}

func TestFieldValidationError(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		err := &FieldValidationError{
			Errors: []FieldError{},
		}
		assert.Equal(t, "validation errors: ", err.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		err := &FieldValidationError{
			Errors: []FieldError{
				{Field: "Name", Tag: "required"},
				{Field: "Email", Tag: "email"},
			},
		}
		assert.Equal(t, "validation errors: field: Name, tag: required; field: Email, tag: email", err.Error())
	})
}
