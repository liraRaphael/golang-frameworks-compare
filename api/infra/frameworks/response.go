package frameworks

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"strings"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

func MarshalResponse(resp responses.Response[any, any]) ([]byte, error) {
	return json.Marshal(resp)
}

func HasBody(r *http.Request) bool {
	return r.Body != nil && r.Body != http.NoBody
}

func ValidateJSONContentType(values ...string) error {
	if len(values) != 1 {
		return errors.New("content type must be application/json")
	}
	mediaType, _, err := mime.ParseMediaType(values[0])
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return errors.New("content type must be application/json")
	}
	return nil
}

func ParseCookies(cookieHeader string) domain.HttpParamsType {
	request := &http.Request{
		Header: http.Header{"Cookie": []string{cookieHeader}},
	}
	cookies := domain.HttpParamsType{}
	for _, cookie := range request.Cookies() {
		cookies[cookie.Name] = append(cookies[cookie.Name], cookie.Value)
	}
	return cookies
}

func SetResponseHeaders(header http.Header, resp responses.Response[any, any]) {
	setParams(header, resp.Headers())
	for name, values := range resp.Cookies() {
		for _, value := range values {
			header.Add("Set-Cookie", (&http.Cookie{Name: name, Value: value}).String())
		}
	}
}

func setParams(header http.Header, params domain.HttpParamsType) {
	for name, values := range params {
		if len(values) == 0 {
			header.Del(name)
			continue
		}
		header.Set(name, values[0])
		for _, value := range values[1:] {
			header.Add(name, value)
		}
	}
}
