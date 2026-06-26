package nethttp

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type adapter struct {
	mux    *http.ServeMux
	handle ports.Handler
}

func NewAdapter(handle ports.Handler) ports.FrameworkAdapter {
	return &adapter{mux: http.NewServeMux(), handle: handle}
}

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller) {
	a.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := []byte{}
		_, err := r.Body.Read(body)
		if err != nil {
			a.handle.ResolveError(ctx, err)
		}

		var input any

		err = json.Unmarshal(body, input)
		if err != nil {
			a.handle.ResolveError(ctx, err)
		}

		headers := domain.NewHttpParam()
		headers.SetAll(r.Header)

		result, err := ctrl.WrapperExecute(ctx, requests.NewRequest(input))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"erro": map[string]any{"codigo": "internal_error", "mensagem": "Erro interno do servidor não mapeado"}})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
	})
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	log.Println("net/http adapter listening on", addr)
	return http.ListenAndServe(addr, a.mux)
}

func (a *adapter) HandleFunc(path string, h func(w http.ResponseWriter, r *http.Request)) {
	a.mux.HandleFunc(path, h)
}

func (a *adapter) ContextWithValue(ctx context.Context, key, value any) context.Context {
	return context.WithValue(ctx, key, value)
}
