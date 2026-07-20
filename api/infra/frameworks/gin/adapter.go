package gin

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

type adapter struct {
	engine *gin.Engine
	handle ports.Handler
	tracer tracing.Tracer
}

func NewAdapter(handle ports.Handler) ports.FrameworkAdapter {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	return &adapter{
		engine: engine,
		handle: handle,
		tracer: tracing.NewTracer("gin-adapter"),
	}
}

func (a *adapter) RegisterRoute(method string, path string, ctrl ports.Controller) {
	a.engine.Handle(method, path, func(c *gin.Context) {
		ctx, span := a.tracer.Start(c.Request.Context(), "http.request")
		defer span.End()

		l := logger.FromContext(ctx)
		ctx = logger.ToContext(ctx, l)

		l.Info("incoming request", logger.LoggerFieldType{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"trace_id": a.tracer.GetTraceID(ctx),
		})

		var body any
		if c.Request.ContentLength > 0 {
			if err := c.ShouldBindJSON(&body); err != nil && err != io.EOF {
				resp := a.handle.ResolveError(ctx, err)
				c.JSON(http.StatusOK, resp)
				return
			}
		}

		headers := domain.NewHttpParam()
		headers.SetAll(c.Request.Header)

		query := domain.NewHttpParam()
		query.SetAll(c.Request.URL.Query())

		pathParams := domain.NewHttpParam()
		for _, param := range c.Params {
			pathParams.Set(param.Key, param.Value)
		}

		req := requests.NewRequestFromParams(body, headers, query, pathParams, &requests.HelloRequest{})
		result, err := ctrl.WrapperExecute(ctx, req)
		if err != nil {
			resp := a.handle.ResolveError(ctx, err)
			c.JSON(http.StatusOK, resp)
			return
		}

		resp := a.handle.Handle(ctx, http.StatusOK, result, nil)
		c.JSON(http.StatusOK, resp)
	})
}

func (a *adapter) Use(middleware ports.Middleware) {
	_ = middleware
}

func (a *adapter) Start(addr string) error {
	return a.engine.Run(addr)
}
