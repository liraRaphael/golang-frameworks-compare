package fiber

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
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
	app    *fiber.App
	handle ports.Handler
	tracer tracing.Tracer
}

func NewAdapter(handle ports.Handler) ports.FrameworkAdapter {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	return &adapter{
		app:    app,
		handle: handle,
		tracer: tracing.NewTracer("fiber-adapter"),
	}
}

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller[any, any]) {
	a.app.Add(method, path, func(c *fiber.Ctx) error {
		baseCtx, span := a.tracer.Start(c.UserContext(), "http.request")
		defer span.End()

		ctx := ports.NewContext(baseCtx)
		l := ctx.Logger()
		ctx = ports.NewContext(logger.ToContext(ctx, l))

		l.Info("incoming request", logger.LoggerFieldType{
			"method":   c.Method(),
			"path":     c.Path(),
			"trace_id": a.tracer.GetTraceID(ctx),
		})

		var body any
		if len(c.Body()) > 0 {
			if err := json.Unmarshal(c.Body(), &body); err != nil {
				resp := a.handle.ResolveError(ctx, err)
				return c.Status(http.StatusOK).JSON(resp)
			}
		}

		headers := domain.HttpParamsType{}
		c.Request().Header.VisitAll(func(key, value []byte) {
			k := string(key)
			headers[k] = append(headers[k], string(value))
		})

		query := domain.HttpParamsType{}
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			k := string(key)
			query[k] = append(query[k], string(value))
		})

		pathParams := domain.HttpParamsType{}
		for k, v := range c.AllParams() {
			pathParams[k] = domain.HttpParamType{v}
		}

		rawReq := requests.NewRequestFromParams[any, any](body, headers, query, pathParams, nil)
		req := requestWrapper{
			Request:     rawReq,
			headers:     headers,
			queryParams: query,
			pathParams:  pathParams,
		}

		result, err := ctrl.WrapperExecute(ctx, req)
		if err != nil {
			resp := a.handle.ResolveError(ctx, err)
			return c.Status(http.StatusOK).JSON(resp)
		}

		resp := a.handle.Handle(ctx, http.StatusOK, result, nil, nil)
		return c.Status(http.StatusOK).JSON(resp)
	})
}

func (a *adapter) Use(middleware ports.Middleware) {
	a.app.Use(func(c *fiber.Ctx) error {
		return nil
	})
}

func (a *adapter) Start(addr string) error {
	return a.app.Listen(addr)
}
