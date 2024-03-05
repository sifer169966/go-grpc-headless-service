package config

import (
	"strings"

	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App       appConfig
	GRPClient gRPCClientConfig
}

type appConfig struct {
	RESTPort string `envconfig:"APP_REST_PORT" example:"8443"`
}

type gRPCClientConfig struct {
	ServerHost string `envconfig:"GRPC_CLIENT_SERVER_HOST"`
	ServerPort string `envconfig:"GRPC_CLIENT_SERVER_PORT"`
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
