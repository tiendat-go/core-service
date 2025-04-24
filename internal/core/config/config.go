package config

import (
	"fmt"

	"github.com/spf13/viper"
	"google.golang.org/grpc/keepalive"
)

type HttpServerConfig struct {
	Port uint
}

type GrpcServerConfig struct {
	Port            uint32
	KeepaliveParams keepalive.ServerParameters
	KeepalivePolicy keepalive.EnforcementPolicy
}

type Config struct {
	ServiceName string `mapstructure:"SERVICE_NAME"`
	ServicePort int    `mapstructure:"SERVICE_PORT"`
	RegistryURL string `mapstructure:"REGISTRY_URL"`
}

func InitConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error reading config: %w", err))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %w", err))
	}

	return &cfg
}
