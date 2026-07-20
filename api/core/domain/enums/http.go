package enums

type HttpMethod string

const (
	MethodGet     HttpMethod = "GET"
	MethodPost    HttpMethod = "POST"
	MethodPut     HttpMethod = "PUT"
	MethodDelete  HttpMethod = "DELETE"
	MethodPatch   HttpMethod = "PATCH"
	MethodOptions HttpMethod = "OPTIONS"
	MethodHead    HttpMethod = "HEAD"
)
