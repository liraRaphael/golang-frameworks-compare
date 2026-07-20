package config

import "github.com/kelseyhightower/envconfig"

type AppConfig struct {
	Port         string `envconfig:"PORT" default:"8080"`
	PprofPort    string `envconfig:"PPROF_PORT" default:"8081"`
	SwaggerPort  string `envconfig:"SWAGGER_PORT" default:"8082"`
	OtelEndpoint string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4318"`
	ServiceName  string `envconfig:"SERVICE_NAME" default:"go-framework-bench"`
	Framework    string `envconfig:"FRAMEWORK" default:"nethttp"`
}

func init() {
	_ = envconfig.Process("APP", &appConfig)
}

var appConfig = AppConfig{}

func NewAppConfig() AppConfig {
	return appConfig
}
