package errors

import "fmt"

type ClientError struct {
	StatusCode int
	Body       []byte
	Message    string
	Err        error
}

func (e *ClientError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("client error: %s (status: %d, err: %v)", e.Message, e.StatusCode, e.Err)
	}
	return fmt.Sprintf("client error: %s (status: %d)", e.Message, e.StatusCode)
}
