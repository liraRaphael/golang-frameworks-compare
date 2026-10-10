package echo

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/liraraphael/go-framework-bench/api/core/domain"
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
	echo   *echo.Echo
	handle ports.Handler
	tracer tracing.Tracer
}

func NewAdapter(handle ports.Handler) ports.FrameworkAdapter {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	return &adapter{
		echo:   e,
		handle: handle,
		tracer: tracing.NewTracer("echo-adapter"),
	}
}

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller[any, any, any, any]) {
	a.echo.Add(method, path, func(c echo.Context) error {
		baseCtx, span := a.tracer.Start(c.Request().Context(), "http.request")
		defer span.End()

		ctx := context.NewContext(baseCtx)
		l := ctx.Logger()
		ctx = context.NewContext(logger.ToContext(ctx, l))

		l.Info("incoming request", logger.LoggerFieldType{
			"method":   c.Request().Method,
			"path":     c.Request().URL.Path,
			"trace_id": a.tracer.GetTraceID(ctx),
		})

		var body any
		if c.Request().ContentLength > 0 {
			if err := c.Bind(&body); err != nil && err != io.EOF {
				resp := a.handle.ResolveError(ctx, err)
				return a.writeResponse(ctx, c, resp)
			}
		}

		headers := domain.HttpParamsType{}
		for k, v := range c.Request().Header {
			headers[k] = domain.HttpParamType(v)
		}

		query := domain.HttpParamsType{}
		for k, v := range c.QueryParams() {
			query[k] = domain.HttpParamType(v)
		}

		cookies := frameworks.ParseCookies(c.Request().Header.Get("Cookie"))
		pathParams := domain.HttpParamsType{}
		for _, name := range c.ParamNames() {
			pathParams[name] = domain.HttpParamType{c.Param(name)}
		}

		rawReq := requests.NewRequestFromParams[any, any](body, headers, query, pathParams, cookies)
		req := requestWrapper{
			Request:     rawReq,
			headers:     headers,
			queryParams: query,
			pathParams:  pathParams,
			cookies:     cookies,
		}

		result, err := ctrl.Execute(ctx, req)
		if err != nil {
			resp := a.handle.ResolveError(ctx, err)
			return a.writeResponse(ctx, c, resp)
		}

		return a.writeResponse(ctx, c, result)
	})
}

func (a *adapter) writeResponse(ctx ports.Context, c echo.Context, resp responses.Response[any, any]) error {
	b, err := frameworks.MarshalResponse(resp)
	if err != nil {
		ctx.Logger().Error("failed to encode response", err, nil)
		return c.NoContent(http.StatusInternalServerError)
	}
	frameworks.SetResponseHeaders(c.Response().Header(), resp)
	return c.Blob(resp.StatusCode(), "application/json", b)
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	return a.echo.Start(addr)
}
