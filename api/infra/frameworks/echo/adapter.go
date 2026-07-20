package echo

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

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

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller) {
	a.echo.Add(method, path, func(c echo.Context) error {
		ctx, span := a.tracer.Start(c.Request().Context(), "http.request")
		defer span.End()

		l := logger.FromContext(ctx)
		ctx = logger.ToContext(ctx, l)

		l.Info("incoming request", logger.LoggerFieldType{
			"method":   c.Request().Method,
			"path":     c.Request().URL.Path,
			"trace_id": a.tracer.GetTraceID(ctx),
		})

		var body any
		if c.Request().ContentLength > 0 {
			if err := c.Bind(&body); err != nil && err != io.EOF {
				resp := a.handle.ResolveError(ctx, err)
				return c.JSON(http.StatusOK, resp)
			}
		}

		headers := domain.NewHttpParam()
		headers.SetAll(c.Request().Header)

		query := domain.NewHttpParam()
		query.SetAll(c.QueryParams())

		pathParams := domain.NewHttpParam()
		for _, name := range c.ParamNames() {
			pathParams.Set(name, c.Param(name))
		}

		req := requests.NewRequestFromParams(body, headers, query, pathParams, &requests.HelloRequest{})
		result, err := ctrl.WrapperExecute(ctx, req)
		if err != nil {
			resp := a.handle.ResolveError(ctx, err)
			return c.JSON(http.StatusOK, resp)
		}

		resp := a.handle.Handle(ctx, http.StatusOK, result, nil)
		return c.JSON(http.StatusOK, resp)
	})
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	return a.echo.Start(addr)
}
