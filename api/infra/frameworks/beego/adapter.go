package beego

import (
	"encoding/json"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

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
	handler := func(ctx *context.Context) {
		rCtx, span := a.tracer.Start(ctx.Request.Context(), "http.request")
		defer span.End()

		l := logger.FromContext(rCtx)
		rCtx = logger.ToContext(rCtx, l)

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

		headers := domain.NewHttpParam()
		headers.SetAll(ctx.Request.Header)

		query := domain.NewHttpParam()
		query.SetAll(ctx.Request.URL.Query())

		pathParams := domain.NewHttpParam()
		for k, v := range ctx.Input.Params() {
			key := k
			if len(key) > 0 && key[0] == ':' {
				key = key[1:]
			}
			pathParams.Set(key, v)
		}

		cookies := domain.NewHttpParam()

		req := requests.NewRequestFromParams(body, headers, query, pathParams, cookies)
		result, err := ctrl.WrapperExecute(rCtx, req)
		if err != nil {
			resp := a.handle.ResolveError(rCtx, err)
			a.writeResponse(ctx, resp)
			return
		}

		resp := a.handle.Handle(rCtx, http.StatusOK, result, nil)
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

func (a *adapter) writeResponse(ctx *context.Context, resp any) {
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
