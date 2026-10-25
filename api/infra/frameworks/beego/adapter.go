package beego

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/beego/beego/v2/server/web"
	bctx "github.com/beego/beego/v2/server/web/context"
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

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller[any, any, any, any]) {
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
		if frameworks.HasBody(ctx.Request) {
			if err := frameworks.ValidateJSONContentType(ctx.Request.Header.Values("Content-Type")...); err != nil {
				resp := a.handle.ResolveError(rCtx, domainerrors.NewRequestError(err))
				a.writeResponse(rCtx, ctx, resp)
				return
			}
			if err := json.NewDecoder(ctx.Request.Body).Decode(&body); err != nil && err != io.EOF {
				resp := a.handle.ResolveError(rCtx, domainerrors.NewRequestError(err))
				a.writeResponse(rCtx, ctx, resp)
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

		cookies := frameworks.ParseCookies(ctx.Request.Header.Get("Cookie"))
		pathParams := domain.HttpParamsType{}
		for k, v := range ctx.Input.Params() {
			key := k
			if len(key) > 0 && key[0] == ':' {
				key = key[1:]
			}
			pathParams[key] = domain.HttpParamType{v}
		}

		rawReq := requests.NewRequestFromParams[any, any](body, headers, query, pathParams, cookies)
		req := requestWrapper{
			Request:     rawReq,
			headers:     headers,
			queryParams: query,
			pathParams:  pathParams,
			cookies:     cookies,
		}

		result, err := ctrl.Execute(rCtx, req)
		if err != nil {
			resp := a.handle.ResolveError(rCtx, err)
			a.writeResponse(rCtx, ctx, resp)
			return
		}

		a.writeResponse(rCtx, ctx, result)
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

func (a *adapter) writeResponse(ctx ports.Context, beegoCtx *bctx.Context, resp responses.Response[any, any]) {
	b, err := frameworks.MarshalResponse(resp)
	if err != nil {
		ctx.Logger().Error("failed to encode response", err, nil)
		http.Error(beegoCtx.ResponseWriter, "failed to encode response", http.StatusInternalServerError)
		return
	}
	frameworks.SetResponseHeaders(beegoCtx.ResponseWriter.Header(), resp)
	beegoCtx.ResponseWriter.Header().Set("Content-Type", "application/json")
	beegoCtx.ResponseWriter.WriteHeader(resp.StatusCode())
	if _, err := beegoCtx.ResponseWriter.Write(b); err != nil {
		ctx.Logger().Error("failed to write response", err, nil)
	}
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	a.server.Run(addr)
	return nil
}
