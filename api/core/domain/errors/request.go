package errors

type RequestError struct {
	err error
}

func NewRequestError(err error) *RequestError {
	return &RequestError{err: err}
}

func (e *RequestError) Error() string {
	return "invalid request"
}

func (e *RequestError) Unwrap() error {
	return e.err
}
