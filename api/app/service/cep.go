package service

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	domainerrors "github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

var cepPattern = regexp.MustCompile(`^[0-9]{8}$`)

type cepService struct {
	client ports.Client
}

func NewCEPService(client ports.Client) ports.CepService {
	return &cepService{client: client}
}

func (s *cepService) Find(ctx ports.Context, cep string) (responses.CepAddress, error) {
	if !cepPattern.MatchString(cep) {
		return responses.CepAddress{}, &domainerrors.CEPInvalidError{CEP: cep}
	}

	endpoint := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	res, err := s.client.Do(ctx, enums.MethodGet, endpoint, requests.NewRequest[any, any](nil, nil))
	if err != nil {
		return responses.CepAddress{}, &domainerrors.CEPUpstreamError{StatusCode: http.StatusBadGateway, Message: "falha ao consultar ViaCEP", Err: err}
	}

	if res.StatusCode() != http.StatusOK {
		return responses.CepAddress{}, &domainerrors.CEPUpstreamError{StatusCode: res.StatusCode(), Message: "ViaCEP respondeu com status inesperado"}
	}

	resBody, ok := res.Body().(responses.CepAddress)
	if !ok {
		return resBody, &domainerrors.CEPUpstreamError{StatusCode: http.StatusBadGateway, Message: "Falha ao converter tipagem na resposta do serviço em ViaCEP"}
	}

	return resBody, nil
}
