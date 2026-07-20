package nethttp

import (
	"encoding/json"
	"net/http"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

type requestWrapper struct {
	requests.Request[any, any]
	headers     domain.HttpParamsType
	queryParams domain.HttpParamsType
	pathParams  domain.HttpParamsType
	cookies     domain.HttpParamsType
}

func (w requestWrapper) Headers() domain.HttpParamsType     { return w.headers }
func (w requestWrapper) QueryParams() domain.HttpParamsType { return w.queryParams }
func (w requestWrapper) PathParams() domain.HttpParamsType  { return w.pathParams }
func (w requestWrapper) Cookies() domain.HttpParamsType     { return w.cookies }

type adapter struct {
	mux    *http.ServeMux
	handle ports.Handler
	tracer tracing.Tracer
}

func NewAdapter(handle ports.Handler) ports.FrameworkAdapter {
	return &adapter{
		mux:    http.NewServeMux(),
		handle: handle,
		tracer: tracing.NewTracer("nethttp-adapter"),
	}
}

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller[any, any]) {
	a.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ctx, span := a.tracer.Start(r.Context(), "http.request")
		defer span.End()

		l := logger.FromContext(ctx)
		ctx = logger.ToContext(ctx, l)

		l.Info("incoming request", logger.LoggerFieldType{
			"method":   r.Method,
			"path":     r.URL.Path,
			"trace_id": a.tracer.GetTraceID(ctx),
		})

		var body any
		if r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				resp := a.handle.ResolveError(ctx, err)
				a.writeResponse(w, resp)
				return
			}
		}

		headers := domain.HttpParamsType{}
		for k, v := range r.Header {
			headers[k] = domain.HttpParamType(v)
		}

		query := domain.HttpParamsType{}
		for k, v := range r.URL.Query() {
			query[k] = domain.HttpParamType(v)
		}

		rawReq := requests.NewRequestFromParams[any, any](body, headers, query, nil, nil)
		req := requestWrapper{
			Request:     rawReq,
			headers:     headers,
			queryParams: query,
		}

		result, err := ctrl.WrapperExecute(ctx, req)
		if err != nil {
			resp := a.handle.ResolveError(ctx, err)
			a.writeResponse(w, resp)
			return
		}

		resp := a.handle.Handle(ctx, http.StatusOK, result, nil, nil)
		a.writeResponse(w, resp)
	})
}

func (a *adapter) writeResponse(w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(resp)
	w.Write(b)
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	return http.ListenAndServe(addr, a.mux)
}
