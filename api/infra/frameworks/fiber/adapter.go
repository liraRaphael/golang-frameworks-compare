package fiber

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
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

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller[any, any, any, any]) {
	a.app.Add(method, path, func(c *fiber.Ctx) error {
		baseCtx, span := a.tracer.Start(c.UserContext(), "http.request")
		defer span.End()

		ctx := context.NewContext(baseCtx)
		l := ctx.Logger()
		ctx = context.NewContext(logger.ToContext(ctx, l))

		l.Info("incoming request", logger.LoggerFieldType{
			"method":   c.Method(),
			"path":     c.Path(),
			"trace_id": a.tracer.GetTraceID(ctx),
		})

		var body any
		if len(c.Body()) > 0 {
			if err := json.Unmarshal(c.Body(), &body); err != nil {
				resp := a.handle.ResolveError(ctx, err)
				return a.writeResponse(ctx, c, resp)
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

		cookies := domain.HttpParamsType{}
		c.Request().Header.VisitAllCookie(func(key, value []byte) {
			name := string(key)
			cookies[name] = append(cookies[name], string(value))
		})
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

func (a *adapter) writeResponse(ctx ports.Context, c *fiber.Ctx, resp responses.Response[any, any]) error {
	b, err := frameworks.MarshalResponse(resp)
	if err != nil {
		ctx.Logger().Error("failed to encode response", err, nil)
		return c.SendStatus(http.StatusInternalServerError)
	}
	for name, values := range resp.Headers() {
		if len(values) == 0 {
			continue
		}
		c.Set(name, values[0])
		for _, value := range values[1:] {
			c.Append(name, value)
		}
	}
	for name, values := range resp.Cookies() {
		for _, value := range values {
			c.Append(fiber.HeaderSetCookie, (&http.Cookie{Name: name, Value: value}).String())
		}
	}
	c.Status(resp.StatusCode())
	c.Set(fiber.HeaderContentType, "application/json")
	return c.Send(b)
}

func (a *adapter) Use(middleware ports.Middleware) {
	a.app.Use(func(c *fiber.Ctx) error {
		return nil
	})
}

func (a *adapter) Start(addr string) error {
	return a.app.Listen(addr)
}
