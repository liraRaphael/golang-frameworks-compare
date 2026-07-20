package main

import (
	"context"
	"fmt"
	"time"

	"github.com/liraraphael/go-framework-bench/api/adapter/controller"
	"github.com/liraraphael/go-framework-bench/api/adapter/handler"
	"github.com/liraraphael/go-framework-bench/api/adapter/presenter"
	"github.com/liraraphael/go-framework-bench/api/adapter/router"
	"github.com/liraraphael/go-framework-bench/api/adapter/validation"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/core/usecases"
	"github.com/liraraphael/go-framework-bench/api/infra/config"
	"github.com/liraraphael/go-framework-bench/api/infra/diagnostic"
	"github.com/liraraphael/go-framework-bench/api/infra/docs/swagger"
	"github.com/liraraphael/go-framework-bench/api/infra/frameworks/beego"
	"github.com/liraraphael/go-framework-bench/api/infra/frameworks/chi"
	"github.com/liraraphael/go-framework-bench/api/infra/frameworks/echo"
	"github.com/liraraphael/go-framework-bench/api/infra/frameworks/fiber"
	"github.com/liraraphael/go-framework-bench/api/infra/frameworks/gin"
	"github.com/liraraphael/go-framework-bench/api/infra/frameworks/nethttp"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/metrics"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

func main() {
	cfg := config.NewAppConfig()

	// Initialize Observability
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize Logger
	l := logger.Initialize()
	l.Info("Starting application", logger.LoggerFieldType{
		"service": cfg.ServiceName,
		"port":    cfg.Port,
	})

	// 2. Initialize Tracing
	shutdownTracing, err := tracing.InitTracer(ctx, tracing.TracerConfig{
		ServiceName: cfg.ServiceName,
		Endpoint:    cfg.OtelEndpoint,
	})
	if err != nil {
		l.Error("failed to initialize tracer", err, nil)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdownTracing(ctx); err != nil {
				l.Error("failed to shutdown tracer", err, nil)
			}
		}()
	}

	// 3. Initialize Metrics
	shutdownMetrics, err := metrics.InitMetrics(ctx, metrics.MetricsConfig{
		ServiceName: cfg.ServiceName,
		Endpoint:    cfg.OtelEndpoint,
	})
	if err != nil {
		l.Error("failed to initialize metrics", err, nil)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdownMetrics(ctx); err != nil {
				l.Error("failed to shutdown metrics", err, nil)
			}
		}()
	}

	// 4. Initialize Swagger
	swag := swagger.NewRegistry(cfg.ServiceName, "Benchmarking API for Go Frameworks", "1.0.0")

	// 5. Start Diagnostic Servers
	diagnostic.StartPprof(fmt.Sprintf(":%s", cfg.PprofPort))

	hl := handler.NewStandardHandler()
	val := validation.NewPlaygroundValidator()

	helloPresenter := presenter.NewHelloPresenter()
	helloWorldUsecase := usecases.NewHelloUseCase(helloPresenter)
	helloWorldController := controller.NewHelloController(helloWorldUsecase, val)

	healthController := controller.NewHealthController()

	var fwAdapter ports.FrameworkAdapter
	switch cfg.Framework {
	case "gin":
		fwAdapter = gin.NewAdapter(hl)
	case "fiber":
		fwAdapter = fiber.NewAdapter(hl)
	case "echo":
		fwAdapter = echo.NewAdapter(hl)
	case "chi":
		fwAdapter = chi.NewAdapter(hl)
	case "beego":
		fwAdapter = beego.NewAdapter(hl)
	default:
		l.Info("framework not specified or unrecognized, defaulting to nethttp", logger.LoggerFieldType{"framework": cfg.Framework})
		fwAdapter = nethttp.NewAdapter(hl)
	}

	stdRouter := router.NewStandardRouter(fwAdapter)

	// Cast to concrete type to use WithSwagger
	if r, ok := stdRouter.(*router.StandardRouter); ok {
		r.WithSwagger(swag)
	}

	stdRouter.AddRoute("GET", "/hello", helloWorldController, swagger.NewRouteMetadata().
		WithSummary("Gera um cumprimento").
		WithDescription("Retorna uma mensagem de hello baseada no nome fornecido").
		WithTags("Greetings").
		WithInput(new(requests.HelloRequest)).
		WithOutput(new(responses.HelloOutput)))

	stdRouter.AddRoute("GET", "/health", healthController, swagger.NewRouteMetadata().
		WithSummary("Verifica a saúde da aplicação").
		WithDescription("Retorna o status atual da aplicação").
		WithTags("Health").
		WithOutput(new(map[string]string{"status": "ok"})))

	swag.StartSwaggerUI(fmt.Sprintf(":%s", cfg.SwaggerPort))

	l.Info("Server listening", logger.LoggerFieldType{"port": cfg.Port})
	if err := stdRouter.Start(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		l.Fatal("Server failed to start", logger.LoggerFieldType{"error": err})
	}

}
