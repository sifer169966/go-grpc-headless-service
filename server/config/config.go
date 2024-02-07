package config

import (
	"strings"

	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App appConfig
}

type appConfig struct {
	GRPCPort string `envconfig:"APP_GRPC_PORT" example:"8443"`
	GRPCHost string `envconfig:"APP_GRPC_HOST" example:"8443"`
}

var config Config

// Init is application config initialization ...
func Init() error {
	err := godotenv.Load()
	if err != nil {
		envFileNotFound := strings.Contains(err.Error(), "no such file or directory")
		if !envFileNotFound {
			log.Println("cound not read environment", "error", err)
		} else {
			log.Println("use OS environment")
		}
	}
	err = envconfig.Process("", &config)
	if err != nil {
		return err
	}
	return nil
}

func Get() *Config {
	return &config
}
