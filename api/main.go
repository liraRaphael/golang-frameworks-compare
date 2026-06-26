package main

import (
	"fmt"
	"log"

	"github.com/liraraphael/go-framework-bench/api/adapter/controller"
	"github.com/liraraphael/go-framework-bench/api/adapter/handler"
	"github.com/liraraphael/go-framework-bench/api/adapter/presenter"
	"github.com/liraraphael/go-framework-bench/api/adapter/router"
	"github.com/liraraphael/go-framework-bench/api/core/usecases"
	"github.com/liraraphael/go-framework-bench/api/infra/config"
	"github.com/liraraphael/go-framework-bench/api/infra/frameworks/nethttp"
)

func main() {
	cfg := config.NewAppConfig()
	hl := handler.NewStandardHandler()

	helloPresenter := presenter.NewHelloPresenter()
	helloWorldUsecase := usecases.NewHelloUseCase(helloPresenter)
	helloWorldController := controller.NewHelloController(helloWorldUsecase)

	healthController := controller.NewHealthController()

	nethttp := nethttp.NewAdapter(hl)
	stdRouter := router.NewStandardRouter(nethttp)

	stdRouter.AddRoute("GET", "/hello", helloWorldController)

	stdRouter.AddRoute("GET", "/health", healthController)

	log.Printf("server listening on :%s", cfg.Port)
	if err := stdRouter.Start(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatal(err)
	}
}
