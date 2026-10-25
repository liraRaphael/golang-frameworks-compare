package nethttp

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	domainerrors "github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/context"
	"github.com/liraraphael/go-framework-bench/api/infra/frameworks"
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

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller[any, any, any, any]) {
	a.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		baseCtx, span := a.tracer.Start(r.Context(), "http.request")
		defer span.End()

		ctx := context.NewContext(baseCtx)
		l := ctx.Logger()
		ctx = context.NewContext(logger.ToContext(ctx, l))

		l.Info("incoming request", logger.LoggerFieldType{
			"method":   r.Method,
			"path":     r.URL.Path,
			"trace_id": a.tracer.GetTraceID(ctx),
		})

		var body any
		if frameworks.HasBody(r) {
			if err := frameworks.ValidateJSONContentType(r.Header.Values("Content-Type")...); err != nil {
				resp := a.handle.ResolveError(ctx, domainerrors.NewRequestError(err))
				a.writeResponse(ctx, w, resp)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
				resp := a.handle.ResolveError(ctx, domainerrors.NewRequestError(err))
				a.writeResponse(ctx, w, resp)
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

		cookies := frameworks.ParseCookies(r.Header.Get("Cookie"))
		rawReq := requests.NewRequestFromParams[any, any](body, headers, query, nil, cookies)
		req := requestWrapper{
			Request:     rawReq,
			headers:     headers,
			queryParams: query,
			cookies:     cookies,
		}

		result, err := ctrl.Execute(ctx, req)
		if err != nil {
			resp := a.handle.ResolveError(ctx, err)
			a.writeResponse(ctx, w, resp)
			return
		}

		a.writeResponse(ctx, w, result)
	})
}

func (a *adapter) writeResponse(ctx ports.Context, w http.ResponseWriter, resp responses.Response[any, any]) {
	b, err := frameworks.MarshalResponse(resp)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		ctx.Logger().Error("failed to encode response", err, nil)
		return
	}
	frameworks.SetResponseHeaders(w.Header(), resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode())
	if _, err := w.Write(b); err != nil {
		ctx.Logger().Error("failed to write response", err, nil)
	}
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	return http.ListenAndServe(addr, a.mux)
}
