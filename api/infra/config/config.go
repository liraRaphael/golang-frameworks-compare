package config

import "github.com/kelseyhightower/envconfig"

type AppConfig struct {
	Port string `envconfig:"PORT" default:"8080"`
}

func init() {
	_ = envconfig.Process("APP", &appConfig)
}

var appConfig = AppConfig{}

func NewAppConfig() AppConfig {
	return appConfig
}
