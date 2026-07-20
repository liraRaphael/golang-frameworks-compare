package presenter

import (
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/stretchr/testify/assert"
)

func TestHelloPresenter_Output(t *testing.T) {
	p := NewHelloPresenter()
	res, err := p.Output(requests.HelloRequest{Name: "Presenter"})

	assert.NoError(t, err)
	assert.Equal(t, "Hello, Presenter", res.Message)
}
