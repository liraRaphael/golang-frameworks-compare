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

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller) {
	a.app.Add(method, path, func(c *fiber.Ctx) error {
		ctx, span := a.tracer.Start(c.UserContext(), "http.request")
		defer span.End()

		l := logger.FromContext(ctx)
		ctx = logger.ToContext(ctx, l)

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

		headers := domain.NewHttpParam()
		headersMap := make(map[string][]string)
		c.Request().Header.VisitAll(func(key, value []byte) {
			k := string(key)
			headersMap[k] = append(headersMap[k], string(value))
		})
		headers.SetAll(headersMap)

		query := domain.NewHttpParam()
		queryMap := make(map[string][]string)
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			k := string(key)
			queryMap[k] = append(queryMap[k], string(value))
		})
		query.SetAll(queryMap)

		pathParams := domain.NewHttpParam()
		for k, v := range c.AllParams() {
			pathParams.Set(k, v)
		}

		req := requests.NewRequestFromParams(body, headers, query, pathParams, &requests.HelloRequest{})
		result, err := ctrl.WrapperExecute(ctx, req)
		if err != nil {
			resp := a.handle.ResolveError(ctx, err)
			return c.Status(http.StatusOK).JSON(resp)
		}

		resp := a.handle.Handle(ctx, http.StatusOK, result, nil)
		return c.Status(http.StatusOK).JSON(resp)
	})
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	return a.app.Listen(addr)
}
