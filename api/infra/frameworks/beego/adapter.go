package beego

import (
	"encoding/json"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	bctx "github.com/beego/beego/v2/server/web/context"
	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/context"
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
	server *web.HttpServer
	handle ports.Handler
	tracer tracing.Tracer
}

func NewAdapter(handle ports.Handler) ports.FrameworkAdapter {
	server := web.NewHttpSever()
	return &adapter{
		server: server,
		handle: handle,
		tracer: tracing.NewTracer("beego-adapter"),
	}
}

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller[any, any]) {
	handler := func(ctx *bctx.Context) {
		baseCtx, span := a.tracer.Start(ctx.Request.Context(), "http.request")
		defer span.End()

		rCtx := context.NewContext(baseCtx)
		l := rCtx.Logger()
		rCtx = context.NewContext(logger.ToContext(rCtx, l))

		l.Info("incoming request", logger.LoggerFieldType{
			"method":   ctx.Request.Method,
			"path":     ctx.Request.URL.Path,
			"trace_id": a.tracer.GetTraceID(rCtx),
		})

		var body any
		if ctx.Request.ContentLength > 0 {
			if err := json.NewDecoder(ctx.Request.Body).Decode(&body); err != nil {
				resp := a.handle.ResolveError(rCtx, err)
				a.writeResponse(ctx, resp)
				return
			}
		}

		headers := domain.HttpParamsType{}
		for k, v := range ctx.Request.Header {
			headers[k] = domain.HttpParamType(v)
		}

		query := domain.HttpParamsType{}
		for k, v := range ctx.Request.URL.Query() {
			query[k] = domain.HttpParamType(v)
		}

		pathParams := domain.HttpParamsType{}
		for k, v := range ctx.Input.Params() {
			key := k
			if len(key) > 0 && key[0] == ':' {
				key = key[1:]
			}
			pathParams[key] = domain.HttpParamType{v}
		}

		rawReq := requests.NewRequestFromParams[any, any](body, headers, query, pathParams, nil)
		req := requestWrapper{
			Request:     rawReq,
			headers:     headers,
			queryParams: query,
			pathParams:  pathParams,
		}

		result, err := ctrl.WrapperExecute(rCtx, req)
		if err != nil {
			resp := a.handle.ResolveError(rCtx, err)
			a.writeResponse(ctx, resp)
			return
		}

		resp := a.handle.Handle(rCtx, http.StatusOK, result, nil, nil)
		a.writeResponse(ctx, resp)
	}

	switch method {
	case http.MethodGet:
		a.server.Get(path, handler)
	case http.MethodPost:
		a.server.Post(path, handler)
	case http.MethodPut:
		a.server.Put(path, handler)
	case http.MethodDelete:
		a.server.Delete(path, handler)
	default:
		a.server.Any(path, handler)
	}
}

func (a *adapter) writeResponse(ctx *bctx.Context, resp any) {
	ctx.Output.Header("Content-Type", "application/json")
	b, _ := json.Marshal(resp)
	_ = ctx.Output.Body(b)
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	a.server.Run(addr)
	return nil
}
